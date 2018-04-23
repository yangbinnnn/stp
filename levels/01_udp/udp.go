package main

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"os"
)

type UdpServer struct {
	LisentAddr string
}

func (us UdpServer) Start() {
	udpAddr, err := net.ResolveUDPAddr("udp", us.LisentAddr)
	checkError(err)

	conn, err := net.ListenUDP("udp", udpAddr)
	checkError(err)

	for {
		us.handlerUdpClient(conn)
	}
}

func (us UdpServer) handlerUdpClient(conn *net.UDPConn) {
	var echo []byte
	_, addr, err := conn.ReadFromUDP(echo)
	if err != io.EOF {
		checkError(err)
	}
	conn.WriteToUDP(echo, addr)
}

type UdpClient struct {
	ServerAddr string
	conn       *net.UDPConn
}

func (uc *UdpClient) Connet() {
	udpAddr, err := net.ResolveUDPAddr("udp", uc.ServerAddr)
	checkError(err)

	uc.conn, err = net.DialUDP("udp", nil, udpAddr)
	checkError(err)
}

func (uc *UdpClient) Send(text string) {
	_, err := uc.conn.Write([]byte(text))
	checkError(err)
}

func (uc *UdpClient) Recive() {
	result, err := ioutil.ReadAll(uc.conn)
	checkError(err)

	fmt.Printf("result: %s", result)
	os.Exit(0)
}

func checkError(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}

func main() {
	serverAddr := flag.String("serverAddr", ":8000", "server addr")
	lisentAddr := flag.String("lisentAddr", ":8000", "lisent addr")
	role := flag.String("role", "", "roler")
	msg := flag.String("msg", "hello udp", "text msg")
	flag.Parse()

	if *role == "srv" {
		us := &UdpServer{
			LisentAddr: *lisentAddr,
		}
		us.Start()
	} else if *role == "cli" {
		uc := &UdpClient{
			ServerAddr: *serverAddr,
		}
		uc.Connet()
		uc.Send(*msg)
		uc.Recive()
	} else {
		flag.Usage()
	}

}
