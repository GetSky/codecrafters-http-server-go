package main

import (
	"fmt"
	"net"
	"os"
	"strings"
)

type RequestHandler func(req string) (body string)

var handlers map[string]map[string]RequestHandler

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

	handlers = map[string]map[string]RequestHandler{
		"GET": {
			"":     mainPageHandler,
			"echo": echoHandler,
		},
	}

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

	request := strings.Fields(strings.Split(string(buf[:l]), "\r\n")[0])
	handler, isExist := handlers[request[0]][strings.Split(request[1], "/")[1]]
	if isExist == false {
		_, err = conn.Write([]byte(generateNotFoundResponse()))
		if err != nil {
			fmt.Println("Error writing status to connection: ", err.Error())
			return
		}
		return
	}

	_, err = conn.Write([]byte(generateSuccessResponse(handler(request[1]))))
	if err != nil {
		fmt.Println("Error writing status to connection: ", err.Error())
		return
	}
}

func mainPageHandler(_ string) (body string) {
	return
}

func echoHandler(req string) (body string) {
	return strings.Split(req, "/")[2]
}

func generateSuccessResponse(body string) string {
	return generateResponse(body, 200, "OK")
}

func generateNotFoundResponse() string {
	return generateResponse("", 404, "Not Found")
}

func generateResponse(body string, code int, status string) string {
	var builder strings.Builder
	builder.WriteString(generateStatusLine(code, status))
	builder.WriteString(generateHeaders(body))
	builder.WriteString(generateBody(body))
	return builder.String()
}

func generateStatusLine(code int, status string) string {
	return fmt.Sprintf("HTTP/1.1 %d %s\r\n", code, status)
}

func generateHeaders(body string) string {
	return fmt.Sprintf("Content-Type: text/plain\r\nContent-Length: %d\r\n\r\n", len(body))
}

func generateBody(context string) string {
	return context
}
