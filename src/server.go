package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
)

type Client struct {
	conn         net.Conn
	username     string
	status       string //online or offline
	send_chan    chan Message
	recieve_chan chan Message
	writer       *bufio.Writer
	reader       *bufio.Reader
}

type Message struct {
	username string
	text     string
}

// NewClient creates a new chat client and returns it.
// It gets first message from client and uses it as username.
func NewClient(conn net.Conn) Client {
	reader := bufio.NewReader(conn)
	usr, _ := reader.ReadString('\n')
	usr = usr[0 : len(usr)-1]

	log.Println("New user in chat: " + usr)
	tmp := Client{
		conn:         conn,
		username:     usr,
		status:       "online",
		send_chan:    make(chan Message),
		recieve_chan: make(chan Message),
		writer:       bufio.NewWriter(conn),
		reader:       reader}
	return tmp
}

//RecvMessage recives messages from the clients
//As parameters it takes two channels:
//First channel is for all messages in chat
//Second channel is for running listing function
func (c *Client) RecvMessage(common_messages_chan chan Message, l chan Client) {
	for {
		str, err := c.reader.ReadString('\n')
		if err != nil {
			log.Println("User " + c.username + " left the chat")
			c.status = "offline"
			c.conn.Close()
			break
		}
		str = str[0 : len(str)-1]

		if strings.Index(str, "bye") > -1 {
			log.Println("User " + c.username + " left the chat")
			c.status = "offline"
			c.conn.Close()
			common_messages_chan <- Message{username: c.username,
				text: str}
			break
		} else if strings.Index(str, "list") > -1 {
			l <- *c
		} else {
			//Delete comment below to print all messages into server console
			//log.Println("["+c.username+"]"+str)
			common_messages_chan <- Message{username: c.username,
				text: str}
		}
	}
}

//SendMessage gets messages from recieve channel of each client and sends them to them
func (c *Client) SendMessage() {
	for {
		message := <-c.send_chan
		fmt.Fprintf(c.conn, message.text+"\n")
	}
}

type ClientSlice []Client //because here is no other way to use slice

//Broadcast sends messages to each client except sender
//As params it takes two channels:
//First channel is for all messages in chat
//Second channel is for saving message history
func (c *ClientSlice) Broadcast(common_messages_chan chan Message, history chan Message) {
	for {
		message := <-common_messages_chan
		history <- message
		for _, usr := range *c {
			if message.username != usr.username {
				usr.send_chan <- Message{username: message.username,
					text: "[" + message.username + "] " + message.text}
			}
		}
	}
}

//List sends list of users to specified client
//It takes as first parameter chan of client whom to send
func (c *ClientSlice) List(l chan Client) {
	for {
		user_to_list := <-l
		user_to_list.send_chan <- Message{username: "server", text: "Users in chat:"}
		for _, usr := range *c {
			user_to_list.send_chan <- Message{username: "server", text: usr.username + " | " + usr.status}
		}
	}
}

/* ---------------------- History ------------------------------ */
/*
//CreateHistory cares about chat history
//As parameters it takes:
//First is channel of new messages, that comes here from broadcast function
//Second is size of a history
func (history *Queue) CreateHistory(new_message chan Message, size_of_a_history int) {
	for {
		tmp_msg := <-new_message
		if history.count == size_of_a_history {
			history.Pop()
		}
		history.Push(&Node{tmp_msg})
	}
}

//History function sends last N messages to specified client
//As first parameter it takes user, whom to send
func (history *Queue) History(c Client) {
	messages_list := history.List()
	for _, tmp_msg := range messages_list {
		c.send_chan <- Message{username: tmp_msg.username,
			text: "[" + tmp_msg.username + "]" + tmp_msg.text}
	}
}
*/

type History_t struct {
	income  chan Message
	outcome chan Client
	history Queue
	size    int
}

func NewHistory(size_of_a_history int) History_t {
	tmp_queue := NewQueue(size_of_a_history)
	tmp_history := History_t{
		income:  make(chan Message),
		outcome: make(chan Client),
		history: *tmp_queue,
		size:    size_of_a_history}
	return tmp_history
}

func (h *History_t) History() {
	for {
		select {
		case tmp_msg := <-h.income:
			if h.history.count >= h.size {
				h.history.Pop()
			}
			h.history.Push(&Node{tmp_msg})
		case tmp_client := <-h.outcome:
			messages_list := h.history.List()
			for _, tmp_msg := range messages_list {
				tmp_client.send_chan <- Message{username: tmp_msg.username,
					text: "[" + tmp_msg.username + "]" + tmp_msg.text}
			}
		}
	}
}

//////////////////////////////// MAIN ////////////////////////////////
func main() {
	if len(os.Args) < 2 {
		log.Println("Please, use port number as parameter")
		return
	}
	port := os.Args[1]
	ln, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatal("Can't run server on port "+port+" because of ", err)
	}
	log.Println("Server works!")

	var client_list ClientSlice

	ten_messages := NewHistory(10) // instead of //ten_messages := NewQueue(10)

	var common_messages_chan chan Message = make(chan Message)
	var send_list_of_clients_chan chan Client = make(chan Client)
	// no need //var history chan Message = make(chan Message)

	go ten_messages.History() //instead of//go ten_messages.CreateHistory(history, 10)
	go client_list.Broadcast(common_messages_chan, ten_messages.income)
	go client_list.List(send_list_of_clients_chan)

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Println("Can't get connection")
		}

		client := NewClient(conn)
		client_list = append(client_list, client)

		ten_messages.outcome <- client //instead of //go ten_messages.History(client)

		go client.RecvMessage(common_messages_chan, send_list_of_clients_chan)
		go client.SendMessage()
	}
}
