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
    b := make([]byte, 1024)
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
		fmt.Printf("Client %d> ", id)
		text, err := reader.ReadString('\n')

		if err == io.EOF {
			fmt.Println("\nexiting...")
			break
		}

        if len(text) > 1 {
            fmt.Fprint(conn, text)
        }
	}
}
