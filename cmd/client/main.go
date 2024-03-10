package main

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net"
	"net/netip"
	"os"
	"time"
)

const (
	serverPort = 9001
	bufSize    = 4096

	asciiArt = "\n" +
		"   .----.\n" +
		"   |C>_ |\n" +
		" __|____|__\n" +
		"|  ______--|\n" +
		"`-/.::::.\\-'a\n" +
		" `--------'\n"
)

var (
	open = true
	ip   = net.ParseIP("127.0.0.1")
)

func handleIncoming(conn net.Conn) {
	b := make([]byte, bufSize)
	for open {
		conn.SetReadDeadline(time.Now().Add(200 * time.Millisecond))
		n, err := conn.Read(b)

		if err != nil {
			// check if already handled closing
			if !open {
				return
			}

			if netErr, ok := err.(net.Error); !(ok && netErr.Timeout()) {
				log.Println("\nConnection shut down by server")
				open = false
				break
			}
		}

		// print new messages if there are any
		if n > 0 {
			fmt.Print(string(b[:n]))
		}
	}
}

func main() {
	localTCP := net.TCPAddr{
		IP: ip,
	}
	serverTCP := net.TCPAddr{
		IP:   ip,
		Port: serverPort,
	}

	connTCP, err := net.DialTCP("tcp", &localTCP, &serverTCP)

	if err != nil {
		log.Printf("Failed to establish tcp connection to server: %v\nExiting..", err)
		return
	}
	defer connTCP.Close()

	// determine local tcp port to bind the same to udp socket
	localAddrPort, _ := netip.ParseAddrPort(connTCP.LocalAddr().String())
	localPort := localAddrPort.Port()

	localUDP := net.UDPAddr{
		IP:   ip,
		Port: int(localPort),
	}

	serverUDP := net.UDPAddr{
		IP:   ip,
		Port: serverPort,
	}

	connUDP, err := net.DialUDP("udp", &localUDP, &serverUDP)

	if err != nil {
		log.Printf("Failed to reach server udp socket: %v\nExiting..", err)
		return
	}
	defer connUDP.Close()

	// get client id from server
	b := make([]byte, 64)
	connTCP.SetReadDeadline(time.Now().Add(5 * time.Second))
	_, err = connTCP.Read(b)

	if err != nil {
		log.Printf("Failed to receive client id from server: %v\nExiting..", err)
		return
	}

	buffer := bytes.NewReader(b)
	var id uint64
	binary.Read(buffer, binary.NativeEndian, &id)
	fmt.Printf("Client %d connected on address %v on port %v\n", id, ip, localPort)

	// asynchronously read incomming messages from connections
	go handleIncoming(connTCP)
	go handleIncoming(connUDP)

	reader := bufio.NewReader(os.Stdin)
	for open {
		read, err := reader.ReadBytes('\n')

		if err == io.EOF {
			fmt.Println("\nexiting...")
			open = false
			break
		}

		// match on read bytes, omiting newline
		switch string(read[:len(read)-1]) {
		case "U":
			connUDP.Write([]byte(asciiArt))
		default:
			if len(read) > 1 {
				connTCP.Write(read)
			}
		}

	}
}
