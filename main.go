package main

import (
	"fmt"
	"io"
	"net/http"

	"golang.org/x/net/websocket"
)

type Server struct {
	Conns map[*websocket.Conn]bool
}

func NewServer() *Server {
	return &Server{
		Conns: make(map[*websocket.Conn]bool),
	}
}

func (s *Server) handleWs(ws *websocket.Conn) {
	defer func() {
		fmt.Println("client disconnected: ", ws.RemoteAddr())
		delete(s.Conns, ws)
		ws.Close()
	}()
	fmt.Println("new incoming connection from client:", ws.RemoteAddr())
	s.Conns[ws] = true

	s.readLoop(ws)
}

func (s *Server) readLoop(ws *websocket.Conn) {
	buf := make([]byte, 1024)
	for {
		n, err := ws.Read(buf)
		if err != nil {
			if err == io.EOF {
				break
			}

			fmt.Println("read error:", err)
			break
		}
		msg := buf[:n]
		fmt.Println(string(msg))

		go s.broadcast(msg)
	}
}

func (s *Server) broadcast(msg []byte) {
	for ws := range s.Conns {
		go func(ws *websocket.Conn) {
			if _, err := ws.Write(msg); err != nil {
				fmt.Println("Unable to write message")
			}
		}(ws)
	}
}

func main() {
	server := NewServer()

	http.Handle("/ws", websocket.Handler(server.handleWs))
	http.ListenAndServe(":3000", nil)
}
