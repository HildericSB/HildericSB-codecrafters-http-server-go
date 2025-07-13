package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// Ensures gofmt doesn't remove the "net" and "os" imports above (feel free to remove this!)
var _ = net.Listen
var _ = os.Exit

var FILE_DIRECTORY = "/tmp/"

const CRLF = "\r\n"

type request struct {
	path       string
	headers    map[string]string
	connection net.Conn
	method     string
	body       string
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

	buffer := make([]byte, 1024)
	_, err := conn.Read(buffer)
	if err != nil {
		fmt.Println("Error reading connection : ", err)
	}

	req := string(buffer)
	lines := strings.Split(req, "\r\n")
	request := request{
		path:       strings.Split(lines[0], " ")[1],
		method:     strings.Split(lines[0], " ")[0],
		connection: conn,
	}
	parseRequest(&request, lines)

	fmt.Println("Path : " + request.path)
	fmt.Println("Headers : \n", request.headers)

	var rep string

	pathSplit := strings.Split(request.path, "/")
	pathSplitLength := len(pathSplit)

	if request.path == "/" {
		rep = "HTTP/1.1 200 OK\r\n\r\n"
	} else if pathSplitLength >= 2 {
		switch pathSplit[1] {
		case "echo":
			rep = handleEcho(pathSplit)
		case "user-agent":
			rep = handleUserAgent(request.headers)
		case "files":
			if request.method == "GET" {
				rep = handleFileRead(pathSplit)
			}

			if request.method == "POST" {
				rep = handleFileUpload(request)
			}
		}

	}

	if rep == "" {
		rep = "HTTP/1.1 404 Not Found\r\n\r\n"
	}

	fmt.Println(rep)

	conn.Write([]byte(rep))
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func handleFileUpload(req request) string {
	i, err := strconv.Atoi(req.headers["Content-Length"])
	check(err)

	fileData := req.body[:i]

	err = os.WriteFile(FILE_DIRECTORY+filepath.Base(req.path), []byte(fileData), 0666)
	check(err)

	return "HTTP/1.1 201 Created\r\n\r\n"
}

func parseRequest(request *request, lines []string) {
	// Put HTTP headers in a map
	headers := make(map[string]string)

	bodyMode := false
	for _, line := range lines[1:] {
		// If line is empty, there is no more header, it's the body now
		if line == "" {
			bodyMode = true
			continue
		}

		if !bodyMode {
			headerSplit := strings.SplitN(line, ":", 2)
			for i, v := range headerSplit {
				headerSplit[i] = strings.TrimSpace(v)
			}
			headers[headerSplit[0]] = headerSplit[1]
		} else {
			request.body = line
		}

	}

	request.headers = headers
}

func handleFileRead(pathsplit []string) string {
	if len(pathsplit) < 3 {
		fmt.Println("Not enough arg in url for file reading")
		return "HTTP/1.1 400 Bad Request\r\n\r\n"
	}

	fileContent, err := os.ReadFile(FILE_DIRECTORY + pathsplit[2])
	if err != nil {
		fmt.Println("Error reading file ", FILE_DIRECTORY+pathsplit[2], err)
		return "HTTP/1.1 404 Not Found\r\n\r\n"
	}

	return fmt.Sprintf(
		"HTTP/1.1 200 OK\r\n"+
			"Content-Type: application/octet-stream\r\n"+
			"Content-Length: %d\r\n"+
			"\r\n"+
			"%s", len(fileContent), fileContent)
}

func handleEcho(pathSplit []string) string {
	content := pathSplit[2]
	rep := fmt.Sprintf(
		"HTTP/1.1 200 OK\r\n"+
			"Content-Type: text/plain\r\n"+
			"Content-Length: %d\r\n"+
			"\r\n"+
			"%s", len(content), content)

	return rep
}

func handleUserAgent(headers map[string]string) string {
	if headers["User-Agent"] == "" {
		fmt.Println("No user-agent headers")
	}

	return fmt.Sprintf(
		"HTTP/1.1 200 OK\r\n"+
			"Content-Type: text/plain\r\n"+
			"Content-Length: %d\r\n"+
			"\r\n"+
			"%s", len(headers["User-Agent"]), headers["User-Agent"])
}
