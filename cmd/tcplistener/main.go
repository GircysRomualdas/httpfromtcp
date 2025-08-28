package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"strings"
)

func main() {
	port := ":42069"
	listener, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatal("Error listening:", err)
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatal("Error accepting connection:", err)
		}
		fmt.Println("Connection accepted")

		lines := getLinesChannel(conn)
		for line := range lines {
			fmt.Println(line)
		}
	}
}

func getLinesChannel(f io.ReadCloser) <-chan string {
	lines := make(chan string, 1)

	go func() {
		defer f.Close()
		defer close(lines)

		currentLineContents := ""
		for {
			buffer := make([]byte, 8)
			n, err := f.Read(buffer)
			if err != nil {
				if currentLineContents != "" {
					lines <- currentLineContents
				}
				if errors.Is(err, io.EOF) {
					break
				}
				log.Printf("error: %s\n", err.Error())
				return
			}

			str := string(buffer[:n])
			parts := strings.Split(str, "\n")

			for i := 0; i < len(parts)-1; i++ {
				lines <- currentLineContents + parts[i]
				currentLineContents = ""
			}

			currentLineContents += parts[len(parts)-1]
		}
	}()

	return lines
}
