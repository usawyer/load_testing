package html

import (
	"io"
	"net/url"
	"strings"

	"golang.org/x/net/html"
)

func ParsePageAndFormBody(body io.Reader) (io.Reader, error) {
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
