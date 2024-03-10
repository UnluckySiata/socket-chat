package main

import (
	"encoding/binary"
	"fmt"
	"log"
	"net"
	"net/netip"
	"time"
)

const (
	serverPort = 9001
	bufSize    = 4096
)

var (
	ip          = net.ParseIP("127.0.0.1")
	connections = make([]net.Conn, 0, 64)
	addrPorts   = make(map[netip.AddrPort]uint64)
)

func handleTCP(id uint64, conn net.Conn) {
	b := make([]byte, bufSize)
	defer conn.Close()

	// send client its id
	binary.NativeEndian.PutUint64(b, id)
	conn.Write(b)

	// get connections AddrPort to later identify clients
	remoteAddrPort, _ := netip.ParseAddrPort(conn.RemoteAddr().String())
	addrPorts[remoteAddrPort] = id
	fmt.Printf("Client %d established connection from %v\n", id, remoteAddrPort)

	for {
		conn.SetReadDeadline(time.Now().Add(time.Second))
		n, err := conn.Read(b)

		// terminate connection on non-timeout error (eg. EOF)
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				continue
			} else {
				break
			}
		}

		fullMsg := fmt.Sprintf("Client %d: %s", id, string(b[:n]))
		log.Print(fullMsg)

		for _, c := range connections {
			if c != conn && c != nil {
				c.Write([]byte(fullMsg))
			}
		}
	}
	// remove client from available connections
	connections[id] = nil
	delete(addrPorts, remoteAddrPort)
	log.Printf("Client %d disconnected\n", id)
}

func handleUDP() {
	addr := net.UDPAddr{
		Port: serverPort,
		IP:   ip,
	}
	b := make([]byte, bufSize)

	conn, err := net.ListenUDP("udp", &addr)

	if err != nil {
		log.Printf("Failed to open udp socket: %v\n", err)
		return
	}
	defer conn.Close()

	for {
		conn.SetReadDeadline(time.Now().Add(200 * time.Millisecond))
		n, sender, err := conn.ReadFromUDP(b)

		if err == nil {
			// get client id from incomming AddrPort
			senderAddrPort := sender.AddrPort()
			id := addrPorts[senderAddrPort]

			received := string(b[:n])
			fullMsg := fmt.Sprintf("Client %d: %s", id, received)
			log.Printf("Client %d via UDP: %s", id, received)

			for receiverAddrPort, _ := range addrPorts {
				if senderAddrPort != receiverAddrPort {
					udpAddr := net.UDPAddrFromAddrPort(receiverAddrPort)
					conn.WriteToUDP([]byte(fullMsg), udpAddr)
				}
			}
		}
	}
}

func main() {
	addr := net.TCPAddr{
		Port: serverPort,
		IP:   ip,
	}

	listener, err := net.ListenTCP("tcp", &addr)

	if err != nil {
		log.Printf("Failed to open tcp socket: %v\nExiting...", err)
		return
	}
	defer listener.Close()

	fmt.Printf("Server listening on address %v on port %v\n", ip, serverPort)

	go handleUDP()

	var id uint64
	for id = 0; ; id++ {
		conn, err := listener.Accept()

		if err != nil {
			log.Printf("Failed to accept tcp connection: %v\n", err)
		} else {
			connections = append(connections, conn)
			go handleTCP(id, conn)
		}
	}
}
