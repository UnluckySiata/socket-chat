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

func main() {

	conn, err := net.Dial("tcp", "localhost:9001")
	defer conn.Close()

	if err != nil {
		log.Fatalln(err)
	}

	// get client id from server
	b := make([]byte, 4096)
	conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	_, err = conn.Read(b)

	if err != nil {
		log.Fatalln(err)
	}

	buffer := bytes.NewReader(b)
	var id uint64
	binary.Read(buffer, binary.NativeEndian, &id)

	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Printf("Enter a command: [c]hat\n> ")
		read, err := reader.ReadBytes('\n')

        if err == io.EOF {
            fmt.Println("\nexiting...")
            break
        }

        conn.SetReadDeadline(time.Now().Add(200 * time.Millisecond))
        n, _ := conn.Read(b)

        // print new messages if there are any
        if n > 0 {
            fmt.Print(string(b[:n]))
        }

        cmd := read[0]
		switch cmd {
		case 'c':
			fmt.Printf("Client %d> ", id)
			text, _ := reader.ReadString('\n')

			if len(text) > 1 {
                conn.Write([]byte(text))
			}
		default:
		}
	}
}
