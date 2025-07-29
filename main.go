package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/codecrafters-io/http-server-starter-go/config"
	"github.com/codecrafters-io/http-server-starter-go/handler"
	"github.com/codecrafters-io/http-server-starter-go/http"
	"github.com/codecrafters-io/http-server-starter-go/router"
)

var FILE_DIRECTORY = config.DEFAULT_FILE_DIR

type Server struct {
	port                      string
	fileDir                   string
	listener                  net.Listener
	router                    *router.Router
	isUp                      bool
	connections               map[net.Conn]bool // map for O(1) removal
	connMutex                 sync.Mutex
	connectionWaitGroup       sync.WaitGroup // To track number of running connections
	numberOfConnectionsWorker int
	connectionsChan           chan net.Conn
	shutdownChan              chan struct{}
}

func NewServer(port string) (*Server, error) {
	// Create TCP listener
	if port == "" {
		port = config.DEFAULT_PORT
	}

	router := router.NewRouter()

	server := Server{
		port:                      port,
		fileDir:                   FILE_DIRECTORY,
		router:                    router,
		isUp:                      false,
		connections:               make(map[net.Conn]bool),
		shutdownChan:              make(chan struct{}),
		connectionsChan:           make(chan net.Conn, 100), // Using buffered chan so if new connections queue up, Accept() continues accepting
		numberOfConnectionsWorker: 10,
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
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c
	fmt.Printf("Shutting down signal gracefully shutting down...\n")

	s.ShutDown()
}

func (s *Server) ShutDown() {
	if !s.isUp {
		return
	}
	s.isUp = false

	// Signal ShutDown to worker
	close(s.shutdownChan)

	if s.listener != nil {
		s.listener.Close()
	}

	// Wait for all connections to finish with timeout
	done := make(chan struct{})
	go func() {
		s.connectionWaitGroup.Wait()
		close(done)
	}()

	select {
	case <-done:
		fmt.Println("All connections closed gracefully")
	case <-time.After(10 * time.Second):
		fmt.Println("Timer is finished")

		// Force close remaining connections
		s.connMutex.Lock()
		for conn := range s.connections {
			conn.Close()
			delete(s.connections, conn)
		}
		s.connMutex.Unlock()
	}

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

	for i := range s.numberOfConnectionsWorker {
		go s.startConnectionHandler(i)
	}

	// Listen for new connections
	for s.isUp {
		conn, err := listener.Accept()
		if err != nil {
			if !s.isUp {
				// Server is shutting down, this is expected
				break
			}
			fmt.Println("Error accepting connection: ", err.Error())
			continue
		}

		select {
		case s.connectionsChan <- conn:
			// Connection sent to worker ppol
		default:
			// Channel is full, reject the connection
			fmt.Println("Worker pool busy, rejecting connection")
			conn.Close()
		}

	}

	// Server is shutting down, wait for active connections to finish
	s.connectionWaitGroup.Wait()

	return nil
}

func (s *Server) startConnectionHandler(workerID int) {
	for {
		select {
		case conn := <-s.connectionsChan:
			fmt.Printf("Worker %d handling connection\n", workerID)
			s.handleConnection(conn)
		case <-s.shutdownChan:
			fmt.Printf("Worker %d shutting down\n", workerID)
			return
		}
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
	s.connMutex.Lock()
	s.connections[conn] = true
	s.connMutex.Unlock()
	s.connectionWaitGroup.Add(1)

	defer func() {
		s.connectionWaitGroup.Done()
		s.connMutex.Lock()
		conn.Close()
		delete(s.connections, conn)
		s.connMutex.Unlock()
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
