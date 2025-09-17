package main

import (
    "errors"
    "fmt"
    "io"
    "log"
    "os"
)

func main() {
    file, err := os.Open("messages.txt")
    if err != nil {
        log.Fatalln(err)
    }

    buf := make([]byte, 8)
    for {
        _, err = file.Read(buf)
        if errors.Is(err, io.EOF) {
            break
        } else if err != nil {
            log.Fatalln(err)
        }
        fmt.Printf("read: %s\n", string(buf))
    }
}
