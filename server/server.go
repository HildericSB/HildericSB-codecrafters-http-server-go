package server

import (
	"fmt"
	"io"
	"net"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/codecrafters-io/http-server-starter-go/config"
	"github.com/codecrafters-io/http-server-starter-go/handler"
	"github.com/codecrafters-io/http-server-starter-go/http"
	"github.com/codecrafters-io/http-server-starter-go/middleware"
	"github.com/codecrafters-io/http-server-starter-go/router"
)

type Server struct {
	Port                      string
	startTime                 time.Time
	fileDir                   string
	listener                  net.Listener
	router                    *router.Router
	numberOfConnectionsWorker int
	connectionsChan           chan net.Conn
	connectionWaitGroup       sync.WaitGroup
	openConnections           int64
	shutDownSignal            chan struct{}
}

func NewServerWithDefaults() (*Server, error) {
	cfg := config.DefaultConfig()
	return NewServer(&cfg)
}

func NewServer(cfg *config.Config) (*Server, error) {
	router := router.NewRouter()

	router.Use(
		middleware.GzipMiddleware(),
	)

	server := Server{
		Port:                      cfg.Port,
		fileDir:                   cfg.FileDir,
		router:                    router,
		shutDownSignal:            make(chan struct{}),
		connectionsChan:           make(chan net.Conn, 100), // Using buffered chan so if new connections queue up, Accept() continues accepting
		numberOfConnectionsWorker: 10,
	}

	router.Handle("/files", handler.NewFileHandler(server.fileDir))
	router.Handle("/echo", handler.NewEchoHandler())
	router.Handle("/user-agent", handler.NewUserAgentHandler())
	router.Handle("/health", handler.NewHealthHandler(&server.startTime))

	return &server, nil
}

func (s *Server) ShutDown() {
	// Signal ShutDown to worker and main routine
	close(s.shutDownSignal)

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
		fmt.Println("Graceful shutdown timeout - some connections may be force-closed")
	}

}

func (s *Server) gracefulShutdownRoutine() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c
	fmt.Printf("Shutting down signal gracefully shutting down...\n")

	s.ShutDown()
}

func (s *Server) Start() error {
	listener, err := net.Listen("tcp", "0.0.0.0:"+s.Port)
	if err != nil {
		fmt.Println("Failed to bind to port ", s.Port)
		return err
	}
	s.listener = listener
	fmt.Println("Server listening on :", s.Port)

	go s.gracefulShutdownRoutine()

	for i := range s.numberOfConnectionsWorker {
		go s.startConnectionHandler(i)
	}

	s.startTime = time.Now()

	// Listen for new connections
	for {
		conn, err := listener.Accept()
		if err != nil {
			select {
			case <-s.shutDownSignal:
				fmt.Println("Shutdown signal received, listener probably closed.")
				return nil
			default:
				fmt.Println("Error accepting connection: ", err.Error())
				continue
			}
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
}

func (s *Server) startConnectionHandler(workerID int) {
	for {
		select {
		case conn := <-s.connectionsChan:
			fmt.Printf("Worker %d handling connection\n", workerID)
			s.handleConnection(conn)
		case <-s.shutDownSignal:
			fmt.Printf("Worker %d shutting down\n", workerID)
			return
		}
	}
}

func (s *Server) Stop() {
	s.listener.Close()
}

func (s *Server) handleConnection(conn net.Conn) {
	s.connectionWaitGroup.Add(1)
	atomic.AddInt64(&s.openConnections, 1)
	defer func() {
		conn.Close()
		atomic.AddInt64(&s.openConnections, -1)
		s.connectionWaitGroup.Done()
	}()

	conn.SetReadDeadline(time.Now().Add(30 * time.Second))
	for {
		request, err := http.ParseRequest(conn)
		if err != nil {
			if err == io.EOF {
				return // Client closed the connection
			}
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				return // Timeout
			}
			fmt.Println("Error parsing request:", err)
			return
		}

		// Clear read deadline during processing
		conn.SetDeadline(time.Time{})

		response := http.NewResponse(request)
		s.router.ServeHTTP(request, response)
		response.SendToClient(request)

		if request.Headers["Connection"] == "close" {
			break
		}

		// Set keep-alive timeout for next request
		conn.SetReadDeadline(time.Now().Add(time.Second * 60))
	}
}

func (s *Server) GetOpenConnections() int {
	return int(atomic.LoadInt64(&s.openConnections))
}
