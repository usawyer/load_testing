package test

import (
	"fmt"
	"golang.org/x/net/html"
	"io"
	"log/slog"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"time"
)

type Client struct {
	client    *http.Client
	rate      <-chan time.Time
	serialNum int
}

func New(maxRPC int, serialNum int) (*Client, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, fmt.Errorf("error init cookie jar for client №%d: %v", serialNum, err)
	}

	return &Client{
		client: &http.Client{
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
			Jar: jar,
		},
		rate:      time.Tick(time.Millisecond * 1200 / time.Duration(maxRPC)),
		serialNum: serialNum,
	}, nil
}

func (c *Client) InitTest(startPage string) error {
	slog.Info(fmt.Sprintf("init test for client №%d", c.serialNum))

	url, err := url.Parse(startPage)
	if err != nil {
		return fmt.Errorf("error to parse start page for client №%d: %v", c.serialNum, err)
	}

	res, err := c.makeGetRequest(url.String())
	if err != nil {
		return fmt.Errorf("error to make GET request for client №%d: %v", c.serialNum, err)
	}
	defer res.Body.Close()

	redirectUrl, err := res.Location()
	if err != nil {
		return fmt.Errorf("error to get redirect url for client №%d: %v", c.serialNum, err)
	}

	slog.Info(fmt.Sprintf("redirect url for client №%d: %s", c.serialNum, redirectUrl.String()))

	for redirectUrl.Path != "/passed" {
		res, err := c.makeGetRequest(redirectUrl.String())
		if err != nil {
			return fmt.Errorf("error to make GET request for client №%d: %v", c.serialNum, err)
		}

		body, err := c.makeBody(res.Body)
		if err != nil {
			return fmt.Errorf("error to make request body for client №%d: %v", c.serialNum, err)
		}

		res, err = c.makePostRequest(redirectUrl.String(), body)
		if err != nil {
			return fmt.Errorf("error to make POST request for client №%d: %v", c.serialNum, err)
		}

		if res.StatusCode == http.StatusFound {
			redirectUrl, err = res.Location()
			if err != nil {
				return fmt.Errorf("error to get redirect url for client №%d: %v", c.serialNum, err)
			}

			slog.Info(fmt.Sprintf("redirect url for client №%d: %s", c.serialNum, redirectUrl.String()))
		} else {
			return fmt.Errorf("no redirect url for client №%d: %v", c.serialNum, res.Status)
		}
	}

	return nil
}

func (c *Client) makeGetRequest(url string) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	<-c.rate
	return c.client.Do(req)
}

func (c *Client) makePostRequest(url string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodPost, url, body)
	if err != nil {
		return nil, err
	}

	<-c.rate
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	return c.client.Do(req)
}

func (c *Client) makeBody(body io.Reader) (io.Reader, error) {
	doc, err := html.Parse(body)
	if err != nil {
		return nil, err
	}

	values := make(url.Values)

	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode {
			switch n.Data {
			case "select":
				name, value := extractSelectInfo(n)
				values.Set(name, value)
			case "input":
				name, value := extractInputInfo(n)
				if name != "" {
					values.Set(name, value)
				}
			}
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(doc)

	return strings.NewReader(values.Encode()), nil
}

func extractSelectInfo(n *html.Node) (string, string) {
	var name, longestValue string

	for _, a := range n.Attr {
		if a.Key == "name" {
			name = a.Val
			break
		}
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.ElementNode && c.Data == "option" {
			value := getAttrValue(c, "value")
			if len(value) > len(longestValue) {
				longestValue = value
			}
		}
	}

	return name, longestValue
}

func extractInputInfo(n *html.Node) (string, string) {
	var name, inputType string

	for _, a := range n.Attr {
		if a.Key == "name" {
			name = a.Val
		}
		if a.Key == "type" {
			inputType = a.Val
		}
	}

	if inputType == "text" {
		return name, "test"
	} else if inputType == "radio" {
		return name, findLongestValue(n)
	}

	return "", ""
}

func findLongestValue(n *html.Node) string {
	var longest string

	for c := n.Parent.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.ElementNode && c.Data == "input" && getAttrValue(c, "type") == "radio" {
			value := getAttrValue(c, "value")
			if len(value) > len(longest) {
				longest = value
			}
		}
	}

	return longest
}

func getAttrValue(n *html.Node, key string) string {
	for _, a := range n.Attr {
		if a.Key == key {
			return a.Val
		}
	}
	return ""
}
