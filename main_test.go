package slowhttp

import (
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
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
		Reader: strings.NewReader("HTTP/2.5 OK\nMessage: be cool : )\n\nbody!"),
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

	if resp.Headers.Get("Message") != "be cool : )" {
		t.Errorf("want header Message:\nbe cool : ), got:\n%s", resp.Headers.Get("Message"))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	if string(body) != "body!" {
		t.Errorf("body - want: 'body', got: %s", string(body))
	}
}

func Test_Client_Do(t *testing.T) {
	c := Client{}

	want := "this is \nthe body!!\n\n"
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(want))
	})
	go func() {
		err := http.ListenAndServe(":1234", nil)
		panic(err)
	}()

	for {
		_, err := http.Get("http://localhost:1234")
		if err == nil {
			break
		}
		t.Log(err)
		time.Sleep(time.Second)
	}

	req, err := NewRequest(MethodPost, "http://localhost:1234", nil, strings.NewReader("somebody"))
	if err != nil {
		t.Fatal(err)
	}

	resp, err := c.Do(req)
	if err != nil {
		t.Fatal(err)
	}

	if resp.Status != StatusOK {
		t.Errorf("server responded with status %s (%d)", resp.Status.String(), resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	got := string(body)
	if got != want {
		t.Errorf("body - want: %s, got: %s", want, got)
	}

	err = resp.Body.Close()
	if err != nil {
		t.Fatal(err)
	}
}
