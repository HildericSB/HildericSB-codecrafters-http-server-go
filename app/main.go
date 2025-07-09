package main

import (
	"fmt"
	"net"
	"os"
	"strings"
)

// Ensures gofmt doesn't remove the "net" and "os" imports above (feel free to remove this!)
var _ = net.Listen
var _ = os.Exit

func main() {

	listener, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		panic(err)
	}
	defer listener.Close()

	fmt.Println("Server listening on :4221")

	conn, err := listener.Accept()
	if err != nil {
		fmt.Println("Error accepting connection: ", err.Error())
		panic(err)
	}
	defer conn.Close()

	buffer := make([]byte, 1024)
	_, err = conn.Read(buffer)
	if err != nil {
		fmt.Println("Error reading connection : ", err)
	}

	req := string(buffer)
	lines := strings.Split(req, "\r\n")
	path := strings.Split(lines[0], " ")[1]

	fmt.Println(path)

	var rep string

	pathSplit := strings.Split(path, "/")

	if path == "/" {
		rep = "HTTP/1.1 200 OK\r\n\r\n"
	} else if len(pathSplit) >= 2 && pathSplit[1] == "echo" {
		var content string
		if len(pathSplit) > 2 {
			content = pathSplit[2]
		}
		rep = fmt.Sprintf(
			"HTTP/1.1 200 OK\r\n"+
				"Content-Type: text/plain\r\n"+
				"Content-Length: %d\r\n"+
				"\r\n"+
				"%s", len(content), content)
	} else {
		rep = "HTTP/1.1 404 Not Found\r\n\r\n"
	}

	fmt.Println(rep)

	conn.Write([]byte(rep))
}
