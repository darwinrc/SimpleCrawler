package mock_service

import (
	"fmt"
	"github.com/jarcoal/httpmock"
)

type HttpMock struct {
	Url string
}

type Responder struct {
	Url  string
	Body string
}

func NewHttpMock(url string) *HttpMock {
	return &HttpMock{
		Url: url,
	}
}

func (m *HttpMock) RegisterResponders() {
	responders := []Responder{
		{
			Url: m.Url + "robots.txt",
			Body: `User-agent: *
Crawl-delay: 2`,
		},
		{
			Url: m.Url,
			Body: fmt.Sprintf(`<!DOCTYPE html>
<html lang= lang="en-GB">
	<body>
		<a href="%show-we-work">learn more</a>
		<a href="%scareer">Join us</a></p>
		<a href="%scontact-us">Contact us</a>
	</body>
</html>`, m.Url, m.Url, m.Url),
		},
		{
			Url: m.Url + "how-we-work",
			Body: fmt.Sprintf(`<!DOCTYPE html>
<html lang= lang="en-GB">
	<body>
		<a href="%scases">Cases</a>
		<a href="%speople">People</a></p>
		<a href="%s">Home</a>
	</body>
</html>`, m.Url, m.Url, m.Url),
		},
		{
			Url: m.Url + "career",
			Body: fmt.Sprintf(`<!DOCTYPE html>
<html lang= lang="en-GB">
	<body>
		<a href="%sapply">Apply</a>
		<a href="%svisit">Visit us</a></p>
		<a href="%s">Home</a>
	</body>
</html>`, m.Url, m.Url, m.Url),
		},
		{
			Url: m.Url + "contact-us",
			Body: fmt.Sprintf(`<!DOCTYPE html>
<html lang= lang="en-GB">
	<body>
		<a href="%saddress">Address</a>
		<a href="%sform">Form</a></p>
		<a href="%s">Home</a>
	</body>
</html>`, m.Url, m.Url, m.Url),
		},
		{
			Url:  m.Url + "people",
			Body: "Dummy text",
		},
		{
			Url:  m.Url + "visit",
			Body: "Dummy text",
		},
		{
			Url:  m.Url + "apply",
			Body: "Dummy text",
		},
		{
			Url:  m.Url + "cases",
			Body: "Dummy text",
		},
		{
			Url:  m.Url + "address",
			Body: "Dummy text",
		},
		{
			Url:  m.Url + "form",
			Body: "Dummy text",
		},
	}

	httpmock.Activate()
	for _, r := range responders {
		httpmock.RegisterResponder("GET", r.Url, httpmock.NewStringResponder(200, r.Body))
	}
}

func (m *HttpMock) DeactivateAndReset() {
	httpmock.DeactivateAndReset()
}
