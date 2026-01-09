package server

import (
	"log"
	"net"
	"strconv"
)

type PublicListener struct {
	router *Router
}

func NewPublicListener(router *Router) *PublicListener {
	return &PublicListener{router: router}
}

func (p *PublicListener) Listen(port int) {
	ln, err := net.Listen("tcp", ":"+strconv.Itoa(port))
	if err != nil {
		log.Println("failed to bind public port", port, err)
		return
	}

	log.Println("public listener active on", port)

	for {
		conn, err := ln.Accept()
		if err != nil {
			continue
		}
		go p.handleConn(conn)
	}
}
