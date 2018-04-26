package main

import (
	"bufio"
	"flag"
	"fmt"
	"net"
	"os"
	"time"
)

type Client struct {
	addr string
	conn *net.TCPConn
}

func (c *Client) Ping() {
	c.checkConn()

	bufReader := bufio.NewReader(c.conn)
	bufWriter := bufio.NewWriter(c.conn)
	bufConn := bufio.NewReadWriter(bufReader, bufWriter)
	_, err := bufConn.WriteString("ping->\n")
	checkError(err)
	bufConn.Flush()

	pong, err := bufConn.ReadString('\n')
	checkError(err)
	fmt.Println(pong)
}

func (c *Client) checkConn() {
	if c.conn == nil {
		saddr, err := net.ResolveTCPAddr("tcp", c.addr)
		checkError(err)
		c.conn, err = net.DialTCP("tcp", nil, saddr)
		checkError(err)
	}
}

func checkError(err error) {
	if err != nil {
		fmt.Println("error:", err.Error())
		os.Exit(1)
	}
}

func main() {
	saddr := flag.String("saddr", "127.0.0.1:8000", "server addr")
	flag.Parse()

	c := Client{
		addr: *saddr,
	}
	for i := 0; i < 5; i++ {
		c.Ping()
		time.Sleep(2 * time.Second)
	}
	fmt.Println("sleep 5s")
	time.Sleep(5 * time.Second)

}
