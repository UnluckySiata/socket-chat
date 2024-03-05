package main

import (
	"fmt"
	"log"
	"net"
	"time"
)

func handleTCP(conn net.Conn)  {
    b := make([]byte, 1024)
    defer conn.Close()

    for {
        conn.SetReadDeadline(time.Now().Add(time.Second))
        n, err := conn.Read(b)

        if err != nil {
            if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
                continue
            } else {
                break
            }
        }

        fmt.Printf("received %s\n", string(b[:n]))
    }
}

func main() {

    listener, err := net.Listen("tcp", "localhost:9001")
    defer listener.Close()

	if err != nil {
		log.Fatalln(err)
	}

    for {
        conn, err := listener.Accept()
        if err != nil {
            log.Fatalln(err)
        }

        go handleTCP(conn)
    }
}
