package main

import (
	"bufio"
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

	// buffer := make([]byte, 1024)
	// n, err := conn.Read(buffer)

	// if err != nil {
	// 	fmt.Println("Error reading connection", err.Error())
	// }

	// fmt.Printf("Bytes read : %d \n", n)

	// Reading the request
	scanner := bufio.NewScanner(conn)

	lineNumber := 1

	var method, path, version string
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			break
		}

		if lineNumber == 1 {
			fields := strings.Fields(line)

			if len(fields) < 3 {
				fmt.Println("ERROR : request is incorect")
			}

			method = fields[0]
			path = fields[1]
			version = fields[2]
		}
		lineNumber++
	}

	if path != "/" {
		conn.Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n"))
		method = method
		version = version
	} else {
		conn.Write([]byte("HTTP/1.1 200 OK\r\n\r\n"))

	}
}
