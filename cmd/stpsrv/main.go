package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"time"

	"github.com/olekukonko/tablewriter"
	"github.com/takama/daemon"
	"github.com/yangbinnnn/stp"
)

const (
	// VERSION 0.0.1
	// 0.0.2 macos support
	// 0.0.3 service supported
	VERSION = "0.0.3"

	name        = "stpsrv"
	description = "stpsrv quickly create ssh tunnel service"
)

var service, _ = daemon.New(name, description)

func listClients(serverAddr string) []*stp.Client {
	u := url.URL{Scheme: "http", Host: serverAddr, Path: "/showClient"}
	httpClient := http.Client{
		Timeout: 3 * time.Second,
	}
	resp, err := httpClient.Get(u.String())
	if err != nil {
		fmt.Println("http get error", err.Error())
		return nil
	}
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("http read resp error", err.Error())
		return nil
	}
	clients := []*stp.Client{}
	err = json.Unmarshal(data, &clients)
	if err != nil {
		fmt.Println("http unmarshal error", err.Error())
		return nil
	}
	return clients
}

func showClientCmd(serverAddr string) {
	clients := listClients(serverAddr)
	if clients == nil {
		fmt.Println("no client")
		return
	}
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"num", "name", "port", "addr", "online"})
	for i, client := range clients {
		table.Append([]string{strconv.Itoa(i), client.Name, client.Port, client.Addr, strconv.FormatBool(client.IsOnline)})
	}
	table.Render()
}

func connectClientCmd(num int, user string, serverAddr string) {
	clients := listClients(serverAddr)
	if clients == nil {
		fmt.Println("no client")
		return
	}
	client := clients[num]
	cmd := exec.Command("ssh", "-o", "ServerAliveInterval=30", "-o", "ServerAliveCountMax=3000", "-p", client.Port, fmt.Sprintf("%s@127.0.0.1", user))
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	err := cmd.Run()
	if err != nil {
		fmt.Println(err.Error())
	}
}

func loadSSHKey(path string) (private string, public string) {
	privateByte, err := ioutil.ReadFile(path)
	if err != nil {
		fmt.Println("load private key fail, error", err.Error())
		os.Exit(1)
	}
	// plus ".pub"
	publicByte, err := ioutil.ReadFile(path + ".pub")
	if err != nil {
		fmt.Println("load private key fail, error", err.Error())
		os.Exit(1)
	}
	private = string(privateByte)
	public = string(publicByte)
	return
}

func main() {
	var (
		showVersion bool
		showClient  bool
		connectNum  int
		connectUser string

		install    bool
		uninstall  bool
		background bool
		start      bool
		stop       bool
		status     bool
		logFile    string
		cfgFile    string
	)

	pwd, _ := os.Getwd()
	flag.StringVar(&cfgFile, "cfg", fmt.Sprintf("%s/cfg.json", pwd), "stpsrv service cfg file")

	// CLI
	flag.BoolVar(&showVersion, "v", false, "show version")
	flag.BoolVar(&showClient, "l", false, "list clients")
	flag.IntVar(&connectNum, "c", -1, "num connect to ssh client")
	flag.StringVar(&connectUser, "u", "root", "ssh connect user")
	// service
	flag.BoolVar(&install, "install", false, "install service")
	flag.BoolVar(&uninstall, "uninstall", false, "uninstall service")
	flag.BoolVar(&background, "d", false, "run as daemon service")
	flag.BoolVar(&stop, "stop", false, "stop stpsrv service")
	flag.BoolVar(&start, "start", false, "start stpsrv service")
	flag.BoolVar(&status, "status", false, "status stpsrv service")
	flag.StringVar(&logFile, "logfile", "/var/log/stpsrv.log", "stpsrv service log")
	flag.Parse()

	if len(os.Args) == 1 {
		flag.Usage()
		return
	}

	if showVersion {
		fmt.Println(VERSION)
		return
	}

	ParseConfig(cfgFile)

	if showClient {
		showClientCmd(Config().ListentAddr)
		return
	}

	if connectNum != -1 {
		connectClientCmd(connectNum, connectUser, Config().ListentAddr)
		return
	}

	file, err := os.OpenFile(logFile, os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	mw := io.MultiWriter(os.Stdout, file)
	log.SetOutput(mw)

	// service
	if uninstall {
		service.Stop()
		status, err := service.Remove()
		log.Println(status)
		if err != nil {
			log.Println(err.Error())
		}
		return
	}

	if stop {
		status, err := service.Stop()
		log.Println(status)
		if err != nil {
			log.Println(status, err.Error())
		}
		return
	}

	if start {
		status, err := service.Start()
		log.Println(status)
		if err != nil {
			log.Println(status, err.Error())
		}
		return
	}

	if status {
		status, err := service.Status()
		log.Println(status)
		if err != nil {
			log.Println(status, err.Error())
		}
		return
	}

	args := []string{"-cfg", cfgFile}
	if install {
		status, err := service.Install(args...)
		log.Println(status)
		if err != nil {
			log.Println(status, err.Error())
			return
		}
		return
	}

	if background {
		status, err := service.Install(args...)
		if err != nil && err != daemon.ErrAlreadyInstalled {
			log.Println(status, err.Error())
			return
		}
		status, err = service.Start()
		if err != nil {
			log.Println(status, err.Error())
			return
		}
		return
	}

	if Config().SSHRSAPath == "" {
		if Config().SSHUser == "root" {
			Config().SSHRSAPath = fmt.Sprintf("/root/.ssh/id_rsa")
		} else {
			if runtime.GOOS == "darwin" {
				Config().SSHRSAPath = fmt.Sprintf("/Users/%s/.ssh/id_rsa", Config().SSHUser)
			} else {
				Config().SSHRSAPath = fmt.Sprintf("/home/%s/.ssh/id_rsa", Config().SSHUser)
			}
		}
	}
	privateKey, publicKey := loadSSHKey(Config().SSHRSAPath)
	s := stp.NewSTPServer(Config().AuthKey, Config().ListentAddr, Config().SSHAddr, privateKey, publicKey, Config().SSHUser, Config().PortRange)
	s.Start()
}
