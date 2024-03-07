package main

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"time"
)

var closed = false

func handleIncoming(conn net.Conn) {
    b := make([]byte, 4096)
    for {
		conn.SetReadDeadline(time.Now().Add(200 * time.Millisecond))
		n, err := conn.Read(b)

        if err != nil {
            if netErr, ok := err.(net.Error); !(ok && netErr.Timeout()) {
                fmt.Println("\nConnection shut down by server")
                closed = true
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
	conn, err := net.Dial("tcp", "localhost:9001")
	defer conn.Close()

	if err != nil {
		log.Fatalln(err)
	}

	// get client id from server
	b := make([]byte, 64)
	conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	_, err = conn.Read(b)

	if err != nil {
		log.Fatalln(err)
	}

	buffer := bytes.NewReader(b)
	var id uint64
	binary.Read(buffer, binary.NativeEndian, &id)
    fmt.Println("Hello Client ", id)

    go handleIncoming(conn)
	reader := bufio.NewReader(os.Stdin)
	for {
        if closed {
            break
        }
		read, err := reader.ReadBytes('\n')

		if err == io.EOF {
			fmt.Println("\nexiting...")
			break
		}

		if len(read) > 1 {
			conn.Write(read)
		}
	}
}
