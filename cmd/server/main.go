package main

import (
	"encoding/binary"
	"fmt"
	"log"
	"net"
	"time"
)

var connections = make([]net.Conn, 0, 64)

func handleTCP(id uint64, conn net.Conn) {
	b := make([]byte, 1024)
	defer conn.Close()

	// send client its id
	binary.NativeEndian.PutUint64(b, id)
	conn.Write(b)

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

func main() {

	listener, err := net.Listen("tcp", "localhost:9001")
	defer listener.Close()

	if err != nil {
		log.Fatalln(err)
	}

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
