package main

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
)

type RequestHandler func(r Request) (rsp Response)

var handlers map[string]map[string]RequestHandler

type Request struct {
	method  string
	path    string
	headers map[string]string
	body    string
}

type Response struct {
	body    string
	headers map[string]string
	code    string
}

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
		"POST": {
			"files": uploadFileHandler,
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
		l := strings.Split(h, ":")
		if len(l) < 2 {
			break
		}
		headers[strings.TrimSpace(l[0])] = strings.TrimSpace(l[1])
	}

	r := Request{
		method:  request[0],
		path:    request[1],
		headers: headers,
		body:    sr[len(sr)-1],
	}

	handler, isExist := handlers[r.method][strings.Split(r.path, "/")[1]]

	encoding := "text/plain"
	encoders := strings.Split(r.headers["Accept-Encoding"], ",")
	fmt.Printf("Error writing status to connection: %s\n", encoders)
	for _, enc := range encoders {
		if strings.TrimSpace(enc) == "gzip" {
			encoding = "gzip"
		}
	}

	if isExist == false {
		_, err = conn.Write([]byte(generateResponse(Response{"", nil, "404"}, encoding)))
		if err != nil {
			fmt.Println("Error writing status to connection: ", err.Error())
			return
		}
		return
	}

	_, err = conn.Write([]byte(generateResponse(handler(r), encoding)))
	if err != nil {
		fmt.Println("Error writing status to connection: ", err.Error())
		return
	}
}

func uploadFileHandler(r Request) Response {
	err := os.WriteFile(fmt.Sprintf("%s%s", os.Args[2], strings.Split(r.path, "/")[2]), []byte(r.body), 0644)
	if err != nil {
		return Response{"", nil, "500"}

	}

	return Response{
		"",
		map[string]string{
			"Content-Type": "text/plain",
		},
		"201",
	}
}
func mainPageHandler(_ Request) Response {
	return Response{"", nil, "200"}
}

func userAgentHandler(r Request) Response {
	body := r.headers["User-Agent"]
	return Response{
		body,
		map[string]string{
			"Content-Type":   "text/plain",
			"Content-Length": strconv.Itoa(len(body)),
		},
		"200",
	}
}

func echoHandler(r Request) Response {
	body := strings.Split(r.path, "/")[2]
	return Response{
		body,
		map[string]string{
			"Content-Type":   "text/plain",
			"Content-Length": strconv.Itoa(len(body)),
		},
		"200",
	}
}

func filesHandler(r Request) Response {
	data, err := os.ReadFile(fmt.Sprintf("%s%s", os.Args[2], strings.Split(r.path, "/")[2]))
	if err != nil {
		return Response{
			"",
			map[string]string{
				"Content-Type":   "text/plain",
				"Content-Length": "0",
			},
			"404",
		}
	}

	body := string(data)
	return Response{
		body,
		map[string]string{
			"Content-Type":   "application/octet-stream",
			"Content-Length": strconv.Itoa(len(body)),
		},
		"200",
	}
}

func generateResponse(r Response, encoding string) string {
	var builder strings.Builder

	if encoding == "gzip" {
		var buf bytes.Buffer
		zw := gzip.NewWriter(&buf)
		_, err := zw.Write([]byte(r.body))
		if err != nil {
			fmt.Println("Error writing status to connection: ", err.Error())
		}
		r.headers["Content-Encoding"] = "gzip"
		r.body = buf.String()
	}

	builder.WriteString(generateStatusLine(r.code))
	builder.WriteString(generateHeaders(r.headers))
	builder.WriteString(generateBody(r.body))
	return builder.String()
}

func generateStatusLine(code string) string {
	var status string
	switch code {
	case "200":
		status = "OK"
	case "201":
		status = "Created"
	case "404":
		status = "Not Found"
	default:
		status = "Error"
	}

	return fmt.Sprintf("HTTP/1.1 %s %s\r\n", code, status)
}

func generateHeaders(headers map[string]string) string {
	var h strings.Builder
	for k, v := range headers {
		h.WriteString(fmt.Sprintf("%s: %s\r\n", k, v))
	}
	h.WriteString("\r\n")
	return h.String()
}

func generateBody(context string) string {
	return context
}
