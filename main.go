package slowhttp

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"net/url"
	"regexp"
	"strconv"
	"strings"
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

var NameByMethod = []string{
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

func (s Status) String() string {
	if int(s) >= len(NameByStatus) {
		return fmt.Sprintf("UNKNOWN (%d)", s)
	}
	return NameByStatus[s]
}

// TODO
const (
	StatusOK         Status = 200
	StatusBadRequest Status = 400
)

// TODO
var StatusByName = map[string]Status{
	"200 OK":          StatusOK,
	"400 Bad Request": StatusBadRequest,
}

// TODO
var NameByStatus = []string{
	StatusOK: "200 OK",
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

	sMethod := NameByMethod[r.Method]
	sTarget := r.URL.Path
	if sTarget == "" {
		sTarget = "/"
	}
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
	Status  Status
	Headers Headers
	Body    io.ReadCloser
}

func ReadResponse(r io.ReadCloser) (*Response, error) {
	buf := bufio.NewReader(r)

	line, err := buf.ReadString('\n')
	if err != nil {
		return nil, err
	}
	line = strings.TrimSpace(line)

	re := regexp.MustCompile(`HTTP/(\d+)\.(\d+) (.*)`)
	matches := re.FindStringSubmatch(line)
	if len(matches) != 4 {
		return nil, fmt.Errorf("couldn't recognize first line '%s' (%d matches)", line, len(matches))
	}

	sMaj := matches[1]
	sMin := matches[2]
	sStat := matches[3]

	maj, err := strconv.Atoi(sMaj)
	if err != nil {
		return nil, fmt.Errorf("invalid major number '%s'", sMaj)
	}
	min, err := strconv.Atoi(sMin)
	if err != nil {
		return nil, fmt.Errorf("invalid minor number '%s'", sMin)
	}

	version := Version{maj, min}
	stat, ok := StatusByName[sStat]
	if !ok {
		return nil, fmt.Errorf("unknown http status '%s'", sStat)
	}

	re = regexp.MustCompile(`(.*?):\s*(.*)`)
	headers := Headers{}
	for {
		line, err := buf.ReadString('\n')
		line = strings.TrimSpace(line)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		if line == "" {
			break
		}

		matches := re.FindStringSubmatch(line)
		if len(matches) != 3 {
			return nil, fmt.Errorf("invalid header line: '%s'", line)
		}

		headers.Add(matches[1], matches[2])
	}

	return &Response{
		Version: version,
		Status:  stat,
		Headers: headers,
		Body: struct {
			io.Reader
			io.Closer
		}{
			Reader: buf,
			Closer: r,
		},
	}, nil
}

type Version struct {
	Major int
	Minor int
}

func (v Version) String() string {
	return fmt.Sprintf("%d.%d", v.Major, v.Minor)
}

type Client struct{}

func (c *Client) Do(req *Request) (*Response, error) {
	conn, err := net.Dial("tcp", req.URL.Host)
	if err != nil {
		return nil, err
	}

	head := req.Head()
	_, err = conn.Write(head)
	if err != nil {
		return nil, err
	}

	if req.Body != nil {
		_, err := io.Copy(conn, req.Body)
		if err != nil {
			return nil, err
		}
	}

	return ReadResponse(conn)
}
