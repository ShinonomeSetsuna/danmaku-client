package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

type ReceiveMessageType struct {
	message string
	font    string
	size    int
	speed   int // from 1 to 10
	color   int // 0x66ccff
}

// 处理跨域请求
var upgrader = websocket.Upgrader{

	CheckOrigin: func(*http.Request) bool { // CROS相关
		return true
	},
}

// 储存client
var clients = make(map[*websocket.Conn]bool)

// 广播消息
var broadcast = make(chan string)

// 与 网页WebSocket 通信
func handleWebSocketForWeb(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	defer ws.Close()

	clients[ws] = true

	for {
		var msg string
		err := ws.ReadJSON(&msg)
		if err != nil {
			log.Printf("error: %v", err)
			delete(clients, ws)
			break
		}
		broadcast <- msg
	}
}

// 广播消息给所有 网页WebSocket 客户端
func handleMessagesSending() {
	for {
		msg := <-broadcast
		for client := range clients {
			err := client.WriteJSON(msg)
			if err != nil {
				log.Printf("error: %v", err)
				client.Close()
				delete(clients, client)
			}
		}
	}
}

// 处理静态网页文件
func handleStaticWebFiles() {
	fs := http.FileServer(http.Dir("./public"))
	http.Handle("/", fs)
}

func main() {
	// 处理 WebSocket 路由
	http.HandleFunc("/ws", handleWebSocketForWeb)

	go handleMessagesSending()

	handleStaticWebFiles()

	fmt.Println("Hello, world!")
	go func() {
		err := http.ListenAndServe(":3000", nil)
		if err != nil {
			log.Fatal(err)
		}
	}()
	for true {
		fmt.Println()
		var msg string
		n, err := fmt.Scanln(&msg)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(n)
		broadcast <- msg
	}
}
