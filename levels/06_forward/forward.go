package main

import (
	"fmt"
	"io"
	"net"
	"os"
)

func checkError(err error) {
	if err != nil {
		fmt.Println("error:", err.Error())
		os.Exit(1)
	}
}

func forward(conn net.Conn, remote net.Conn) {
	go func() {
		defer conn.Close()
		defer remote.Close()
		io.Copy(remote, conn)
	}()
	go func() {
		defer conn.Close()
		defer remote.Close()
		io.Copy(conn, remote)
	}()
}

func main() {
	listener, err := net.Listen("tcp", "127.0.0.1:8000")
	checkError(err)
	for {
		conn, err := listener.Accept()
		checkError(err)
		remoteConn, err := net.Dial("tcp", "61.155.138.173:11580")
		checkError(err)
		go forward(conn, remoteConn)

	}
}
