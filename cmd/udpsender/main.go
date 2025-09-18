package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
)

func main() {
	u, err := net.ResolveUDPAddr("udp", "localhost:42069")
	if err != nil {
		log.Fatalln(err)
	}

	conn, err := net.DialUDP("udp", nil, u)
	if err != nil {
		log.Fatalln(err)
	}

	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print(">")
		input, err := reader.ReadString(byte('\n'))
		if err != nil {
			log.Println(err)
		}

		_, err = conn.Write([]byte(input))
		if err != nil {
			log.Println(err)
		}
	}
}
