package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
)

var fileTypes = map[string]string{
	".html": "text/html",
	".txt":  "text/plain",
	".gif":  "image/gif",
	".jpeg": "image/jpeg",
	".jpg":  "image/jpeg",
	".css":  "text/css",
}

var sem = make(chan int, 10)

func handleConnection(client net.Conn) {
	defer client.Close()

	// Store the HTTP-request in buf.
	buf := make([]byte, 2048)
	_, err := client.Read(buf)
	if err != nil {
		fmt.Println("Error reading from client:", err)
		return
	}

	// Parse the request from the buffer.
	rq, err := http.ReadRequest(bufio.NewReader(bytes.NewReader(buf)))
	if err != nil {
		fmt.Println("Error parsing HTTP request:", err)
		return
	}
	// Handle requests based on the method.
	switch rq.Method {
	case "GET", "POST":
		serveRequest(rq, client)
	case "HEAD", "PUT", "DELETE", "OPTIONS", "TRACE", "PATCH":
		createResponse(rq, client, "501 Not Implemented", false)
		return
	default:
		createResponse(rq, client, "400 Bad Request", false)
		return
	}
	<-sem
}

func serveRequest(r *http.Request, client net.Conn) {
	path := "C:/Users/wiggo/TDA596/lab1/http_server" + r.URL.Path
	if !validFile(path) {
		createResponse(r, client, "400 Bad Request", false)
		return
	}
	if r.Method == "GET" {
		_, err := os.Stat(path)
		if err != nil {
			createResponse(r, client, "404 Not Found", false)
			return
		}
		// If the GET file is valid and exists on the server, open and copy it to the client. Also send a header.

		file, err := os.Open(path)
		if err != nil {
			fmt.Println("Error reading file on server:", err)
			return
		}
		//defer file.Close()
		fmt.Println("wow it work")
		createResponse(r, client, "200 OK", true)

		_, err = io.Copy(client, file)
		if err != nil {
			fmt.Println("Error transferring file back:", err)
		}

	} else {
		// Create or modify newFile
		newFile, err := os.Create(strings.TrimLeft(r.URL.Path, "/"))
		if err != nil {
			fmt.Println("Error creating new file", err)
		}
		// Read data from the body of the request
		bd, err := io.ReadAll(r.Body)
		if err != nil {
			fmt.Println("Error reading request body:", err)
			return
		}

		createResponse(r, client, "201 Created", true)
		// Write the data read from the body to the newFile created
		_, err = newFile.Write(bd)
		if err != nil {
			fmt.Println("Error writing to new file", err)
		}
	}
}

// For GET/POST requests, checks if the filetypes of the requests are allowed.
func validFile(path string) bool {
	lastIndex := strings.LastIndex(path, ".")
	if lastIndex == -1 {
		fmt.Println("Invalid filetype")
		return false
	}
	fileType := path[lastIndex:]
	for key := range fileTypes {
		if key == fileType {
			return true
		}
	}
	return false
}

func contentType(path string) string {
	lastIndex := strings.LastIndex(path, ".")
	if lastIndex == -1 {
		fmt.Println("Invalid content-type")

	}
	fileType := path[lastIndex:]
	return fileTypes[fileType]
}

// Write to the client the headerstring for example "HTTP/1.1 200 OK"
func createResponse(r *http.Request, client net.Conn, status string, b bool) {
	var response http.Response
	path := "C:/Users/wiggo/TDA596/lab1/http_server" + r.URL.Path
	response.Proto = r.Proto
	response.Status = status
	if b {
		client.Write([]byte(response.Proto + " " + response.Status + "\r\n" + "Content-Type: " + contentType(path) + "\r\n\r\n"))
	} else {
		client.Write([]byte(response.Proto + " " + response.Status + "\r\n\r\n"))
	}
}

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Input port number as first argument")
		return
	}
	// Listen for TCP on any IP and on the port specified by the user
	ln, err := net.Listen("tcp", ":"+os.Args[1])
	if err != nil {
		log.Fatalln("Error starting server:", err)
	}
	fmt.Println("HTTP server started on port:", os.Args[1])

	// Accept at most 10 connections (bounded channel)
	for {
		client, err := ln.Accept()
		if err != nil {
			fmt.Println("Error connecting client to server:", err)
			continue
		}
		fmt.Println("Client connected with address:", client.RemoteAddr())
		sem <- 1
		go handleConnection(client)
	}

}
