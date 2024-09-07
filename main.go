package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strings"
	"sync"
	"time"
)

var (
	clients     = make(map[net.Conn]string)
	mutex       sync.Mutex
	chatHistory []string // This will store the history of all chat messages
)

func main() {
	//addr := "localhost:8888"
	l, err := net.Listen("tcp", ":8091")
	if err != nil {
		log.Println("Couldn't listen to network", err)
		return
	}
	defer l.Close()

	fmt.Println("Server started on", "8091")

	for {
		conn, err := l.Accept()
		if err != nil {
			log.Println("Error while accepting connection:", err)
			continue
		}
		go handleConnect(conn)
	}
}

func handleConnect(conn net.Conn) {
	defer conn.Close()

	// Ask for the user's name
	conn.Write([]byte("[ENTER YOUR NAME]: "))
	reader := bufio.NewReader(conn)
	name, err := reader.ReadString('\n')
	if err != nil {
		log.Println("Error reading name:", err)
		return
	}
	name = strings.TrimSpace(name)

	notifyJoinLeave(name, "joined")
	// Check if the name is already taken
	sendChatHistory(conn)
	mutex.Lock()
	if checkName(name) {
		mutex.Unlock()
		conn.Write([]byte("Username is already taken. Please try again.\n"))
		return
	}

	clients[conn] = name
	mutex.Unlock()
	// Notify everyone that a new user has joined
	conn.Write([]byte(fmt.Sprintf("Welcome to the chat, %s!\n", name)))
	fmt.Println(conn.RemoteAddr())
	for {
		msg, err := reader.ReadString('\n')
		if err != nil {
			log.Println("Error reading message:", err)
			break
		}

		msg = strings.TrimSpace(msg)
		if msg == "" {
			continue
		}
		broadcastMessage(name, msg)
		fmt.Println("server")
	}

	// When the user disconnects
	mutex.Lock()
	delete(clients, conn)
	mutex.Unlock()
	notifyJoinLeave(name, "left")
}

func broadcastMessage(senderName, message string) {
	currentTime := time.Now().Format("2006-01-02 15:04:05")
	formattedMessage := fmt.Sprintf("[%s][%s]: %s", currentTime, senderName, message)

	mutex.Lock()

	chatHistory = append(chatHistory, formattedMessage)

	defer mutex.Unlock()

	for client := range clients {
		_, err := fmt.Fprintln(client, formattedMessage)
		if err != nil {
			log.Printf("Error sending message to client: %v", err)
		}
		// fmt.Fprintf(client, "[%s][%s]: ", currentTime,senderName)
	}
	fmt.Println(formattedMessage) // Also log to the server console
}

func notifyJoinLeave(name, action string) {
	// currentTime := time.Now().Format("2006-01-02 15:04:05")
	message := fmt.Sprintf("%s: has %s our chat...", name, action)
	formattedMessage := fmt.Sprintf("%s", message)
	mutex.Lock()
	chatHistory = append(chatHistory, formattedMessage)
	defer mutex.Unlock()

	for client := range clients {
		fmt.Fprintln(client, formattedMessage)
	}
	fmt.Println(formattedMessage) // Also log to the server console
}

func checkName(user string) bool {
	for _, name := range clients {
		if strings.ToLower(name) == strings.ToLower(user) {
			return true
		}
	}
	return false
}

func sendChatHistory(conn net.Conn) {
	mutex.Lock()
	defer mutex.Unlock()
	for _, message := range chatHistory {
		_, err := fmt.Fprintln(conn, message)
		if err != nil {
			log.Printf("Error sending chat history to client: %v", err)
			return
		}
	}
}
