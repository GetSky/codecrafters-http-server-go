package main

import (
	"fmt"
	"net"
	"os"
)

func main() {

	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}

	defer func(l net.Listener) {
		_ = l.Close()
	}(l)
	fmt.Println("Server is listening.")

	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer func(conn net.Conn) {
		_ = conn.Close()
	}(conn)

	_, err := conn.Write([]byte(generateStatusLine()))
	if err != nil {
		fmt.Println("Error writing status to connection: ", err.Error())
		return
	}

	_, err = conn.Write([]byte(generateHeaders()))
	if err != nil {
		fmt.Println("Error writing headers to connection: ", err.Error())
		return
	}

	_, err = conn.Write([]byte(generateBody()))
	if err != nil {
		fmt.Println("Error writing body to connection: ", err.Error())
		return
	}
}

func generateStatusLine() (s string) {
	s = "HTTP/1.1 200 OK\r\n"
	return
}

func generateHeaders() (h string) {
	h = "\r\n"
	return
}

func generateBody() (h string) {
	return
}
