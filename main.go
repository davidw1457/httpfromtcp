package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

func main() {
	file, err := os.Open("messages.txt")
	if err != nil {
		log.Fatalln(err)
	}

	c := getLinesChannel(file)

	for s := range c {
		fmt.Printf("read: %s\n", s)
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
