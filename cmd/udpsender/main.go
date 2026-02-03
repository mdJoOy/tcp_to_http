package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
)

func main() {
	addr, err := net.ResolveUDPAddr("udp", "localhost:8080")
	if err != nil {
		log.Fatal("error resolving address", err)
	}

	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		log.Fatal("error dialing UDP:", err)
	}

	defer conn.Close()

	fmt.Println("UDP sender started. Send messages to localhost:8080")

	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("> ")

		input, err := reader.ReadString('\n')
		if err != nil {
			log.Println("error reading input:", err)
			continue
		}

		_, err = conn.Write([]byte(input))
		if err != nil {
			log.Println("error sending udp packet:", err)
		}
	}
}
