package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
)

func getLinesChannel(f io.ReadCloser) <-chan string {
	out := make(chan string, 1)
	go func() {
		defer f.Close()
		defer close(out)

		str := ""
		for {
			data := make([]byte, 8)
			n, err := f.Read(data)
			if err != nil {
				break
			}
			data = data[:n]

			if i := bytes.IndexByte(data, '\n'); i != -1 {
				str += string(data[:i])
				data = data[i+1:]
				out <- str
				str = ""
			}
			str += string(data)
		}
	}()
	return out
}

func main() {
	f, err := os.Open("messages.txt")
	if err != nil {
		log.Fatal("error", err)
	}

	lines := getLinesChannel(f)
	for line := range lines {
		fmt.Printf("%s\n", line)
	}
}
