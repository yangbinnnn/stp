package main

import (
	"bytes"
	"fmt"
	"io"
)

func main() {
	r, w := io.Pipe()
	go func() {
		fmt.Fprintln(w, "hello world from pipe!")
		w.Close()
	}()
	buf := new(bytes.Buffer)
	buf.ReadFrom(r)
	fmt.Println(buf)
}
