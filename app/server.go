package main

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
)

type RequestHandler func(req string, userAgent map[string]string) (body string, header map[string]string, code string)

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
			"":           mainPageHandler,
			"echo":       echoHandler,
			"user-agent": userAgentHandler,
			"files":      filesHandler,
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

	sr := strings.Split(string(buf[:l]), "\r\n")
	headers := make(map[string]string)
	request := strings.Fields(sr[0])

	for _, h := range sr[1:] {
		l := strings.Fields(h)
		if len(l) < 2 {
			continue
		}
		headers[l[0][:len(l[0])-1]] = l[1]
	}

	handler, isExist := handlers[request[0]][strings.Split(request[1], "/")[1]]
	if isExist == false {
		_, err = conn.Write([]byte(generateNotFoundResponse()))
		if err != nil {
			fmt.Println("Error writing status to connection: ", err.Error())
			return
		}
		return
	}

	_, err = conn.Write([]byte(generateSuccessResponse(handler(request[1], headers))))
	if err != nil {
		fmt.Println("Error writing status to connection: ", err.Error())
		return
	}
}

func mainPageHandler(_ string, _ map[string]string) (body string, header map[string]string, code string) {
	return "", nil, "200"
}

func userAgentHandler(_ string, userAgent map[string]string) (body string, header map[string]string, code string) {
	body = userAgent["User-Agent"]
	return body,
		map[string]string{
			"Content-Type":   "text/plain",
			"Content-Length": strconv.Itoa(len(body)),
		},
		"200"
}

func echoHandler(req string, _ map[string]string) (body string, header map[string]string, code string) {
	body = strings.Split(req, "/")[2]
	return body,
		map[string]string{
			"Content-Type":   "text/plain",
			"Content-Length": strconv.Itoa(len(body)),
		},
		"200"
}

func filesHandler(req string, _ map[string]string) (body string, header map[string]string, code string) {
	data, err := os.ReadFile(fmt.Sprintf("%s%s", os.Args[2], strings.Split(req, "/")[2]))
	if err != nil {
		return "",
			map[string]string{
				"Content-Type":   "text/plain",
				"Content-Length": "0",
			},
			"404"
	}

	body = string(data)
	return body,
		map[string]string{
			"Content-Type":   "application/octet-stream",
			"Content-Length": strconv.Itoa(len(body)),
		},
		"200"
}

func generateSuccessResponse(body string, header map[string]string, code string) string {
	return generateResponse(body, header, code)
}

func generateNotFoundResponse() string {
	return generateResponse("", nil, "404")
}

func generateResponse(body string, header map[string]string, code string) string {
	var builder strings.Builder
	builder.WriteString(generateStatusLine(code))
	builder.WriteString(generateHeaders(header))
	builder.WriteString(generateBody(body))
	return builder.String()
}

func generateStatusLine(code string) string {
	var status string
	switch code {
	case "200":
		status = "OK"
	case "404":
		status = "Not Found"
	default:
		status = "Error"
	}

	return fmt.Sprintf("HTTP/1.1 %s %s\r\n", code, status)
}

func generateHeaders(body map[string]string) string {
	var h strings.Builder
	for k, v := range body {
		h.WriteString(fmt.Sprintf("%s: %s\r\n", k, v))
	}
	h.WriteString("\r\n")
	return h.String()
}

func generateBody(context string) string {
	return context
}
