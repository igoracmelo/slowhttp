package slowhttp

import (
	"io"
	"strings"
	"testing"
)

func Test_Request_Head(t *testing.T) {
	headers := Headers{}
	headers.Set("Authorization", "Bearer blablabla")
	headers.Add("Multiple", "first")
	headers.Add("Multiple", "second")

	r, err := NewRequest(MethodConnect, "http://google.com/search?q=%20", headers, nil)
	if err != nil {
		t.Fatal(err)
	}

	head := string(r.Head())

	tests := []string{
		"CONNECT /search?q=%20 HTTP/1.1\n",
		"\nHost: google.com\n",
		"\nAuthorization: Bearer blablabla\n",
		"\nUser-Agent: slowhttp\n",
		"\nMultiple: first\n",
		"\nMultiple: second\n",
	}

	t.Log(head)
	for _, test := range tests {
		if !strings.Contains(head, test) {
			t.Errorf("must contain '%s'", test)
		}
	}
}

func Test_ReadResponse(t *testing.T) {
	r := struct {
		io.Reader
		io.Closer
	}{
		Reader: strings.NewReader("HTTP/2.5 OK\n\nbody!"),
	}

	resp, err := ReadResponse(r)
	if err != nil {
		t.Fatal(err)
	}

	if resp.Version.Major != 2 {
		t.Errorf("Version.Major - want: 2, got: %v", resp.Version.Major)
	}
	if resp.Version.Minor != 5 {
		t.Errorf("Version.Minor - want: 5, got: %v", resp.Version.Minor)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	if string(body) != "body!" {
		t.Errorf("body - want: 'body', got: %s", string(body))
	}
}
