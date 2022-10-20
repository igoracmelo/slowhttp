package main

import (
	"fmt"
	"regexp"
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

func parseHeaderLine(line string) (string, string, error) {
	if !headerPattern.MatchString(line) {
		return "", "", fmt.Errorf("invalid header line: %s", line)
	}

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

func Contains[T comparable](slice []T, value T) bool {
	for _, v := range slice {
		if v == value {
			return true
		}
	}
	return false
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
