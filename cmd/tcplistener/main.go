package main

import (
	"fmt"
	"log"
	"net"

	"github.com/GircysRomualdas/httpfromtcp/internal/request"
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

		req, err := request.RequestFromReader(conn)
		if err != nil {
			log.Fatal("Error reading request:", err)
		}
		fmt.Println("Request line:")
		fmt.Printf("- Method: %s\n", req.RequestLine.Method)
		fmt.Printf("- Target: %s\n", req.RequestLine.RequestTarget)
		fmt.Printf("- Version: %s\n", req.RequestLine.HttpVersion)
	}
}
