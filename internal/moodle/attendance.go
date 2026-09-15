package moodle

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"time"

	"golang.org/x/net/html"
)

var (
	ErrForeignLink = errors.New("link does not point to Moodle")
	ErrLoginFailed = errors.New("Moodle login failed")
)

type Client struct {
	BaseURL  string
	Username string
	Password string
}

type Result struct {
	OK      bool
	Message string
}

type page struct {
	url *url.URL
	doc *html.Node
}

func (c Client) IsMoodleLink(link string) bool {
	u, err := url.Parse(link)
	return err == nil && c.sameOrigin(u)
}

func (c Client) MarkAttendance(link string) (Result, error) {
	if !c.IsMoodleLink(link) {
		return Result{}, ErrForeignLink
	}

	jar, err := cookiejar.New(nil)
	if err != nil {
		return Result{}, err
	}
	client := &http.Client{Jar: jar, Timeout: 30 * time.Second}

	p, err := fetch(client, http.MethodGet, link, nil)
	if err != nil {
		return Result{}, err
	}

	if form := findLoginForm(p.doc); form != nil {
		p, err = c.login(client, p, form)
		if err != nil {
			return Result{}, err
		}
	}

	return readResult(p), nil
}

func (c Client) login(client *http.Client, p *page, form *html.Node) (*page, error) {
	action, err := p.url.Parse(attr(form, "action"))
	if err != nil {
		return nil, err
	}
	if !c.sameOrigin(action) {
		return nil, fmt.Errorf("login form posts to %s: %w", action.Host, ErrForeignLink)
	}

	token := find(form, func(n *html.Node) bool {
		return n.Data == "input" && attr(n, "name") == "logintoken"
	})
	if token == nil {
		return nil, errors.New("login form has no logintoken")
	}

	values := url.Values{
		"anchor":     {""},
		"logintoken": {attr(token, "value")},
		"username":   {c.Username},
		"password":   {c.Password},
	}

	next, err := fetch(client, http.MethodPost, action.String(), values)
	if err != nil {
		return nil, err
	}

	if findLoginForm(next.doc) != nil {
		msg := find(next.doc, func(n *html.Node) bool {
			return attr(n, "id") == "loginerrormessage"
		})
		if msg != nil && text(msg) != "" {
			return nil, fmt.Errorf("%w: %s", ErrLoginFailed, text(msg))
		}
		return nil, ErrLoginFailed
	}

	return next, nil
}

func (c Client) sameOrigin(u *url.URL) bool {
	base, err := url.Parse(c.BaseURL)
	return err == nil && u.Scheme == base.Scheme && u.Host == base.Host
}

func fetch(client *http.Client, method, link string, form url.Values) (*page, error) {
	var body io.Reader
	if form != nil {
		body = strings.NewReader(form.Encode())
	}

	req, err := http.NewRequest(method, link, body)
	if err != nil {
		return nil, err
	}
	if form != nil {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%s %s: %s", method, resp.Request.URL.Path, resp.Status)
	}

	doc, err := html.Parse(io.LimitReader(resp.Body, 5<<20))
	if err != nil {
		return nil, err
	}

	return &page{url: resp.Request.URL, doc: doc}, nil
}

func findLoginForm(doc *html.Node) *html.Node {
	return find(doc, func(n *html.Node) bool {
		return n.Data == "form" && find(n, func(in *html.Node) bool {
			return in.Data == "input" && attr(in, "name") == "logintoken"
		}) != nil
	})
}

func readResult(p *page) Result {
	for _, alert := range findAll(p.doc, isAlert) {
		msg := text(alert)
		if msg == "" {
			continue
		}
		class := attr(alert, "class")
		return Result{
			OK:      strings.Contains(class, "alert-success"),
			Message: msg,
		}
	}

	title := find(p.doc, func(n *html.Node) bool { return n.Data == "title" })
	msg := "Moodle did not show a confirmation"
	if title != nil {
		msg = fmt.Sprintf("%s (ended up on %q)", msg, text(title))
	}
	return Result{Message: msg}
}

func isAlert(n *html.Node) bool {
	if n.Type != html.ElementNode {
		return false
	}
	for _, class := range strings.Fields(attr(n, "class")) {
		if class == "alert" {
			return true
		}
	}
	return false
}

func attr(n *html.Node, key string) string {
	for _, a := range n.Attr {
		if a.Key == key {
			return a.Val
		}
	}
	return ""
}

func find(n *html.Node, match func(*html.Node) bool) *html.Node {
	if all := findAll(n, match); len(all) > 0 {
		return all[0]
	}
	return nil
}

func findAll(root *html.Node, match func(*html.Node) bool) []*html.Node {
	var found []*html.Node
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode && match(n) {
			found = append(found, n)
		}
		for child := range n.ChildNodes() {
			walk(child)
		}
	}
	walk(root)
	return found
}

func text(n *html.Node) string {
	var sb strings.Builder
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode && (n.Data == "button" || n.Data == "script" || n.Data == "style") {
			return
		}
		if n.Type == html.TextNode {
			sb.WriteString(n.Data)
			sb.WriteByte(' ')
		}
		for child := range n.ChildNodes() {
			walk(child)
		}
	}
	walk(n)
	return strings.Join(strings.Fields(sb.String()), " ")
}
