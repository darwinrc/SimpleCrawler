package service

import (
	"context"
	"github.com/PuerkitoBio/goquery"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"testing"
)

func TestNewCrawlerService(t *testing.T) {
	service := NewCrawlerService()

	assert.NotNil(t, service)
	assert.NotNil(t, service.(*crawlService).visitedURLs)
	assert.NotNil(t, service.(*crawlService).sitemap)
	assert.NotNil(t, service.(*crawlService).visitedMu)
	assert.NotNil(t, service.(*crawlService).sitemapPageMu)
}

func TestGetSubdomain(t *testing.T) {
	service := NewCrawlerService()

	url := "http://parserdigital.com/path"
	subdomain := service.(*crawlService).getSubdomain(url)

	assert.Equal(t, "http://parserdigital.com", subdomain)
}

func TestGetDelay(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	url := "http://parserdigital.com"
	resp := `Crawl-delay: 5`
	httpmock.RegisterResponder("GET", url+"/robots.txt", httpmock.NewStringResponder(200, resp))

	service := NewCrawlerService()
	delay := service.(*crawlService).getDelay(url)

	assert.Equal(t, 5, delay)
}

func TestVisitedLink(t *testing.T) {
	service := NewCrawlerService()

	service.(*crawlService).markVisited("http://parserdigital.com")

	visited := service.(*crawlService).visitedLink("http://parserdigital.com")
	assert.True(t, visited)

	visited = service.(*crawlService).visitedLink("http://other.com")
	assert.False(t, visited)
}

func TestMarkVisited(t *testing.T) {
	service := NewCrawlerService()

	service.(*crawlService).markVisited("http://parserdigital.com")

	visited := service.(*crawlService).visitedURLs["http://parserdigital.com"]
	assert.True(t, visited)
}

func TestDecompress(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	url := "http://parserdigital.com"
	respBody := "compressed data"
	httpmock.RegisterResponder("GET", url, httpmock.NewStringResponder(200, respBody))

	service := NewCrawlerService()
	res, err := http.Get(url)
	assert.NoError(t, err)

	decompressed, err := service.(*crawlService).decompress(res)
	assert.NoError(t, err)

	data, err := io.ReadAll(decompressed)
	assert.NoError(t, err)
	assert.Equal(t, "compressed data", string(data))
}

func TestGetLinks(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	url := "http://parserdigital.com"
	respBody := `<html><body><a href="http://parserdigital.com/link1"></a></body></html>`
	httpmock.RegisterResponder("GET", url, httpmock.NewStringResponder(200, respBody))

	res, err := http.Get(url)
	assert.NoError(t, err)
	defer res.Body.Close()

	doc, err := goquery.NewDocumentFromReader(res.Body)
	assert.NoError(t, err)

	service := NewCrawlerService()
	links, err := service.(*crawlService).getLinks(doc, res, "http://parserdigital.com", url)

	assert.NoError(t, err)
	assert.Equal(t, []string{"http://parserdigital.com/link1"}, links)
}

func TestCrawl(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	url := "http://parserdigital.com"
	respBody := `<html><body><a href="http://parserdigital.com/link1"></a></body></html>`
	httpmock.RegisterResponder("GET", url, httpmock.NewStringResponder(200, respBody))

	broadcast := make(chan []byte, 1)

	service := NewCrawlerService()

	ctx := context.Background()
	reqID := "123"

	go service.Crawl(ctx, reqID, url, broadcast)

	msg := <-broadcast

	expectedMsg := `{"reqId":"123","url":"` + url + `","pages":{"` + url + `":["http://parserdigital.com/link1"]},"status":"ok"}`
	assert.Equal(t, expectedMsg, string(msg))
}
