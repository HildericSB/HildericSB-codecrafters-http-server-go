package main

import (
	"flag"
	"fmt"
	"net"

	"github.com/codecrafters-io/http-server-starter-go/config"
	"github.com/codecrafters-io/http-server-starter-go/handler"
	"github.com/codecrafters-io/http-server-starter-go/http"
	"github.com/codecrafters-io/http-server-starter-go/router"
)

var FILE_DIRECTORY = config.DEFAULT_FILE_DIR

type Server struct {
	port     string
	fileDir  string
	listener net.Listener
	router   *router.Router
}

func NewServer(port string) (*Server, error) {
	// Create TCP listener
	if port == "" {
		port = config.DEFAULT_PORT
	}

	router := router.NewRouter()

	server := Server{
		port:    port,
		fileDir: FILE_DIRECTORY,
		router:  router,
	}

	router.Handle("/files", func(req *http.Request, resp *http.Response) {
		handler.HandleFile(req, resp, server.fileDir)
	})
	router.Handle("/echo", handler.HandleEcho)
	router.Handle("/user-agent", handler.HandleUserAgent)

	return &server, nil
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

		go s.handleConnection(conn)
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

func (s *Server) handleConnection(conn net.Conn) {
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

		s.router.ServeHTTP(request, response)

		response.SendToClient(request)

		if request.Headers["Connection"] == "close" {
			break
		}
	}
}
