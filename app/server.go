package main

import (
	"fmt"
	"net"
	"os"
	"redis/app/parser"
	"sync"
)

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

	go commandHandlerLoop(server)

	if err := server.Start(); err != nil {
		fmt.Println("Failed to start server", "error", err)
		os.Exit(1)
	}
}

func commandHandlerLoop(server *Server) {
	for {
		msg := <-server.msgChan
		fmt.Println("Received message", "message", string(msg.payload), "from", msg.from)

		server.RLock()
		conn := server.peerMap[msg.from]
		server.RUnlock()

		responses := handlePayload(msg.payload)

		for _, response := range responses {
			if _, err := conn.Write(response); err != nil {
				fmt.Println("Failed to write to connection", "error", err)
			}
		}

	}
}

func handlePayload(payload []byte) [][]byte {
	payloadStr := string(payload)

	var responses [][]byte

	p := parser.NewParser(payloadStr)
	cmds, err := p.Parse()
	if err != nil {
		responses = append(responses, []byte("-ERR "+err.Error()+"\r\n"))
		return responses
	}

	if len(cmds) == 0 {
		responses = append(responses, []byte("-ERR empty command\r\n"))
		return responses
	}

	for _, cmd := range cmds {
		if cmd == nil {
			responses = append(responses, []byte("-ERR invalid command\r\n"))
			continue
		}

		response := cmd.Respond()
		responses = append(responses, []byte(response))
	}

	return responses
}
