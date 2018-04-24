package main

import (
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"time"
)

type TCPServer struct {
	serverAddr string
}

func (ts TCPServer) Start() {
	addr, err := net.ResolveTCPAddr("tcp", ts.serverAddr)
	checkError(err)
	lisenter, err := net.ListenTCP("tcp", addr)
	checkError(err)
	for {
		conn, err := lisenter.AcceptTCP()
		if err != nil {
			fmt.Println("conn error", err.Error())
			continue
		}
		fmt.Printf("client %s connected.\n", conn.RemoteAddr().String())
		go ts.handlerEcho(conn)
	}

}

func (ts TCPServer) handlerEcho(conn *net.TCPConn) {
	defer func() {
		conn.Close()
		fmt.Printf("client %s disconnect\n", conn.RemoteAddr().String())
	}()
	for {
		echo := make([]byte, 512)
		rn, err := conn.Read(echo)
		if err != nil {
			if err == io.EOF {
				break
			} else {
				checkError(err)
			}
		}
		fmt.Printf("read %d byte\n", rn)
		wn, err := conn.Write(echo[:rn])
		checkError(err)
		fmt.Printf("write %d byte\n", wn)
	}
}

type TCPClient struct {
	serverAddr string
	conn       *net.TCPConn
}

func (tc *TCPClient) Connect() {
	addr, err := net.ResolveTCPAddr("tcp", tc.serverAddr)
	checkError(err)
	tc.conn, err = net.DialTCP("tcp", nil, addr)
	checkError(err)
	fmt.Printf("connect to %s success.\n", tc.conn.RemoteAddr().String())
}

func (tc TCPClient) Send(text string) {
	wn, err := tc.conn.Write([]byte(text))
	checkError(err)
	fmt.Printf("write %d byte\n", wn)
}

func (tc TCPClient) Recive() {
	for {
		buf := make([]byte, 512)
		rn, err := tc.conn.Read(buf)
		if err != nil {
			if err == io.EOF {
				break
			}
		} else {
			checkError(err)
		}
		fmt.Printf("read %d byte, msg %s\n", rn, buf[:rn])
	}
}

func checkError(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func main() {
	serverAddr := flag.String("serverAddr", "127.0.0.1:8000", "server addr")
	role := flag.String("role", "", "roler [srv|cli]")
	flag.Parse()

	if *role == "srv" {
		srv := TCPServer{
			serverAddr: *serverAddr,
		}
		fmt.Println("lisent on", srv.serverAddr)
		srv.Start()
	} else if *role == "cli" {
		cli := TCPClient{
			serverAddr: *serverAddr,
		}
		cli.Connect()
		go cli.Recive()
		for n := 10; n > 0; n-- {
			cli.Send(time.Now().String())
			time.Sleep(1 * time.Second)
		}
	} else {
		flag.Usage()
	}
}
