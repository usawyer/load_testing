package test_client

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/cookiejar"
	"net/url"

	"github.com/usawyer/load_testing/internal/html"
	"golang.org/x/time/rate"
)

type Client struct {
	client    *http.Client
	limiter   *rate.Limiter
	serialNum int
}

func New(limiter *rate.Limiter, serialNum int) (*Client, error) {
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
		limiter:   limiter,
		serialNum: serialNum,
	}, nil
}

func (c *Client) InitTest(startPage string) error {
	slog.Info(fmt.Sprintf("init test for client №%d", c.serialNum))

	startUrl, err := url.Parse(startPage)
	if err != nil {
		return fmt.Errorf("error to parse start page for client №%d: %v", c.serialNum, err)
	}

	getRes, err := c.makeGetRequest(startUrl.String())
	if err != nil {
		return fmt.Errorf("error to make GET request for client №%d: %v", c.serialNum, err)
	}
	defer getRes.Body.Close()

	redirectUrl, err := getRes.Location()
	if err != nil {
		return fmt.Errorf("error to get redirect url for client №%d: %v", c.serialNum, err)
	}
	slog.Info(fmt.Sprintf("redirect url for client №%d: %s", c.serialNum, redirectUrl.String()))

	for redirectUrl.Path != "/passed" {
		getRes, err = c.makeGetRequest(redirectUrl.String())
		if err != nil {
			return fmt.Errorf("error to make GET request for client №%d: %v", c.serialNum, err)
		}

		body, err := html.ParsePageAndFormBody(getRes.Body)
		if err != nil {
			return fmt.Errorf("error to make request body for client №%d: %v", c.serialNum, err)
		}

		postRes, err := c.makePostRequest(redirectUrl.String(), body)
		if err != nil {
			return fmt.Errorf("error to make POST request for client №%d: %v", c.serialNum, err)
		}

		if postRes.StatusCode != http.StatusFound {
			return fmt.Errorf("no redirect url for client №%d: %v", c.serialNum, postRes.Status)
		} else {
			redirectUrl, err = postRes.Location()
			if err != nil {
				return fmt.Errorf("error to get redirect url for client №%d: %v", c.serialNum, err)
			}
			slog.Info(fmt.Sprintf("redirect url for client №%d: %s", c.serialNum, redirectUrl.String()))
		}
	}

	return nil
}

func (c *Client) makeGetRequest(url string) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	err = c.limiter.Wait(context.Background())
	if err != nil {
		return nil, err
	}

	return c.client.Do(req)
}

func (c *Client) makePostRequest(url string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodPost, url, body)
	if err != nil {
		return nil, err
	}

	err = c.limiter.Wait(context.Background())
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	return c.client.Do(req)
}
