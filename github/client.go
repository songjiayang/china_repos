package github

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
)

const (
	defaultUserAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_13_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/79.0.3945.130 Safari/537.36"
	host             = "https://github.com"
	searchURL        = "https://github.com/search"
)

type Client struct {
	cookie string
	*http.Client
}

func NewClient() *Client {
	return &Client{
		Client: &http.Client{
			Transport: http.DefaultTransport,
		},
	}
}

func (c *Client) LoadCookie(cookieFile string) {
	if cookieFile == "" {
		return
	}

	data, err := ioutil.ReadFile(cookieFile)
	if err != nil {
		log.Panicf("failed to load cookies:<file=%s, error=%s>", cookieFile, err)
	}

	c.cookie = string(data)
}

func (c *Client) NewRequest(method, url string, payload interface{}) (*http.Request, error) {
	var body io.Reader

	if payload != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}

		body = bytes.NewReader(data)
	}

	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, err
	}

	if c.cookie != "" {
		req.Header.Set("Cookie", c.cookie)
	}

	req.Header.Set("User-Agent", defaultUserAgent)
	req.Header.Set("Host", "github.com")

	return req, nil
}

func (c *Client) Do(req *http.Request) (io.ReadCloser, error) {
	res, err := c.Client.Do(req)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != 200 {
		res.Body.Close()

		return nil, fmt.Errorf("do request with error: %s %d %s", req.URL, res.StatusCode, res.Status)
	}

	return res.Body, nil
}
