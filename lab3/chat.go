// Demonstration of channels with a chat application
// Copyright © 2016 Alan A. A. Donovan & Brian W. Kernighan.
// License: https://creativecommons.org/licenses/by-nc-sa/4.0/

// Chat is a server that lets clients chat with each other.

package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
)

type client struct {
	out_channel chan<- string // sends outgoing message as strings
	username    string        // stores client username
}

// Constructor function for client
func newClient(username string, out_channel chan<- string) *client { // Returns a pointer to client struct
	cli := client{username: username, out_channel: out_channel}
	return &cli
}

var (
	entering = make(chan client) // signals a new client entering the system
	leaving  = make(chan client) // signalw when a client leavs the system
	messages = make(chan string) // all incoming client messages
)

// set up a tcp server
func main() {
	listener, err := net.Listen("tcp", "localhost:8000")
	if err != nil {
		log.Fatal(err)
	}

	go broadcaster() // run broadcaster concurently
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Print(err)
			continue
		}
		go handleConn(conn)
	}
}

// manages the communications
func broadcaster() {
	clients := make(map[client]bool) // all connected clients
	for {
		select {
		case msg := <-messages:
			// Broadcast incoming message to all
			// clients' outgoing message channels.
			for cli := range clients {
				cli.out_channel <- msg
			}

		case cli := <-entering:

			// Prints currently in chat if there are any users
			if len(clients) != 0 {
				cli.out_channel <- " " // formatting line
				cli.out_channel <- "Currently in chat: "
				for user := range clients {
					cli.out_channel <- " 	" + user.username
				}
			}
			cli.out_channel <- " " // formatting line

			clients[cli] = true

		case cli := <-leaving:
			delete(clients, cli)
			close(cli.out_channel)
		}
	}
}

func handleConn(conn net.Conn) {
	ch := make(chan string) // outgoing client messages
	go clientWriter(conn, ch)

	//who := conn.RemoteAddr().String()
	//ch <- "You are " + who
	ch <- "Choose username: "

	usernameCh := make(chan string)
	input := bufio.NewScanner(conn)

	go func() {
		if input.Scan() {
			usernameCh <- input.Text()
		} else {
			// Handle the error or disconnection scenario
			usernameCh <- "DefaultUsername"
		}
	}()

	// Wait for the username input
	username := <-usernameCh
	client := newClient(username, ch)

	messages <- client.username + " has arrived"
	messages <- " " // formatting line

	entering <- *client

	for input.Scan() {
		messages <- client.username + ": " + input.Text()
	}
	// NOTE: ignoring potential errors from input.Err()

	leaving <- *client
	messages <- client.username + " has left"

	//maybe some clean up

	conn.Close()
}

func clientWriter(conn net.Conn, ch <-chan string) {
	for msg := range ch {
		fmt.Fprintln(conn, msg) // NOTE: ignoring network errors
	}
}
