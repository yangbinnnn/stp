package stp

import (
	"fmt"
	"io"
	"log"
	"net"

	"golang.org/x/crypto/ssh"
)

type Endpoint struct {
	Host string
	Port string
}

func (endpoint *Endpoint) String() string {
	return fmt.Sprintf("%s:%s", endpoint.Host, endpoint.Port)
}

type SSHtunnel struct {
	Local  *Endpoint
	Server *Endpoint
	Remote *Endpoint

	Config   *ssh.ClientConfig
	StopConn chan bool
}

func (tunnel *SSHtunnel) Start() error {
	// Connect to SSH remote server using serverEndpoint
	serverConn, err := ssh.Dial("tcp", tunnel.Server.String(), tunnel.Config)
	if err != nil {
		log.Println(fmt.Printf("Dial INTO remote server error: %s", err))
		return err
	}

	// Listen on remote server port
	listener, err := serverConn.Listen("tcp", tunnel.Remote.String())
	if err != nil {
		log.Println(fmt.Printf("Listen open port ON remote server error: %s", err))
		return err
	}
	defer listener.Close()

	newConn := make(chan net.Conn)
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				log.Println("accept ", err)
				return
			}
			newConn <- conn
		}
	}()

	// handle incoming connections on reverse forwarded tunnel
	for {
		select {
		case remote := <-newConn:
			// Open a (local) connection to localEndpoint whose content will be forwarded so serverEndpoint
			local, err := net.Dial("tcp", tunnel.Local.String())
			if err != nil {
				log.Fatalln(fmt.Printf("Dial INTO local service error: %s", err))
			}
			go handleClient(remote, local)
		case <-tunnel.StopConn:
			// stop tunnel
			listener.Close()
			tunnel.StopConn <- true
			return nil
		}
	}
}

func (tunnel *SSHtunnel) Stop() {
	tunnel.StopConn <- true
	log.Println("send the stop signal")
	stoped := <-tunnel.StopConn
	if stoped {
		log.Println("tunnel stoped")
	}
}

// From https://sosedoff.com/2015/05/25/ssh-port-forwarding-with-go.html
// Handle local client connections and tunnel data to the remote server
// Will use io.Copy - http://golang.org/pkg/io/#Copy
func handleClient(client net.Conn, remote net.Conn) {
	defer client.Close()
	chDone := make(chan bool)

	// Start remote -> local data transfer
	go func() {
		_, err := io.Copy(client, remote)
		if err != nil && err != io.EOF {
			log.Println(fmt.Sprintf("error while copy remote->local: %s", err))
		}
		chDone <- true
	}()

	// Start local -> remote data transfer
	go func() {
		_, err := io.Copy(remote, client)
		if err != nil && err != io.EOF {
			log.Println(fmt.Sprintf("error while copy local->remote: %s", err))
		}
		chDone <- true
	}()

	<-chDone
}

func privateKeyAuthMethod(privateKey string) (ssh.AuthMethod, error) {
	key, err := ssh.ParsePrivateKey([]byte(privateKey))
	if err != nil {
		return nil, err
	}
	return ssh.PublicKeys(key), nil
}
