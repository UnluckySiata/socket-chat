package main

import (
	"encoding/binary"
	"fmt"
	"log"
	"net"
	"net/netip"
	"time"
)

var (
	port        = 9001
	ip          = net.ParseIP("127.0.0.1")
	connections = make([]net.Conn, 0, 64)
	addrPorts   = make(map[netip.AddrPort]uint64)
)

func handleTCP(id uint64, conn net.Conn) {
	b := make([]byte, 1024)
	defer conn.Close()

	// send client its id
	binary.NativeEndian.PutUint64(b, id)
	conn.Write(b)

	// get connections AddrPort to later identify clients
	remoteAddrPort, _ := netip.ParseAddrPort(conn.RemoteAddr().String())
	addrPorts[remoteAddrPort] = id

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
		fmt.Print(fullMsg)

		for _, c := range connections {
			if c != conn && c != nil {
				c.Write([]byte(fullMsg))
			}
		}
	}
	connections[id] = nil
}

func handleUDP() {
	addr := net.UDPAddr{
		Port: port,
		IP:   ip,
	}
	b := make([]byte, 4096)

	conn, err := net.ListenUDP("udp", &addr)
	defer conn.Close()

	if err != nil {
		log.Fatalln(err)
	}

	for {
		conn.SetReadDeadline(time.Now().Add(200 * time.Millisecond))
		n, sender, err := conn.ReadFromUDP(b)

		if err == nil {
			// get client id from incomming AddrPort
			senderAddrPort := sender.AddrPort()
			id := addrPorts[senderAddrPort]

			fullMsg := fmt.Sprintf("Client %d: %s", id, string(b[:n]))
			fmt.Println(fullMsg)

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
		Port: port,
		IP:   ip,
	}

	listener, err := net.ListenTCP("tcp", &addr)
	defer listener.Close()

	if err != nil {
		log.Fatalln(err)
	}

	go handleUDP()

	var id uint64
	for id = 0; ; id++ {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatalln(err)
		}
		connections = append(connections, conn)

		go handleTCP(id, conn)
	}
}
