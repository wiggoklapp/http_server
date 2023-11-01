package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strings"
)

type Request struct {
	method      string
	path        string
	fileName    string
	contentType string
	body        []byte
}

var allowedFileTypes = []string{".html", ".txt", ".gif", ".jpeg", ".jpg", ".css"}

func startServer(listenPort string) (net.Listener, error) {
	ln, err := net.Listen("tcp", ":"+listenPort)
	if err != nil {
		fmt.Println("Error starting server:", err)
	} else {
		fmt.Println("Server started on port:", listenPort)
	}
	return ln, err
}

func acceptConns(ln net.Listener) (net.Conn, error) {
	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("Error connecting client:", err)
			continue
		}
		fmt.Println("Client connected with address:", conn.RemoteAddr())
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()
	// Store the HTTP-request in buf
	buf := make([]byte, 2048)
	_, err := conn.Read(buf)
	if err != nil {
		fmt.Println("Error reading from client:", err)
		return
	}

	rq, err := http.ReadRequest(bufio.NewReader(bytes.NewReader(buf)))
	if err != nil {
		fmt.Println("Error parsing HTTP request:", err)
		return
	}
	bd, err := io.ReadAll(rq.Body)
	if err != nil {
		fmt.Println("Error reading request body:", err)
		return
	}
	// Create instance of a HTTP request
	request := Request{
		method:      rq.Method,
		path:        "C:/Users/wiggo/TDA596/http_server" + rq.URL.Path,
		fileName:    strings.TrimLeft(rq.URL.Path, "/"),
		contentType: rq.Header.Get("Content-Type"),
		body:        bd,
	}

	switch rq.Method {
	case "GET":
		if !request.validFile() {
			fmt.Println("File not allowed - bad Request (400)")
		} else if request.noFile() {
			fmt.Println("File not found (404)")
		} else {
			fmt.Println(request.contentType)
			serveGetRequest(request, conn)
		}
	case "POST":
		if !request.validFile() {
			fmt.Println("File not allowed - bad Request (400)")
		} else {
			servePostRequest(request)
		}
	case "HEAD", "PUT", "DELETE", "OPTIONS", "TRACE", "PATCH":
		fmt.Println("Not Implemented (501)")
		return
	default:
		fmt.Println("Invalid method - Bad Request (400)")
		return
	}

}

func (r Request) validFile() bool {
	lastIndex := strings.LastIndex(r.path, ".")
	if lastIndex == -1 {
		return false
	}
	fileType := r.path[lastIndex:]
	for _, file := range allowedFileTypes {
		if file == fileType {
			return true
		}
	}
	return false
}

func (r Request) noFile() bool {
	_, err := os.Stat(r.path)
	return os.IsNotExist(err)
}

func serveGetRequest(r Request, conn net.Conn) {

	data, err := os.Open(r.path)
	if err != nil {
		fmt.Println("Error reading file on server:", err)
		return
	}
	responseHeader := "HTTP/1.1 200 OK\r\nContent-Type:" + r.contentType + " \r\n\r\n"
	conn.Write([]byte(responseHeader))
	fmt.Println(responseHeader)

	_, err = io.Copy(conn, data)
	if err != nil {
		fmt.Println("Error transferring file back:", err)
	}

}

func servePostRequest(r Request) {
	if r.noFile() {
		fmt.Println(r.body)

		newFile, err := os.Create(r.fileName)
		if err != nil {
			fmt.Println("Error creating new file", err)
		}
		_, err = newFile.Write(r.body)
		if err != nil {
			fmt.Println("Error writing to new file", err)
		}
	} else {
		//modify the existing file with the new data
	}
}

func main() {

	if len(os.Args) != 2 {
		fmt.Println("Takes port number as argument")
		return
	}

	ln, _ := startServer(os.Args[1])
	acceptConns(ln)
}
