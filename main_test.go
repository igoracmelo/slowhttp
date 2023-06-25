package slowhttp

import (
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
