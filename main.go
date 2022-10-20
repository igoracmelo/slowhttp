package main

// don't clean your code too early!

import (
	"fmt"
	"net"
	"regexp"
)

func main() {
	listener, err := net.Listen("tcp", "localhost:12345")
	if err != nil {
		panic(err)
	}

	conn, err := listener.Accept()
	if err != nil {
		panic(err)
	}

	// ----- handle connection
	for {
		read := make([]byte, 1024)
		n, err := conn.Read(read)
		if err != nil {
			panic(err)
		}

		// readStr = string(read[:n])
		firstLinePattern := regexp.MustCompile(`(\w+) (/\S+) (\S+)`) // METHOD /PATH HTTP/1.0
		x := firstLinePattern.FindAllSubmatch(read, -1)
		fmt.Println(x)
		fmt.Println(n)
	}
}
