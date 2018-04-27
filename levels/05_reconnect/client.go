package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"net"
	"time"
)

type Client struct {
	addr   string
	conn   *net.TCPConn
	notify chan error
}

func NewClient(addr string) *Client {
	c := Client{
		addr:   addr,
		notify: make(chan error),
	}
	c.Connect()
	go c.keepConnect()
	return &c
}

func (c *Client) Ping() {
	bufReader := bufio.NewReader(c.conn)
	bufWriter := bufio.NewWriter(c.conn)
	bufConn := bufio.NewReadWriter(bufReader, bufWriter)
	_, err := bufConn.WriteString("ping->\n")
	if err != nil {
		fmt.Println("error:", err.Error())
		c.notify <- err
		return
	}
	bufConn.Flush()

	pong, err := bufConn.ReadString('\n')
	if err != nil {
		fmt.Println("error:", err.Error())
		c.notify <- err
		return
	}
	fmt.Println(pong)
}

func (c *Client) keepConnect() {
	for {
		select {
		case err := <-c.notify:
			if err == io.EOF {
				c.conn.Close()
				c.Connect()
			}
		case <-time.After(3 * time.Second):
			time.Sleep(1 * time.Second)
		}
	}
}

func (c *Client) Connect() {
	saddr, err := net.ResolveTCPAddr("tcp", c.addr)
	checkError(err)
	c.conn, err = net.DialTCP("tcp", nil, saddr)
	checkError(err)
	fmt.Println("connected to", c.addr)
}

func checkError(err error) {
	if err != nil {
		fmt.Println("error:", err.Error())

	}
}

func main() {
	saddr := flag.String("saddr", "127.0.0.1:8000", "server addr")
	flag.Parse()

	c := NewClient(*saddr)
	c.Ping()
	fmt.Println("sleep 5s, wait timeout")
	time.Sleep(5 * time.Second)
	// close
	c.Ping()

	// close
	// reconect
	for {
		c.Ping()
		time.Sleep(1 * time.Second)
	}
}
