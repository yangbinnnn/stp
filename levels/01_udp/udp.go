package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
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
	echo = make([]byte, 512)
	rn, addr, err := conn.ReadFromUDP(echo)
	if err != io.EOF {
		checkError(err)
	}
	fmt.Printf("read %d byte\n", rn)
	us.reply(conn, addr, echo[:rn])
}

func (us UdpServer) reply(conn *net.UDPConn, addr *net.UDPAddr, msg []byte) {
	wn, err := conn.WriteToUDP(msg, addr)
	checkError(err)
	fmt.Printf("reply %d byte\n", wn)
}

type UdpClient struct {
	ServerAddr string
	conn       *net.UDPConn
}

func (uc *UdpClient) Connect() {
	udpAddr, err := net.ResolveUDPAddr("udp", uc.ServerAddr)
	checkError(err)

	uc.conn, err = net.DialUDP("udp", nil, udpAddr)
	checkError(err)
	fmt.Println("connect", uc.conn.RemoteAddr().String())
}

func (uc *UdpClient) Send(text string) {
	wn, err := uc.conn.Write([]byte(text))
	checkError(err)
	fmt.Printf("write %d byte\n", wn)
}

func (uc *UdpClient) Recive() {
	result := make([]byte, 512)
	// vs ioutil.ReadAll
	rn, err := bufio.NewReader(uc.conn).Read(result)
	checkError(err)

	fmt.Printf("recive %d byte, result: %s\n", rn, result)
	os.Exit(0)
}

func checkError(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}

func main() {
	serverAddr := flag.String("serverAddr", "127.0.0.1:8000", "server addr")
	role := flag.String("role", "", "roler [srv|cli]")
	msg := flag.String("msg", "hello udp", "text msg")
	flag.Parse()

	if *role == "srv" {
		us := &UdpServer{
			LisentAddr: *serverAddr,
		}
		fmt.Println("srv lisent on", us.LisentAddr)
		us.Start()
	} else if *role == "cli" {
		uc := &UdpClient{
			ServerAddr: *serverAddr,
		}
		uc.Connect()
		uc.Send(*msg)
		uc.Recive()
	} else {
		flag.Usage()
	}

}
