package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"sync"
)

var logger = log.Default()

type Message struct {
	from    string
	payload []byte
}

type Server struct {
	sync.RWMutex

	listenAddress string
	listener      net.Listener
	quitChan      chan struct{}
	msgChan       chan Message
	peerMap       map[string]net.Conn
}

func NewServer(listenAddress string) *Server {
	return &Server{
		listenAddress: listenAddress,
		quitChan:      make(chan struct{}),
		msgChan:       make(chan Message, 10),
		peerMap:       map[string]net.Conn{},
	}
}

func (s *Server) Start() error {
	ln, err := net.Listen("tcp", s.listenAddress)
	if err != nil {
		return err
	}
	defer ln.Close()

	s.listener = ln

	go s.acceptLoop()

	<-s.quitChan
	close(s.msgChan)

	return nil
}

func (s *Server) acceptLoop() {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			fmt.Println("Failed to accept connection", "error", err)
		}

		fmt.Println("Accepted connection", "remote_addr", conn.RemoteAddr())

		go s.readConn(conn)
	}
}

func (s *Server) readConn(conn net.Conn) {
	defer conn.Close()
	buf := make([]byte, 1024)

	s.Lock()
	s.peerMap[conn.RemoteAddr().String()] = conn
	s.Unlock()

	for {
		msgLen, err := conn.Read(buf)
		if err != nil {
			fmt.Println("Failed to read from connection", "error", err)
			return
		}
		msg := buf[:msgLen]

		s.msgChan <- Message{
			from:    conn.RemoteAddr().String(),
			payload: msg,
		}
	}
}

func main() {
	fmt.Println("Starting Redis", "version", "0.0.1")

	server := NewServer("0.0.0.0:6379")

	go func() {
		for {
			msg := <-server.msgChan
			fmt.Println("Received message", "message", string(msg.payload), "from", msg.from)

			server.RLock()
			conn := server.peerMap[msg.from]
			server.RUnlock()

			response := HandleRedisCommand(string(msg.payload))

			if _, err := conn.Write(response); err != nil {
				fmt.Println("Failed to write to connection", "error", err)
			}
		}
	}()

	if err := server.Start(); err != nil {
		fmt.Println("Failed to start server", "error", err)
		os.Exit(1)
	}
}

func HandleRedisCommand(message string) []byte {
	command := strings.TrimSuffix(message, "\r\n")
	switch command {
	case "PING":
		return []byte("+PONG\r\n")
	default:
		return []byte("ERR unknown command '" + command + "'")
	}
}
