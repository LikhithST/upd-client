package main

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"net"
	"os"
	"time"
)

type DataChannelMessage struct {
	FrameID                      int64  `json:"frameID"`
	MessageSentTimeLocalMachine1 int64  `json:"messageSentTime_LocalMachine1,omitempty"`
	MessageSentTimeVM1           int64  `json:"messageSentTime_VM1,omitempty"`
	MessageSentTimeVM2           int64  `json:"messageSentTime_VM2,omitempty"`
	MessageSentTimeLocalMachine2 int64  `json:"messageSentTime_LocalMachine2,omitempty"`
	LatencyEndToEnd              int64  `json:"latency_end_to_end,omitempty"`
	MessageSendRate              int64  `json:"message_send_rate,omitempty"`
	Payload                      []byte `json:"payload"`
}

func main() {
	// Define and parse the remote address flag
	addr := flag.String("addr", "127.0.0.1:12345", "remote UDP address")
	flag.Parse()

	// Resolve the remote address
	raddr, err := net.ResolveUDPAddr("udp", *addr)
	if err != nil {
		fmt.Println("Failed to resolve address:", err)
		os.Exit(1)
	}

	// Open a UDP connection
	conn, err := net.DialUDP("udp", nil, raddr)
	if err != nil {
		fmt.Println("Failed to dial UDP:", err)
		os.Exit(1)
	}
	defer conn.Close()

	payloadBytes := make([]byte, 100) // 120000 bytes = 120 KB payload
	frameID := 1
	lastMessageTime := time.Now().UnixMilli()
	// Create a ticker with a 33ms interval
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()
	// Example object to send

	for range ticker.C {
		_, err := rand.Read(payloadBytes)
		if err != nil {
			fmt.Println("Error generating random payload:", err)
			continue
		}

		messageSentTime := time.Now().UnixMilli()
		send_rate := messageSentTime - lastMessageTime
		// Create the DataChannelMessage object

		// Encode to base64 to make it JSON-safe (you can also use hex if needed)
		payload := base64.StdEncoding.EncodeToString(payloadBytes)
		message := DataChannelMessage{
			FrameID:                      int64(frameID),
			MessageSentTimeLocalMachine1: messageSentTime,
			MessageSendRate:              int64(send_rate),
			Payload:                      []byte(payload),
		}
		// Marshal the object to JSON
		data, err := json.Marshal(message)
		if err != nil {
			fmt.Println("Error marshaling message:", err)
			continue
		}

		// Send the JSON payload
		fmt.Println("Sending:", len(data), message.MessageSendRate, "ms", message.FrameID)
		_, err = conn.Write(data)
		if err != nil {
			fmt.Println("Error sending message:", err)
			return
		}

		lastMessageTime = time.Now().UnixMilli()
		// if frameID == 1 {
		// 	fmt.Println("Sent 100 messages, waiting for 1 second...")
		// 	time.Sleep(100 * time.Second)
		// }
		frameID++
	}

	// Optional: receive a response
	recvBuf := make([]byte, 1024)
	n, _, err := conn.ReadFromUDP(recvBuf)
	if err == nil {
		fmt.Println("Received:", string(recvBuf[:n]))
	}
}
