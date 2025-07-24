package main

import (
	"flag"
	"fmt"
	"net"

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

		response := http.NewResponse(request)
		response.SendToClient(request)

		if request.Headers["Connection"] == "close" {
			break
		}
	}
}
