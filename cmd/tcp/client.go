package main

import (
	"io"
	"log"
	"net"
	"os"
)

func main() {
	conn, err := net.Dial("tcp", ":2020")
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	done := make(chan struct{})
	go func() {
		_, _ = io.Copy(os.Stdout, conn) // 将接收到的内容 copy 到标准输出
		done <- struct{}{}
	}()

	mustCopy(conn, os.Stdin) // 将标准输入的内容 copy 到 conn
	<-done
}

func mustCopy(dst io.Writer, src io.Reader) {
	if _, err := io.Copy(dst, src); err != nil {
		log.Fatal(err)
	}
}
