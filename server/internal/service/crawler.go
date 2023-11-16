package service

import (
	"compress/gzip"
	"compress/lzw"
	"compress/zlib"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/dsnet/compress/brotli"
	"io"
	"log"
	"net/http"
	"net/url"
	"server/internal/model"
	"strconv"
	"strings"
	"sync"
	"time"
)

type CrawlerService interface {
	Crawl(ctx context.Context, reqId, url string, broadcast chan []byte)
}

type crawlService struct {
	visitedURLs   map[string]bool
	sitemap       *model.Sitemap
	visitedMu     sync.Mutex
	sitemapPageMu sync.Mutex
}

// NewCrawlerService builds a service
func NewCrawlerService() CrawlerService {
	return &crawlService{
		visitedURLs: make(map[string]bool),
		sitemap: &model.Sitemap{
			Pages: make(map[string][]string),
		},
	}
}

// Crawl visits the url and all the links within the same domain and returns the sitemap for the website
func (s *crawlService) Crawl(ctx context.Context, reqId, url string, broadcast chan []byte) {
	const (
		depth    = 1
		maxDepth = 3
	)

	subdomain := s.getSubdomain(url)
	delay := s.getDelay(subdomain)

	s.crawl(ctx, url, reqId, subdomain, delay, depth, maxDepth, broadcast)

	fmt.Println("done crawling")
}

// getSubdomain parses the url and gets only the subdomain
func (s *crawlService) getSubdomain(urlStr string) string {
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		log.Printf("error parsing url: %s", err)
		return ""
	}

	return parsedURL.Scheme + "://" + parsedURL.Hostname()
}

// getDelay parses the robots file for the url to check how often it can be requested/crawled
func (s *crawlService) getDelay(url string) int {
	resp, err := http.Get(fmt.Sprintf("%s/%s", url, "robots.txt"))
	if err != nil {
		log.Printf("error getting robots file: %s", err)
		return 0
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("error reading robots file: %s", err)
		return 0
	}

	lines := strings.Split(string(body), "\n")
	for _, line := range lines {
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 || !strings.EqualFold(strings.TrimSpace(parts[0]), "Crawl-delay") {
			continue
		}

		delay, err := strconv.Atoi(strings.TrimSpace(parts[1]))
		if err != nil {
			log.Printf("error converting crawl-delay value: %s", err)
			return 0
		}

		return delay
	}

	return 0
}

// crawl recursively goes through the website and builds its sitemap based on the links found in the same subdomain
func (s *crawlService) crawl(ctx context.Context, urlStr, reqId, subdomain string, delay, depth, maxDepth int, broadcast chan []byte) {
	if !strings.Contains(urlStr, subdomain) || depth > maxDepth {
		return
	}

	res, doc, err := s.visit(urlStr, delay)
	if err != nil {
		log.Printf("error visiting the page %s: %s", urlStr, err)
		return
	}
	defer res.Body.Close()

	links, err := s.getLinks(doc, res, subdomain, urlStr)
	if err != nil {
		log.Printf("error getting links from the page %s: %s", urlStr, err)
		return
	}
	s.sitemapPageMu.Lock()
	s.sitemap.Pages[urlStr] = links

	// Send the crawled urls to the websocket connection
	resp := &model.Response{
		Request: model.Request{
			ReqId: reqId,
			Url:   urlStr,
		},
		Sitemap: model.Sitemap{
			Pages: s.sitemap.Pages,
		},
		Status: "ok",
	}

	body, err := json.Marshal(resp)
	if err != nil {
		log.Printf("error marshaling response: %s", err)
		return
	}

	log.Printf("links found in %s: %v", urlStr, links)

	s.sitemapPageMu.Unlock()

	broadcast <- body

	// Process found links in page in parallel
	var linkWg sync.WaitGroup
	for _, link := range links {
		if strings.Contains(link, subdomain) {
			linkWg.Add(1)
			go func(link string) {
				defer linkWg.Done()
				if !s.visitedLink(link) {
					go s.crawl(ctx, link, reqId, subdomain, delay, depth+1, maxDepth, broadcast)
				}
			}(link)
		}
	}

	linkWg.Wait()
}

// visit requests a page and returns its document representation
func (s *crawlService) visit(urlStr string, delay int) (*http.Response, *goquery.Document, error) {
	// Wait for crawl delay from robots.txt
	time.Sleep(time.Duration(delay) * time.Second)

	res, err := http.Get(urlStr)
	if err != nil {
		return nil, nil, errors.New(fmt.Sprintln("error getting url:", err))
	}

	body, err := s.decompress(res)
	if err != nil {
		return nil, nil, errors.New(fmt.Sprintln("error decompressing document:", err))
	}

	doc, err := goquery.NewDocumentFromReader(body)
	if err != nil {
		return nil, nil, errors.New(fmt.Sprintln("error parsing document:", err))
	}

	s.markVisited(urlStr)

	return res, doc, nil
}

// visitedLink checks if the link has been already crawled
func (s *crawlService) visitedLink(link string) bool {
	s.visitedMu.Lock()
	defer s.visitedMu.Unlock()
	return s.visitedURLs[link]
}

// markVisited mark the link as visited to avoid duplicates
func (s *crawlService) markVisited(link string) {
	s.visitedMu.Lock()
	s.visitedURLs[link] = true
	s.visitedMu.Unlock()
}

// decompress takes the "Content-Encoding" header and applies the decompression using each algorithm found
func (s *crawlService) decompress(res *http.Response) (io.ReadCloser, error) {
	encoding := res.Header["Content-Encoding"]
	body := res.Body
	var encErr error

	for _, enc := range encoding {
		switch enc {
		case "br":
			body, encErr = brotli.NewReader(body, nil)
		case "gzip":
			body, encErr = gzip.NewReader(body)
		case "compress":
			body = lzw.NewReader(body, lzw.LSB, 8)
		case "deflate":
			body, encErr = zlib.NewReader(body)
		}
	}

	if encErr != nil {
		return nil, errors.New(fmt.Sprintf("error decompressing: %s", encErr))
	}

	return body, nil
}

// getLinks returns the links found in a document
func (s *crawlService) getLinks(doc *goquery.Document, res *http.Response, subdomain, urlStr string) (links []string, err error) {
	doc.Find("a").Each(func(index int, element *goquery.Selection) {
		href, exists := element.Attr("href")
		if !exists {
			return
		}

		linkURL, err := url.Parse(strings.ReplaceAll(href, " ", ""))
		if err != nil {
			log.Printf("error resolving relative url to absolute url: %s", err)
			return
		}

		absoluteURL := res.Request.URL.ResolveReference(linkURL)
		link := absoluteURL.String()

		// Ensure the link belongs to the same subdomain
		if !strings.Contains(link, subdomain) || link == urlStr {
			return
		}

		links = s.appendSet(links, link)

	})

	return
}

// appendSet appends to the slice only if the link is not already added
func (s *crawlService) appendSet(links []string, link string) []string {
	exists := false
	for _, l := range links {
		if l == link {
			exists = true
			break
		}
	}

	if !exists {
		links = append(links, link)
	}

	return links
}
