package slowhttp

import (
	"fmt"
	"io"
	"net/url"
)

type Method uint8

const (
	MethodGet Method = iota + 1
	MethodHead
	MethodPost
	MethodPut
	MethodDelete
	MethodConnect
	MethodOptions
	MethodTrace
	MethodPatch
)

var MethodByNumber = []string{
	MethodGet:     "GET",
	MethodHead:    "HEAD",
	MethodPost:    "POST",
	MethodPut:     "PUT",
	MethodDelete:  "DELETE",
	MethodConnect: "CONNECT",
	MethodOptions: "OPTIONS",
	MethodTrace:   "TRACE",
	MethodPatch:   "PATCH",
}

type Status uint16

// TODO
const (
	StatusOK       Status = 200
	StatusNotFound Status = 404
)

// TODO
var Statuses = map[string]Status{
	"OK": StatusOK,
}

// TODO: canonicalize
type Headers map[string][]string

func (h Headers) Add(k, v string) {
	h[k] = append(h[k], v)
}

func (h Headers) Set(k, v string) {
	h[k] = []string{v}
}

func (h Headers) Get(k string) string {
	vals := h[k]
	if vals == nil {
		return ""
	}
	if len(vals) == 0 {
		return ""
	}
	return vals[0]
}

func (h Headers) GetAll(k string) []string {
	return h[k]
}

type Request struct {
	Method  Method
	URL     *url.URL
	Version Version
	Headers Headers
	Body    io.Reader
}

func NewRequest(method Method, rawURL string, headers Headers, body io.Reader) (*Request, error) {
	u, err := url.ParseRequestURI(rawURL)
	if err != nil {
		return nil, err
	}

	if headers == nil {
		headers = Headers{}
	}

	headers.Set("User-Agent", "slowhttp")
	headers.Set("Host", u.Host)

	r := Request{
		Method:  method,
		URL:     u,
		Version: Version{1, 1},
		Headers: headers,
		Body:    body,
	}

	return &r, nil
}

func (r *Request) Head() []byte {
	s := ""

	sMethod := MethodByNumber[r.Method]
	sTarget := r.URL.Path
	if r.URL.RawQuery != "" {
		sTarget += "?" + r.URL.RawQuery
	}
	s += fmt.Sprintf("%s %s HTTP/%s\n", sMethod, sTarget, r.Version.String())

	for k, vals := range r.Headers {
		for _, v := range vals {
			s += fmt.Sprintf("%s: %s\n", k, v)
		}
	}

	return []byte(s)
}

type Response struct {
	Version Version
	Body    io.ReadCloser
}

type Version struct {
	Major int
	Minor int
}

func (v Version) String() string {
	return fmt.Sprintf("%d.%d", v.Major, v.Minor)
}
