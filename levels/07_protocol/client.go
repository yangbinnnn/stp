package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
	"time"
)

type Client struct {
	addr   string
	rwConn *bufio.ReadWriter
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
	_, err := c.rwConn.WriteString("ping->\n")
	if err != nil {
		fmt.Println("error:", err.Error())
		c.notify <- err
		return
	}
	c.rwConn.Flush()
	pong, err := c.rwConn.ReadString('\n')
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
				c.Connect()
			}
		case <-time.After(3 * time.Second):
			time.Sleep(1 * time.Second)
		}
	}
}

func (c *Client) auth() error {
	// send name
	_, err := c.rwConn.WriteString("SSHR:01:AUTH:你好世界\n")
	if err != nil {
		fmt.Println("write auth error:", err.Error())
		return err
	}
	c.rwConn.Flush()
	reply, err := c.rwConn.ReadString('\n')
	if err != nil {
		fmt.Println("auth reply error:", err.Error())
		return err
	}
	reply = strings.TrimSpace(reply)
	if reply != "OK" {
		fmt.Println("auth fail, error", reply)
		return errors.New("auth fail")
	}
	return nil
}

func (c *Client) Connect() {
	saddr, err := net.ResolveTCPAddr("tcp", c.addr)
	checkError(err)
	conn, err := net.DialTCP("tcp", nil, saddr)
	checkError(err)
	fmt.Println("connected to", c.addr)
	c.rwConn = bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn))
	if err := c.auth(); err != nil {
		os.Exit(1)
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
