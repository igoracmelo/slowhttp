package main

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

var requestLinePattern = regexp.MustCompile(`(\w+) (\S+) (\S+)`) // METHOD /PATH HTTP/X.X
var headerPattern = regexp.MustCompile(`(\S+): (.*)`)            // My-Header: valid, cool, 1.0
var httpVerbs = []string{"GET", "POST", "PUT"}

type RequestLine struct {
	Method  string
	Path    string
	Version string
}

type Request struct {
	RequestLine
	Headers    map[string]string
	BodyBytes  []byte
	BodyString string
}

func parseRequestString(requestString string) (*Request, error) {
	lines := strings.Split(requestString, "\n")
	reqLine, err := parseRequestLine(lines[0])
	if err != nil {
		return nil, err
	}

	headers := map[string]string{}
	var bodyLine int

	for i, line := range lines[1:] {
		if len(strings.TrimSpace(line)) == 0 {
			bodyLine = i + 2
			break
		}

		name, value, err := parseHeaderLine(line)
		if err != nil {
			return nil, err
		}

		headers[name] = value
	}

	var bodyStr string

	if len(lines) > bodyLine {
		bodyStr = strings.Join(lines[bodyLine:], "\n")
	}

	request := &Request{
		RequestLine: *reqLine,
		Headers:     headers,
		BodyBytes:   []byte(bodyStr),
		BodyString:  bodyStr,
	}

	return request, nil
}

func readSample(t *testing.T, sample string) string {
	b, err := os.ReadFile(sample)
	assert.NoError(t, err)
	return string(b)
}

func TestParseRequest(t *testing.T) {
	t.Run("should parse request", func(t *testing.T) {
		reqString := readSample(t, "sample-01.http")

		expected := &Request{
			RequestLine: RequestLine{
				Method:  "POST",
				Path:    "/",
				Version: "HTTP/1.1",
			},
			Headers: map[string]string{
				"Host":            "localhost:12345",
				"Connection":      "keep-alive",
				"Sec-GPC":         "1",
				"Accept-Language": "pt-BR,pt",
				"Accept-Encoding": "gzip, deflate, br",
			},
			BodyString: "hello world",
			BodyBytes:  []byte("hello world"),
		}

		req, err := parseRequestString(reqString)
		assert.NoError(t, err)
		assert.EqualValues(t, expected, req)
	})

	t.Run("should parse request", func(t *testing.T) {
		reqString := readSample(t, "sample-02.http")

		expected := &Request{
			RequestLine: RequestLine{
				Method:  "POST",
				Path:    "/users",
				Version: "HTTP/1.1",
			},
			Headers: map[string]string{
				"Host": "localhost:12345",
			},
			BodyString: `{ "id": "123", "name": "hello world" }`,
			BodyBytes:  []byte(`{ "id": "123", "name": "hello world" }`),
		}

		req, err := parseRequestString(reqString)
		assert.NoError(t, err)
		assert.EqualValues(t, expected, req)
	})
}

func parseHeaderLine(line string) (string, string, error) {
	m := headerPattern.FindStringSubmatch(line)

	if len(m) != 3 {
		return "", "", fmt.Errorf("invalid header line: %s", line)
	}

	name := m[1]
	value := m[2]

	return name, value, nil
}

func TestParseHeaderLine(t *testing.T) {
	t.Run("should return error for empty line", func(t *testing.T) {
		_, _, err := parseHeaderLine("")
		assert.Error(t, err)
	})

	t.Run("should work for host: google.com", func(t *testing.T) {
		name, value, err := parseHeaderLine("host: google.com")

		assert.NoError(t, err)
		assert.Equal(t, "google.com", value)
		assert.Equal(t, "host", name)
	})
}

func Contains[T comparable](slice []T, value T) bool {
	for _, v := range slice {
		if v == value {
			return true
		}
	}
	return false
}

func parseRequestLine(line string) (*RequestLine, error) {
	m := requestLinePattern.FindStringSubmatch(line)

	if len(m) == 0 {
		return nil, fmt.Errorf("invalid request line: %s", line)
	}

	if !Contains(httpVerbs, m[1]) {
		return nil, fmt.Errorf("invalid http verb: %s", m[1])
	}

	return &RequestLine{
		Method:  m[1],
		Path:    m[2],
		Version: m[3],
	}, nil
}

func TestParseRequestLine(t *testing.T) {
	{
		rl, err := parseRequestLine("GET / HTTP/1.0")

		assert.NoError(t, err)
		assert.Equal(t, "GET", rl.Method)
		assert.Equal(t, "/", rl.Path)
		assert.Equal(t, "HTTP/1.0", rl.Version)
	}

	{
		rl, err := parseRequestLine("GET HTTP/1.0")

		assert.Error(t, err)
		assert.Nil(t, rl)
	}

	{
		rl, err := parseRequestLine("FAKE / HTTP/1.0")

		assert.Error(t, err)
		assert.Nil(t, rl)
	}
}

// func TestA(t *testing.T) {
// 	// x := firstLinePattern.FindAllStringSubmatch("GET / HTTP/1.0", -1)

// 	assert.True(t, requestLinePattern.MatchString("GET / HTTP/1.0"))

// 	m := requestLinePattern.FindStringSubmatch("GET / HTTP/1.0")

// 	method := m[1]
// 	path := m[2]
// 	version := m[3]

// }
