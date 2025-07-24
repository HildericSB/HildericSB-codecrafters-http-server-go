package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/codecrafters-io/http-server-starter-go/config"
	"github.com/codecrafters-io/http-server-starter-go/http"
)

var FILE_DIRECTORY = "/tmp/"

type Server struct {
	port     string
	fileDir  string
	listener net.Listener
}

func NewServer(port string) (*Server, error) {
	// Create TCP listener
	if port == "" {
		port = config.DEFAULT_PORT
	}

	return &Server{
		port: port,
	}, nil
}

func (s *Server) Start() error {
	listener, err := net.Listen("tcp", "0.0.0.0:"+s.port)
	if err != nil {
		fmt.Println("Failed to bind to port ", s.port)
		return err
	}

	s.listener = listener

	fmt.Println("Server listening on :", s.port)

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

func (s *Server) Stop() {
	s.listener.Close()
}

func main() {
	handleCommandLineFlag()

	server, err := NewServer("4221")
	if err != nil {
		fmt.Println(err)
		return
	}

	err = server.Start()
	if err != nil {
		fmt.Println(err)
		return
	}

}

func handleCommandLineFlag() {
	flag.StringVar(&FILE_DIRECTORY, "directory", FILE_DIRECTORY, "specifies the directory where the files are stored, as an absolute path.")
	flag.Parse()
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	for {
		request, err := http.ParseRequest(conn)
		if err != nil {
			fmt.Println("Error parsing request : ", err)
			break
		}

		// If request is nil, keep the connection open and wait for client input
		if request == nil {
			continue
		}

		response := NewResponse(request)
		response.SendToClient(request)

		if request.Headers["Connection"] == "close" {
			break
		}
	}
}

func HandleFileUpload(request *http.Request, response *http.Response) {
	contentLength, err := strconv.Atoi(request.Headers["Content-Length"])
	if err != nil {
		response.StatusCode = 400
		return
	}

	if len(request.Body) < contentLength {
		response.StatusCode = 400
		return
	}

	fileName := filepath.Base(request.Path)
	filePath := filepath.Join(config.DEFAULT_FILE_DIR, fileName)
	fileData := request.Body[:contentLength]

	err = os.WriteFile(filePath, []byte(fileData), 0666)
	if err != nil {
		fmt.Printf("Error writing file %s: %v\n", filePath, err)
		response.StatusCode = 500
		return
	}

	response.StatusCode = 201
}
func HandleFileRead(request *http.Request, response *http.Response) {
	pathsplit := strings.Split(request.Path, "/")

	if len(pathsplit) < 3 {
		response.StatusCode = 400
		return
	}

	content, err := os.ReadFile(config.DEFAULT_FILE_DIR + pathsplit[2])
	if err != nil {
		fmt.Println("Error reading file ", config.DEFAULT_FILE_DIR+pathsplit[2], err)
		response.StatusCode = 404
		return
	}

	response.StatusCode = 200
	response.Headers["Content-Type"] = "application/octet-stream"
	response.Headers["Content-Length"] = strconv.Itoa(len(content))
	response.Body = string(content)
}

func HandleEcho(request *http.Request, response *http.Response) {
	body := strings.TrimPrefix(request.Path, "/echo/")
	response.StatusCode = 200
	response.Headers["Content-Type"] = "text/plain"
	response.Headers["Content-Length"] = strconv.Itoa(len(body))
	response.Body = body
}

func HandleUserAgent(request *http.Request, response *http.Response) {
	userAgent := request.Headers["User-Agent"]
	if userAgent == "" {
		response.StatusCode = 400
		return
	}

	response.StatusCode = 200
	response.Headers["Content-Type"] = "text/plain"
	response.Headers["Content-Length"] = strconv.Itoa(len(userAgent))
	response.Body = userAgent
}

func NewResponse(request *http.Request) *http.Response {
	response := &http.Response{
		Headers: map[string]string{},
	}

	pathSplit := strings.Split(request.Path, "/")
	pathSplitLength := len(pathSplit)

	if request.Path == "/" {
		response.StatusCode = 200
	} else if pathSplitLength >= 2 {
		switch pathSplit[1] {
		case "echo":
			HandleEcho(request, response)
		case "user-agent":
			HandleUserAgent(request, response)
		case "files":
			if request.Method == "GET" {
				HandleFileRead(request, response)
			}

			if request.Method == "POST" {
				HandleFileUpload(request, response)
			}
		}

	}

	if request.Headers["Connection"] == "close" {
		response.Headers["Connection"] = "close"
	}

	if response.StatusCode == 0 {
		response.StatusCode = 404
	}

	response.Connection = request.Connection

	return response
}
