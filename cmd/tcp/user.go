package main

import (
	"strconv"
	"sync/atomic"
	"time"
)

type User struct {
	ID             int
	Addr           string
	EnterAt        time.Time
	MessageChannel chan string
}

var userId uint32 = 1

func GenUserID() int {
	return int(atomic.AddUint32(&userId, 1))
}

func (user *User) String() string {
	return strconv.Itoa(user.ID)
}

type Message struct {
	UserId  int
	Content string
}

func NewMessage(userId int, msg string) *Message {
	return &Message{
		UserId:  userId,
		Content: "user:`" + strconv.Itoa(userId) + "` " + msg,
	}
}
