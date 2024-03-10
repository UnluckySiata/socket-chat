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

const asciiArt = "\n" +
	"   .----.\n" +
	"   |C>_ |\n" +
	" __|____|__\n" +
	"|  ______--|\n" +
	"`-/.::::.\\-'a\n" +
	" `--------'\n"

var (
	open       = true
	ip         = net.ParseIP("127.0.0.1")
	serverPort = 9001
)

func handleIncoming(conn net.Conn) {
	b := make([]byte, 4096)
	for open {
		conn.SetReadDeadline(time.Now().Add(200 * time.Millisecond))
		n, err := conn.Read(b)

		if err != nil {
			if netErr, ok := err.(net.Error); !(ok && netErr.Timeout()) {
				fmt.Println("\nConnection shut down by server")
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
	defer connTCP.Close()

	if err != nil {
		log.Fatalln(err)
	}

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
	defer connUDP.Close()

	if err != nil {
		log.Fatalln(err)
	}

	// get client id from server
	b := make([]byte, 64)
	connTCP.SetReadDeadline(time.Now().Add(5 * time.Second))
	_, err = connTCP.Read(b)

	if err != nil {
		log.Fatalln(err)
	}

	buffer := bytes.NewReader(b)
	var id uint64
	binary.Read(buffer, binary.NativeEndian, &id)
	fmt.Println("Hello Client ", id)

	// asynchronously read incomming messages from connections
	go handleIncoming(connTCP)
	go handleIncoming(connUDP)

	reader := bufio.NewReader(os.Stdin)
	for open {
		read, err := reader.ReadBytes('\n')

		if err == io.EOF {
			fmt.Println("\nexiting...")
			connTCP.Close()
			break
		}

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
