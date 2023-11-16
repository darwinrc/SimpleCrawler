package model

type Request struct {
	ReqId string `json:"reqId,omitempty"`
	Url   string `json:"url,omitempty"`
}

type Response struct {
	Request
	Sitemap
	Status string `json:"status"`
}

type Sitemap struct {
	Pages map[string][]string `json:"pages"`
}
