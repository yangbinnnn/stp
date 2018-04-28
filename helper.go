package stp

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	osuser "os/user"
	"runtime"
	"strings"
)

func AddAuthorizedKey(publicKey, user string) error {
	authPath := "/root/.ssh/authorized_keys"
	if user == "" {
		u, err := osuser.Current()
		if err != nil {
			return err
		}
		user = u.Username
	}
	if user != "root" {
		if runtime.GOOS == "darwin" {
			authPath = fmt.Sprintf("/Users/%s/.ssh/authorized_keys", user)
		} else {
			authPath = fmt.Sprintf("/home/%s/.ssh/authorized_keys", user)
		}
	}
	var file *os.File
	var err error
	if !FileExist(authPath) {
		file, err = os.Create(authPath)
		if err != nil {
			return err
		}
	}
	if file == nil {
		file, err = os.OpenFile(authPath, os.O_APPEND|os.O_RDWR, os.ModeAppend)
		if err != nil {
			return err
		}
	}
	defer file.Close()

	notExist := true
	rw := bufio.NewReadWriter(bufio.NewReader(file), bufio.NewWriter(file))
	for {
		line, err := rw.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		if line == publicKey {
			notExist = false
			break
		}
	}

	if notExist {
		_, err := rw.WriteString(strings.TrimSpace(publicKey) + "\n")
		if err != nil {
			return err
		}
		log.Println("add authorized public key success", authPath)
		log.Println("public key", publicKey)
	} else {
		log.Println("authorized public key exist")
	}
	return rw.Flush()
}

func FileExist(path string) bool {
	_, err := os.Stat(path)
	return err == nil || os.IsExist(err)
}

func AppendLine(path, line string, unique bool) error {
	file, err := os.OpenFile(path, os.O_APPEND|os.O_RDWR, os.ModeAppend)
	if err != nil {
		return err
	}
	notExist := true
	rw := bufio.NewReadWriter(bufio.NewReader(file), bufio.NewWriter(file))
	if !unique {
		_, err := rw.WriteString(strings.TrimSpace(line) + "\n")
		if err != nil {
			return err
		}
		return nil
	}
	for {
		_line, err := rw.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		if _line == line {
			notExist = false
			break
		}
	}

	if notExist {
		_, err := rw.WriteString(strings.TrimSpace(line) + "\n")
		if err != nil {
			return err
		}
	}
	return rw.Flush()
}

func AddStartUp(cmd string) error {
	rcs := []string{"/etc/rc.local", "/etc/rc.d/after.local"}
	for _, rc := range rcs {
		if FileExist(rc) {
			return AppendLine(rc, cmd, true)
		}
	}
	return errors.New("system not supported")
}
