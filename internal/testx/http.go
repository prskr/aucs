package testx

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

type ResponseRule interface {
	Matches(req *http.Request) bool
	Apply(resp *http.Response)
}

func MockHTTPClient(rules ...ResponseRule) *http.Client {
	return &http.Client{
		Transport: mockRoundTripper(rules),
	}
}

var _ http.RoundTripper = mockRoundTripper(nil)

type mockRoundTripper []ResponseRule

// RoundTrip implements http.RoundTripper.
func (m mockRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	for _, r := range m {
		if r.Matches(req) {
			resp := new(http.Response)
			r.Apply(resp)
			return resp, nil
		}
	}

	return nil, fmt.Errorf("no matching response rule for URL: %s", req.URL.String())
}

var _ ResponseRule = (*SimpleUrlRule)(nil)

func NewSimpleUrlRule(rawUrl string, response []byte) (SimpleUrlRule, error) {
	parsedUrl, err := url.Parse(rawUrl)
	if err != nil {
		return SimpleUrlRule{}, err
	}

	return SimpleUrlRule{
		StatusCode: http.StatusOK,
		URL:        parsedUrl,
		Response:   response,
	}, nil
}

type SimpleUrlRule struct {
	StatusCode int
	URL        *url.URL
	Response   []byte
}

// Apply implements ResponseRule.
func (s SimpleUrlRule) Apply(resp *http.Response) {
	resp.StatusCode = s.StatusCode
	resp.Body = io.NopCloser(bytes.NewReader(s.Response))
}

// Matches implements ResponseRule.
func (s SimpleUrlRule) Matches(req *http.Request) bool {
	return s.URL.Scheme == req.URL.Scheme &&
		s.URL.Host == req.URL.Host &&
		s.URL.Path == req.URL.Path &&
		s.URL.Query().Encode() == req.URL.Query().Encode()
}
