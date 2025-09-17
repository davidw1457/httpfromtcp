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
	l, err := net.Listen("tcp", ":42069")
	if err != nil {
		log.Fatalln(err)
	}
	defer l.Close()

	for {
		conn, err := l.Accept()
		if err != nil {
			log.Fatalln(err)
		}

		fmt.Println("Connection accepted")

		c := getLinesChannel(conn)

		for s := range c {
			fmt.Printf("%s\n", s)
		}
		fmt.Println("Connection closed")
	}

}

func getLinesChannel(f io.ReadCloser) <-chan string {
	c := make(chan string)

	go func() {
		defer f.Close()
		defer close(c)
		var line string
		for {
			buf := make([]byte, 8)
			_, err := f.Read(buf)
			if errors.Is(err, io.EOF) {
				break
			} else if err != nil {
				log.Fatalln(err)
			}

			split := strings.Split(string(buf), "\n")
			line += split[0]
			for len(split) > 1 {
				c <- line
				split = split[1:]
				line = split[0]
			}
		}
	}()

	return c
}
