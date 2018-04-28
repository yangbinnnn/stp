package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"net"
	"os"
	"strings"
	"time"
)

type Server struct {
	addr    string
	timeout int64
}

func (s *Server) auth(rwConn *bufio.ReadWriter) error {
	authString, err := rwConn.ReadString('\n')
	if err != nil {
		fmt.Println("parse auth string fail")
		return err
	}
	// SSHR:01:AUTH:TOKEN
	authString = strings.TrimSpace(authString)
	items := strings.Split(authString, ":")
	if items[0] != "SSHR" {
		fmt.Println("unknow protocol")
		rwConn.WriteString("NO\n")
		rwConn.Flush()
		return errors.New("unknow protocol")
	}
	if items[3] != "你好世界" {
		fmt.Println("invaild token")
		rwConn.WriteString("NO\n")
		rwConn.Flush()
		return errors.New("invaild token")
	}
	fmt.Println("token:", items[3])
	_, err = rwConn.WriteString("OK\n")
	err = rwConn.Flush()
	return err
}

func (s *Server) handler(conn *net.TCPConn) {
	fmt.Println("handling new connection...")

	defer func() {
		fmt.Println("closing connection...")
		conn.Close()
	}()

	rw := bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn))
	err := s.auth(rw)
	if err != nil {
		fmt.Println("client auth error:", err.Error())
		return
	}

	for {
		// SetDeadline
		now := time.Now()
		conn.SetDeadline(now.Add(time.Duration(s.timeout) * time.Second))
		fmt.Printf("Now: %s, +%ds\n", now.Format(time.RFC822), s.timeout)
		// Block read
		ping, err := rw.ReadString('\n')
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println(ping)
		// echo
		_, err = rw.WriteString("<-pong\n")
		if err != nil {
			fmt.Println(err)
			return
		}
		rw.Flush()
	}

}

func (s *Server) Start() {
	laddr, err := net.ResolveTCPAddr("tcp", s.addr)
	checkError(err)

	listener, err := net.ListenTCP("tcp", laddr)
	checkError(err)

	fmt.Println("listen on", s.addr)

	for {
		conn, err := listener.AcceptTCP()
		if err != nil {
			fmt.Println("accept error:", err.Error())
			continue
		}
		go s.handler(conn)
	}
}

func checkError(err error) {
	if err != nil {
		fmt.Println("error:", err.Error())
		os.Exit(1)
	}
}

func main() {
	laddr := flag.String("laddr", "127.0.0.1:8000", "listen addr")
	timeout := flag.Int64("timeout", 3, "timeout")
	flag.Parse()
	s := Server{
		addr:    *laddr,
		timeout: *timeout,
	}
	s.Start()
}
