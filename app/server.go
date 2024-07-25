package main

import (
	"fmt"
	"net"
	"os"
	"strings"
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
		go handleRequest(conn)
	}
}

func handleRequest(conn net.Conn) {
	defer func(conn net.Conn) {
		_ = conn.Close()
	}(conn)

	buf := make([]byte, 4096)
	l, err := conn.Read(buf)
	if err != nil {
		fmt.Println("Error reading:", err.Error())
		return
	}

	request := strings.Split(string(buf[:l]), "\r\n")
	fmt.Printf("Message received: %s\n", request)

	if checkRequest(request[0]) {
		_, err = conn.Write([]byte(generateSuccessResponse()))

	} else {
		_, err = conn.Write([]byte(generateNotFoundResponse()))
	}

	if err != nil {
		fmt.Println("Error writing status to connection: ", err.Error())
		return
	}
}

func checkRequest(req string) bool {
	requestLine := strings.Fields(req)
	if len(requestLine) < 3 {
		fmt.Println("Invalid request.")
		return false
	}
	if requestLine[0] != "GET" || requestLine[1] != "/" {
		fmt.Println("Not found page.")
		return false
	}

	return true
}

func generateSuccessResponse() string {
	return generateResponse(200, "OK")
}

func generateNotFoundResponse() string {
	return generateResponse(404, "Not Found")
}

func generateResponse(code int, status string) string {
	var builder strings.Builder
	builder.WriteString(generateStatusLine(code, status))
	builder.WriteString(generateHeaders())
	builder.WriteString(generateBody())
	return builder.String()
}

func generateStatusLine(code int, status string) string {
	return fmt.Sprintf("HTTP/1.1 %d %s\r\n", code, status)
}

func generateHeaders() (h string) {
	h = "\r\n"
	return
}

func generateBody() (h string) {
	return
}
