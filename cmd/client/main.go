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
	"unsafe"
)

const (
	serverPort    = 9001
	multicastPort = 9002
	bufSize       = 4096
	idLen         = 8

	asciiArt = "\n" +
		"   .----.\n" +
		"   |C>_ |\n" +
		" __|____|__\n" +
		"|  ______--|\n" +
		"`-/.::::.\\-'a\n" +
		" `--------'\n"
)

var (
	open        = true
	IP          = net.ParseIP("127.0.0.1")
	multicastIP = net.ParseIP("224.0.0.91")
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

func handleIncomingMulticast(conn *net.UDPConn, ownAddr *net.UDPAddr, ownID uint64) {
	b := make([]byte, bufSize)

	var senderID uint64

	for open {
		conn.SetReadDeadline(time.Now().Add(200 * time.Millisecond))
		n, _, err := conn.ReadFromUDP(b)

		if err != nil {
			// check if already handled closing
			if !open {
				return
			}

			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				continue
			} else {
				log.Println("\nMulticast connection shut down")
				break
			}
		}

		buffer := bytes.NewReader(b)
		binary.Read(buffer, binary.NativeEndian, &senderID)

		if n > 0 && ownID != senderID {
			fmt.Println(string(b[idLen:n]))
		}
	}
}

func main() {
	localTCP := net.TCPAddr{
		IP: IP,
	}
	serverTCP := net.TCPAddr{
		IP:   IP,
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
		IP:   IP,
		Port: int(localPort),
	}

	serverUDP := net.UDPAddr{
		IP:   IP,
		Port: serverPort,
	}

	connUDP, err := net.DialUDP("udp", &localUDP, &serverUDP)

	if err != nil {
		log.Printf("Failed to reach server udp socket: %v\nExiting..", err)
		return
	}
	defer connUDP.Close()

	multicastAddr := net.UDPAddr{
		IP:   multicastIP,
		Port: multicastPort,
	}

	multicastListener, err := net.ListenMulticastUDP("udp", nil, &multicastAddr)
	if err != nil {
		log.Printf("Failed to listen to multicast: %v\nExiting..", err)
		return
	}
	defer multicastListener.Close()

	multicastSender, err := net.DialUDP("udp4", nil, &multicastAddr)
	if err != nil {
		log.Printf("Failed to reach multicast udp socket: %v\nExiting..", err)
		return
	}
	defer multicastSender.Close()

	// get client id from server
	b := make([]byte, bufSize)
	connTCP.SetReadDeadline(time.Now().Add(5 * time.Second))
	_, err = connTCP.Read(b)

	if err != nil {
		log.Printf("Failed to receive client id from server: %v\nExiting..", err)
		return
	}

	buffer := bytes.NewReader(b)
	var id uint64
	binary.Read(buffer, binary.NativeEndian, &id)
	fmt.Printf("Client %d connected on address %v on port %v\n", id, IP, localPort)

	// asynchronously read incomming messages from connections
	go handleIncoming(connTCP)
	go handleIncoming(connUDP)
	go handleIncomingMulticast(multicastListener, &multicastAddr, id)

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
		case "M":
			message := fmt.Sprintf("Client %d (multicast): %s", id, asciiArt)

			idBytes := (*[idLen]byte)(unsafe.Pointer(&id))
			for i := 0; i < idLen; i++ {
				b[i] = idBytes[i]
			}
			for i := 0; i < len(message); i++ {
				b[i+idLen] = message[i]
			}

			multicastSender.Write(b[:idLen+len(message)])
		default:
			if len(read) > 1 {
				connTCP.Write(read)
			}
		}

	}
}
