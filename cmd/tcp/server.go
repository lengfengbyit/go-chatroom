package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"time"
)

func main() {
	listener, err := net.Listen("tcp", ":2020")
	if err != nil {
		panic(err)
	}

	go broadcaster()

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println(err)
			continue
		}

		go handleConn(conn)

	}
}

var messageChannel = make(chan *Message)
var enteringChannel = make(chan *User)
var leavingChannel = make(chan *User)

func handleConn(conn net.Conn) {
	defer conn.Close()

	// 1. 新用户进来， 构建该用户的实例
	user := &User{
		ID:             GenUserID(),
		Addr:           conn.RemoteAddr().String(),
		EnterAt:        time.Now(),
		MessageChannel: make(chan string, 8),
	}

	// 2. 由于当前是一个新的goroutine中进行读写操作的，所以需要开一个goroutine
	// 用于写操作。读写 goroutine 之间可以通过 channel 进行通信
	go sendMessage(conn, user.MessageChannel)

	// 3. 给当前用户发送欢迎信息， 向所有用户告知新用户到来
	user.MessageChannel <- "Welcome, " + user.String()
	messageChannel <- NewMessage(user.ID, "has enter")

	// 4. 记录到全局用户列表中， 避免用锁
	enteringChannel <- user

	// 4.1 提出超时用户
	var userActive = make(chan struct{})
	go func() {
		d := 10 * time.Second
		timer := time.NewTicker(d)
		for {
			select {
			case <-timer.C:
				conn.Write([]byte("timeout..."))
				conn.Close()
				return
			case <-userActive:
				timer.Reset(d)
			}
		}
	}()

	// 5. 循环读取用户输入
	input := bufio.NewScanner(conn)
	for input.Scan() {
		messageChannel <- NewMessage(user.ID, input.Text())

		// 表示用户活跃
		userActive <- struct{}{}
	}

	if err := input.Err(); err != nil {
		log.Println("read err:", err)
	}

	// 6. 用户离开
	leavingChannel <- user

	messageChannel <- NewMessage(user.ID, "has left")

}

// sendMessage 发送消息
func sendMessage(conn io.Writer, ch <-chan string) {
	for msg := range ch {
		_, _ = fmt.Fprintln(conn, msg)
	}
}

// broadcaster 用于记录聊天室用户，并进行消息广播
// 1. 新用户进来， 2. 用户普通消息 3. 用户离开
func broadcaster() {
	users := make(map[*User]struct{})

	for {
		select {
		case user := <-enteringChannel:
			// 新用户进入
			users[user] = struct{}{}
		case user := <-leavingChannel:
			if user == nil {
				continue
			}
			// 用户离开
			delete(users, user)
			// 避免goroutine泄露
			close(user.MessageChannel)
			user = nil
		case msg := <-messageChannel:
			// 给所有在线用户发送消息
			for user := range users {
				if user.ID == msg.UserId {
					continue
				}
				user.MessageChannel <- msg.Content
			}
		}
	}
}
