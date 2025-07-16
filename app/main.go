package main

import (
	"bytes"
	"compress/gzip"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

var FILE_DIRECTORY = "/tmp/"

const CRLF = "\r\n"

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

	// Create TCP listener on 4221
	listener, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		panic(err)
	}
	defer listener.Close()

	fmt.Println("Server listening on :4221")

	// Listen for new connections
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
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
		}

		// If request is nil, keep the connection open and wait for client input
		if request == nil {
			continue
		}

		response := createResponse(*request)
		response.sendToClient(request)

		if request.headers["Connection"] == "close" {
			break
		}
	}
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func handleFileUpload(req request) response {
	i, err := strconv.Atoi(req.headers["Content-Length"])
	check(err)

	fileData := req.body[:i]

	err = os.WriteFile(FILE_DIRECTORY+filepath.Base(req.path), []byte(fileData), 0666)
	check(err)

	return response{
		statusCode: 201,
	}
}

func parseRequest(conn net.Conn) (*request, error) {

	// Create a buffer and read the HTTP request from connection
	buffer := make([]byte, 1024)
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

func createResponse(request request) response {
	var response response
	pathSplit := strings.Split(request.path, "/")
	pathSplitLength := len(pathSplit)

	if request.path == "/" {
		response.statusCode = 200
	} else if pathSplitLength >= 2 {
		switch pathSplit[1] {
		case "echo":
			response = handleEcho(request)
		case "user-agent":
			response = handleUserAgent(request)
		case "files":
			if request.method == "GET" {
				response = handleFileRead(request)
			}

			if request.method == "POST" {
				response = handleFileUpload(request)
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

func handleFileRead(request request) response {
	pathsplit := strings.Split(request.path, "/")

	if len(pathsplit) < 3 {
		fmt.Println("Not enough arg in url for file reading")
		return badRequest400Reponse()
	}

	fileContent, err := os.ReadFile(FILE_DIRECTORY + pathsplit[2])
	if err != nil {
		fmt.Println("Error reading file ", FILE_DIRECTORY+pathsplit[2], err)
		return notFound404Reponse()
	}

	return response{
		statusCode: 200,
		headers: map[string]string{
			"Content-Type":   "application/octet-stream",
			"Content-Length": strconv.Itoa(len(fileContent)),
		},
		body: string(fileContent),
	}
}

func handleEcho(request request) response {
	body := strings.TrimPrefix(request.path, "/echo/")
	return response{
		statusCode: 200,
		headers: map[string]string{
			"Content-Type":   "text/plain",
			"Content-Length": strconv.Itoa(len(body)),
		},
		body: body,
	}
}

func handleUserAgent(request request) response {
	userAgentHeader := request.headers["User-Agent"]
	if userAgentHeader == "" {
		fmt.Println("Error : No user-agent headers")
		return response{
			statusCode: 300,
		}
	}

	return response{
		statusCode: 200,
		headers: map[string]string{
			"Content-Type":   "text/plain",
			"Content-Length": strconv.Itoa(len(userAgentHeader)),
		},
		body: userAgentHeader,
	}
}

func notFound404Reponse() response {
	return response{
		statusCode: 404,
	}
}

func badRequest400Reponse() response {
	return response{
		statusCode: 400,
	}
}

func (r *response) sendToClient(request *request) error {
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
		check(err)

		body = buf.String()
		r.headers["Content-Length"] = strconv.Itoa(len(body))
	}

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
