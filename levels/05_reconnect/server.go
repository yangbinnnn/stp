package main

import "bufio"
import "flag"
import "fmt"
import "net"
import "os"
import "time"

type Server struct {
	addr    string
	timeout int64
}

func (s *Server) handler(conn *net.TCPConn) {
	fmt.Println("handling new connection...")

	defer func() {
		fmt.Println("closing connection...")
		conn.Close()
	}()

	bufReader := bufio.NewReader(conn)
	bufWriter := bufio.NewWriter(conn)
	for {
		// SetDeadline
		now := time.Now()
		conn.SetDeadline(now.Add(time.Duration(s.timeout) * time.Second))
		fmt.Printf("Now: %s, +%ds\n", now.Format(time.RFC822), s.timeout)
		// Block read
		ping, err := bufReader.ReadString('\n')
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println(ping)
		// echo
		_, err = bufWriter.WriteString("<-pong\n")
		if err != nil {
			fmt.Println(err)
			return
		}
		bufWriter.Flush()
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
