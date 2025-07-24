package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"os/signal"
	"time"

	"github.com/codecrafters-io/http-server-starter-go/config"
	"github.com/codecrafters-io/http-server-starter-go/handler"
	"github.com/codecrafters-io/http-server-starter-go/http"
	"github.com/codecrafters-io/http-server-starter-go/router"
)

var FILE_DIRECTORY = config.DEFAULT_FILE_DIR

type Server struct {
	port        string
	fileDir     string
	listener    net.Listener
	router      *router.Router
	isUp        bool
	connections map[net.Conn]bool // map for O(1) removal
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
		isUp:    false,
	}

	router.Handle("/files", func(req *http.Request, resp *http.Response) {
		handler.HandleFile(req, resp, server.fileDir)
	})
	router.Handle("/echo", handler.HandleEcho)
	router.Handle("/user-agent", handler.HandleUserAgent)

	return &server, nil
}

func (s *Server) gracefulShutdownRoutine() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
	fmt.Printf("Shutting down signal gracefully shutting down")

	s.ShutDown()
}

func (s *Server) ShutDown() {
	s.isUp = false

	// Wait a bit for current requests to complete
	time.Sleep(5 * time.Second)

	for conn := range s.connections {
		conn.Close()
	}

	s.listener.Close()
}

func (s *Server) Start() error {
	s.isUp = true
	listener, err := net.Listen("tcp", "0.0.0.0:"+s.port)
	if err != nil {
		fmt.Println("Failed to bind to port ", s.port)
		return err
	}
	s.listener = listener
	fmt.Println("Server listening on :", s.port)

	go s.gracefulShutdownRoutine()

	// Listen for new connections
	for s.isUp {
		conn, err := listener.Accept()
		if err != nil {
			if !s.isUp {
				// Server is now down, error can happen
				break
			}
			fmt.Println("Error accepting connection: ", err.Error())
			continue
		}

		go s.handleConnection(conn)
	}

	return nil
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
	s.connections[conn] = true
	defer func() {
		conn.Close()
		delete(s.connections, conn)
	}()

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
