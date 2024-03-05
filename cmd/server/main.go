package main

import (
	"encoding/binary"
	"fmt"
	"log"
	"net"
	"time"
)

func handleTCP(id uint64, conn net.Conn) {
	b := make([]byte, 1024)
	defer conn.Close()

	// send client its id
	binary.NativeEndian.PutUint64(b, id)
	conn.Write(b)

	for {
		conn.SetReadDeadline(time.Now().Add(5 * time.Second))
		n, err := conn.Read(b)

		// terminate connection on non-timeout error (eg. EOF)
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				continue
			} else {
				break
			}
		}

		received := string(b[:n-1]) // omit newline at the end
		fmt.Printf("Client %d: %s\n", id, received)
	}
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

		go handleTCP(id, conn)
	}
}
