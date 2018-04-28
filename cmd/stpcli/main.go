package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/takama/daemon"
	"github.com/yangbinnnn/stp"
)

const (
	// VERSION 0.0.1
	// 0.0.2 clean offline tunnel
	// 0.0.3 service supported
	VERSION = "0.0.3"

	name        = "stpcli"
	description = "stpcli quickly create ssh tunnel"
)

var service, _ = daemon.New(name, description)

func main() {
	var (
		showVersion bool
		name        string // client name
		serverUrl   string
		localPort   string
		authKey     string
		install     bool
		uninstall   bool
		background  bool
		stop        bool
		start       bool
		status      bool
		logfile     string
	)
	flag.BoolVar(&showVersion, "v", false, "show version")
	flag.StringVar(&name, "n", "", "client name")
	flag.StringVar(&serverUrl, "h", "ws://127.0.0.1:10000", "stp server connect url")
	flag.StringVar(&localPort, "p", "22", "stp local forward port")
	flag.StringVar(&authKey, "key", "tunnelkey", "stp auth key")
	flag.BoolVar(&install, "install", false, "install service")
	flag.BoolVar(&uninstall, "uninstall", false, "uninstall service")
	flag.BoolVar(&background, "d", false, "run as daemon service")
	flag.BoolVar(&stop, "stop", false, "stop stpcli service")
	flag.BoolVar(&start, "start", false, "start stpcli service")
	flag.BoolVar(&status, "status", false, "status stpcli service")
	flag.StringVar(&logfile, "logfile", "/var/log/stpcli.log", "stpcli service log")
	flag.Parse()

	if len(os.Args) == 1 {
		flag.Usage()
		return
	}

	if showVersion {
		fmt.Println(VERSION)
		return
	}

	file, err := os.OpenFile(logfile, os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	mw := io.MultiWriter(os.Stdout, file)
	log.SetOutput(mw)

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

	if name == "" || serverUrl == "" {
		log.Println("need name or serverUrl")
		return
	}

	args := []string{"-h", serverUrl, "-key", authKey, "-p", localPort, "-n", name}
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
	// retry forever
	cli := stp.NewSTPClient(authKey, serverUrl, localPort, name)
	for {
		err := cli.Login()
		if err != nil {
			log.Println(err.Error())
			time.Sleep(3 * time.Second)
			log.Printf("retry login")
			continue
		} else {
			cli.Daemon()
		}
	}
}
