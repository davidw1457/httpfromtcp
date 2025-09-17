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

	var line string

	for {
		buf := make([]byte, 8)
		_, err = file.Read(buf)
		if errors.Is(err, io.EOF) {
			break
		} else if err != nil {
			log.Fatalln(err)
		}

		split := strings.Split(string(buf), "\n")
		line += split[0]
		for len(split) > 1 {
			fmt.Printf("read: %s\n", line)
			split = split[1:]
			line = split[0]
		}
	}
}
