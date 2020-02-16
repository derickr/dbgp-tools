package server

import (
	"fmt"
	"io"
	"net"
	"sync"
	"time"
)

type Handler interface {
	Handle(conn net.Conn) error
}

type Server struct {
	address    *net.TCPAddr
	group      *sync.WaitGroup
	stop       bool
	serverType string
}

func NewServer(serverType string, address *net.TCPAddr, group *sync.WaitGroup) *Server {
	return &Server{
		address,
		group,
		false,
		serverType,
	}
}

func (server *Server) Listen(handler Handler) {
	server.group.Add(1)
	defer server.group.Done()

	listener, err := net.ListenTCP("tcp", server.address)
	if err != nil {
		panic(err)
	}

	defer server.closeConnection(listener)

	fmt.Printf("Started %s server on %s\n", server.serverType, server.address)

	for {
		if server.stop {
			break
		}

		_ = listener.SetDeadline(time.Now().Add(time.Second * 2))
		conn, err := listener.Accept()
		if err != nil {
			if opErr, ok := err.(*net.OpError); ok && opErr.Timeout() {
				continue
			}
			fmt.Print(err)
			continue
		}
		go server.handleConnection(conn, handler)
	}

	fmt.Printf("Shutdown %s server\n", server.serverType)
}

func (server *Server) handleConnection(conn net.Conn, handler Handler) {
	defer server.closeConnection(conn)
	server.group.Add(1)
	defer server.group.Done()
	fmt.Printf("- Start new %s connection from %s\n", server.serverType, conn.RemoteAddr())
	err := handler.Handle(conn)

	if err != nil {
		fmt.Printf("  - Handler response error: %s\n", err)
	}

	fmt.Printf("- Closing %s connection from %s\n", server.serverType, conn.RemoteAddr())
}

func (server *Server) closeConnection(closer io.Closer) {
	err := closer.Close()
	if err != nil {
		panic(err)
	}
}

func (server *Server) Stop() {
	server.stop = true
}
