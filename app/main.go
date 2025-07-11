package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"strings"
)

// Ensures gofmt doesn't remove the "net" and "os" imports above (feel free to remove this!)
var _ = net.Listen
var _ = os.Exit

var FILE_DIRECTORY = "/"

const CRLF = "\r\n"

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
	flag.StringVar(&FILE_DIRECTORY, "directory", "/", "specifies the directory where the files are stored, as an absolute path.")
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
	path := strings.Split(lines[0], " ")[1]

	// Put HTTP headers in a map
	headers := make(map[string]string)
	for _, line := range lines[1:] {
		// If line is empty, there is no more header
		if line == "" {
			break
		}
		headerSplit := strings.SplitN(line, ":", 2)
		for i, v := range headerSplit {
			headerSplit[i] = strings.TrimSpace(v)
		}
		headers[headerSplit[0]] = headerSplit[1]
	}

	fmt.Println("Path : " + path)
	fmt.Println("Headers : \n", headers)

	var rep string

	pathSplit := strings.Split(path, "/")
	pathSplitLength := len(pathSplit)

	if path == "/" {
		rep = "HTTP/1.1 200 OK\r\n\r\n"
	} else if pathSplitLength >= 2 {
		switch pathSplit[1] {
		case "echo":
			rep = handleEcho(pathSplit)
		case "user-agent":
			rep = handleUserAgent(headers)
		case "files":
			rep = handleFileRead(pathSplit)
		}

	}

	if rep == "" {
		rep = "HTTP/1.1 404 Not Found\r\n\r\n"
	}

	fmt.Println(rep)

	conn.Write([]byte(rep))
}

func handleFileRead(pathsplit []string) string {
	if len(pathsplit) < 3 {
		fmt.Println("Not enough arg in url for file reading")
		return "HTTP/1.1 400 Bad Request\r\n\r\n"
	}

	fileContent, err := os.ReadFile(FILE_DIRECTORY + pathsplit[2])
	if err != nil {
		fmt.Println("Error reading file ", pathsplit[2])
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
