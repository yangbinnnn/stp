package main

import (
	"fmt"
	"io"
	"net"
	"os"
)

type Proxy struct {
	frontendAddr string
	backendAddr  string
	frontendConn *net.TCPConn
	backendConn  *net.TCPConn
}

func (p *Proxy) Start() {
	// connect to backend
	baddr, err := net.ResolveTCPAddr("tcp", p.backendAddr)
	if err != nil {
		fmt.Printf("invalid addr %s, error %s", p.backendAddr, err.Error())
		os.Exit(1)
	}
	p.backendConn, err = net.DialTCP("tcp", nil, baddr)
	if err != nil {
		fmt.Printf("connect backend %s fail, error %s", p.backendAddr)
		os.Exit(1)
	}

	// listen on frontend
	faddr, err := net.ResolveTCPAddr("tcp", p.frontendAddr)
	if err != nil {
		fmt.Printf("invalid addr %s, error %s", p.backendAddr, err.Error())
		os.Exit(1)
	}
	listener, err := net.ListenTCP("tcp", faddr)
	if err != nil {
		fmt.Printf("listen on %s fail, error %s", p.frontendAddr, err.Error())
	}
	fmt.Printf("listen on %s", p.frontendAddr)
	conn, _ := listener.AcceptTCP()
	p.frontendConn = conn
	p.transport()
}

func (p Proxy) transport() {
	go func() {
		_, err := io.Copy(p.frontendConn, p.backendConn)
		fmt.Printf("[->]transport error %s\n", err.Error())
	}()
	_, err := io.Copy(p.backendConn, p.frontendConn)
	fmt.Printf("[<-]transport error %s\n", err.Error())
}

func main() {
	proxy := Proxy{
		frontendAddr: "127.0.0.1:1122",
		backendAddr:  "www.baidu.com:80",
	}
	proxy.Start()
}
