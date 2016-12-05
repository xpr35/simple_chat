package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
)

//RecvMessage recieves messages from clients
//As parameter is uses connection
func RecvMessage(conn net.Conn) {
	reader := bufio.NewReader(conn)
	for {
		str, err := reader.ReadString('\n')
		if err != nil {
			log.Fatal("Connection to server is down")
		}
		str = str[0 : len(str)-1]
		fmt.Println(str)
	}
}

func main() {
	if len(os.Args) < 3 {
		log.Println("Please, use server(e.g. localhost:6578) and username as first and second params")
		return
	}
	server_addr := os.Args[1]
	username := os.Args[2]
	conn, err := net.Dial("tcp", server_addr)
	if err != nil {
		log.Fatal("Can't connect to ", server_addr, err)
	}
	fmt.Println("Connected to " + os.Args[1])
	fmt.Fprintf(conn, username+"\n")

	writer := bufio.NewWriter(conn)

	go RecvMessage(conn)
	for {
		rd := bufio.NewReader(os.Stdin)
		message, err := rd.ReadString('\n')
		if err != nil {
			log.Println("Can't read message from stdin")
		}
		writer.WriteString(message)
		writer.Flush()
	}
}
