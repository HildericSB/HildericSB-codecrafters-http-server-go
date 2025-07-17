package main

import (
	"bytes"
	"compress/gzip"
	"flag"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

var FILE_DIRECTORY = "/tmp/"

const (
	CRLF             = "\r\n"
	DEFAULT_PORT     = "4221"
	BUFFER_SIZE      = 4096
	DEFAULT_FILE_DIR = "/tmp/"
)

type request struct {
	path       string
	headers    map[string]string
	connection net.Conn
	method     string
	body       string
}

type response struct {
	statusCode int
	body       string
	headers    map[string]string
	connection net.Conn
}

func main() {

	handleCommandLineFlag()

	// Create TCP listener
	port := DEFAULT_PORT
	listener, err := net.Listen("tcp", "0.0.0.0:"+port)
	if err != nil {
		fmt.Println("Failed to bind to port ", port)
		return
	}
	defer listener.Close()

	fmt.Println("Server listening on :4221")

	// Listen for new connections
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			continue
		}

		go handleConnection(conn)
	}
}

func handleCommandLineFlag() {
	flag.StringVar(&FILE_DIRECTORY, "directory", FILE_DIRECTORY, "specifies the directory where the files are stored, as an absolute path.")
	flag.Parse()
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	for {
		request, err := parseRequest(conn)
		if err != nil {
			fmt.Println("Error parsing request : ", err)
			break
		}

		// If request is nil, keep the connection open and wait for client input
		if request == nil {
			continue
		}

		response := createResponse(request)
		response.sendToClient(request)

		if request.headers["Connection"] == "close" {
			break
		}
	}
}

func parseRequest(conn net.Conn) (*request, error) {

	// Create a buffer and read the HTTP request from connection
	buffer := make([]byte, BUFFER_SIZE)
	_, err := conn.Read(buffer)
	if err != nil {
		if err != io.EOF {
			fmt.Println("Error reading client request from connection")
			return nil, err
		}
		return nil, nil
	}

	req := string(buffer)
	lines := strings.Split(req, "\r\n")

	//Read path and method type
	request := request{
		path:       strings.Split(lines[0], " ")[1],
		method:     strings.Split(lines[0], " ")[0],
		connection: conn,
	}

	// Read headers and body
	headers := make(map[string]string)

	for i, line := range lines[1:] {
		// If line is empty, there is no more header, next line is the body
		if line == "" {
			request.body = lines[i+2]
			break
		}

		headerSplit := strings.SplitN(line, ":", 2)
		for i, v := range headerSplit {
			headerSplit[i] = strings.TrimSpace(v)
		}
		headers[headerSplit[0]] = headerSplit[1]

	}

	request.headers = headers

	return &request, nil
}

func createResponse(request *request) response {
	response := response{
		headers: map[string]string{},
	}

	pathSplit := strings.Split(request.path, "/")
	pathSplitLength := len(pathSplit)

	if request.path == "/" {
		response.statusCode = 200
	} else if pathSplitLength >= 2 {
		switch pathSplit[1] {
		case "echo":
			handleEcho(request, &response)
		case "user-agent":
			handleUserAgent(request, &response)
		case "files":
			if request.method == "GET" {
				handleFileRead(request, &response)
			}

			if request.method == "POST" {
				handleFileUpload(request, &response)
			}
		}

	}

	if request.headers["Connection"] == "close" {
		response.headers["Connection"] = "close"
	}

	if response.statusCode == 0 {
		response.statusCode = 404
	}

	response.connection = request.connection

	return response
}

func (r *response) sendToClient(request *request) error {
	var statusMessage string
	switch r.statusCode {
	case 404:
		statusMessage = "Not Found"
	case 201:
		statusMessage = "Created"
	case 400:
		statusMessage = "Bad request"
	case 200:
		statusMessage = "OK"
	default:
		panic("HTTP statusCode unknown")
	}

	body := r.body
	encodings := request.headers["Accept-Encoding"]

	if strings.Contains(encodings, "gzip") {
		r.headers["Content-Encoding"] = "gzip"
		var buf bytes.Buffer
		zw := gzip.NewWriter(&buf)
		_, err := zw.Write([]byte(body))
		if err != nil {
			fmt.Println("Error encoding the body")
		}
		err = zw.Close()
		if err != nil {
			return err
		}

		body = buf.String()
		r.headers["Content-Length"] = strconv.Itoa(len(body))
	}

	rep := "HTTP/1.1 " + strconv.Itoa(r.statusCode) + " " + statusMessage + CRLF
	for k, v := range r.headers {
		rep = rep + k + ":" + v + CRLF
	}

	rep = rep + CRLF + body

	fmt.Println("\n HTTP reponse : \n" + rep)

	_, err := r.connection.Write([]byte(rep))
	if err != nil {
		fmt.Println("Error writing http response to client ", err)
		return err
	}

	return nil
}

func handleFileUpload(request *request, response *response) {
	contentLength, err := strconv.Atoi(request.headers["Content-Length"])
	if err != nil {
		response.statusCode = http.StatusBadRequest
		return
	}

	if len(request.body) < contentLength {
		response.statusCode = http.StatusBadRequest
		return
	}

	fileName := filepath.Base(request.path)
	filePath := filepath.Join(FILE_DIRECTORY, fileName)
	fileData := request.body[:contentLength]

	err = os.WriteFile(filePath, []byte(fileData), 0666)
	if err != nil {
		fmt.Printf("Error writing file %s: %v\n", filePath, err)
		response.statusCode = http.StatusInternalServerError
		return
	}

	response.statusCode = http.StatusCreated
}
func handleFileRead(request *request, response *response) {
	pathsplit := strings.Split(request.path, "/")

	if len(pathsplit) < 3 {
		response.statusCode = http.StatusBadRequest
		return
	}

	content, err := os.ReadFile(FILE_DIRECTORY + pathsplit[2])
	if err != nil {
		fmt.Println("Error reading file ", FILE_DIRECTORY+pathsplit[2], err)
		response.statusCode = http.StatusNotFound
		return
	}

	response.statusCode = http.StatusOK
	response.headers["Content-Type"] = "application/octet-stream"
	response.headers["Content-Length"] = strconv.Itoa(len(content))
	response.body = string(content)
}

func handleEcho(request *request, response *response) {
	body := strings.TrimPrefix(request.path, "/echo/")
	response.statusCode = http.StatusOK
	response.headers["Content-Type"] = "text/plain"
	response.headers["Content-Length"] = strconv.Itoa(len(body))
	response.body = body
}

func handleUserAgent(request *request, response *response) {
	userAgent := request.headers["User-Agent"]
	if userAgent == "" {
		response.statusCode = http.StatusBadRequest
		return
	}

	response.statusCode = http.StatusOK
	response.headers["Content-Type"] = "text/plain"
	response.headers["Content-Length"] = strconv.Itoa(len(userAgent))
	response.body = userAgent
}
