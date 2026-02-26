package main

import (
	"crypto/sha256"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/mdjOoy/tcptohttp/internal/headers"
	"github.com/mdjOoy/tcptohttp/internal/request"
	"github.com/mdjOoy/tcptohttp/internal/response"
	"github.com/mdjOoy/tcptohttp/internal/server"
)

const port = 42069

func toStr(sha []byte) string {
	out := ""
	for _, v := range sha {
		out += fmt.Sprintf("%02x", v)
	}
	return out
}
func respond400() []byte {
	return []byte(`<html>
  <head>
    <title>400 Bad Request</title>
  </head>
  <body>
    <h1>Bad Request</h1>
    <p>Your request honestly kinda sucked.</p>
  </body>
</html>`)
}
func respond500() []byte {
	return []byte(`<html>
  <head>
    <title>500 Internal Server Error</title>
  </head>
  <body>
    <h1>Internal Server Error</h1>
    <p>Okay, you know what? This one is on me.</p>
  </body>
</html>`)
}
func respondOk() []byte {
	return []byte(`<html>
  <head>
    <title>200 OK</title>
  </head>
  <body>
    <h1>Success!</h1>
    <p>Your request was an absolute banger.</p>
  </body>
</html>`)
}
func main() {
	s, err := server.Serve(port, func(w *response.Writer, req *request.Request) {
		h := response.GetDefaultHeaders(0)

		h.Replace("Content-Type", "text/html")

		statusCode := response.StatusOK
		body := respondOk()

		if req.RequestLine.RequestTarget == "/yourproblem" {
			statusCode = response.StatusBadRequest
			body = respond400()
		} else if req.RequestLine.RequestTarget == "/myproblem" {
			statusCode = response.StatusInternalServerError
			body = respond500()
		} else if strings.HasPrefix(req.RequestLine.RequestTarget, "/httpbin") {
			target := req.RequestLine.RequestTarget
			//sending req to the original server
			res, err := http.Get("https://httpbin.org/" + strings.TrimPrefix(target, "/httpbin/"))
			if err != nil {
				statusCode = response.StatusBadRequest
				body = respond400()
			}
			//fixing headers so that it supports chunk encoding
			h.Delete("Content-Length")
			h.Set("Transfer-Encoding", "chunked")
			h.Replace("Content-Type", "text/plain")
			h.Set("Trailer", "X-Content-SHA256")
			h.Set("Trailer", "X-Content-Length")
			//writing status line
			w.WriteStatusLine(statusCode)
			//writing headers
			w.WriteHeaders(h)
			//writing chunk body
			data := make([]byte, 1024)
			fullBody := []byte{}
			for {
				n, err := res.Body.Read(data[0:])
				if err != nil {
					break
				}
				fullBody = append(fullBody, data[:n]...)
				w.WriteBody([]byte(fmt.Sprintf("%x\r\n", n)))
				w.WriteBody(data[:n])
				w.WriteBody([]byte("\r\n"))
			}
			sha := sha256.Sum256(fullBody)
			w.WriteBody([]byte("0\r\n"))
			trailers := headers.NewHeaders()
			trailers.Set("X-Content-SHA256", toStr(sha[:]))
			trailers.Set("X-Content-Length", fmt.Sprintf("%d", len(fullBody)))
			trailers.Set("testing", "demo")
			w.WriteHeaders(trailers)
			return
		} else if req.RequestLine.RequestTarget == "/video" {
			f, _ := os.ReadFile("assets/vim.mp4")
			h.Replace("Content-Type", "video/mp4")
			h.Replace("Content-Length", fmt.Sprintf("%d", len(f)))
			w.WriteStatusLine(response.StatusOK)
			w.WriteHeaders(h)
			w.WriteBody(f)
			return
		}

		w.WriteStatusLine(statusCode)
		h.Replace("Content-Length", fmt.Sprintf("%d", len(body)))
		w.WriteHeaders(h)
		w.WriteBody(body)
	})
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
	defer s.Close()
	log.Println("Server started on port", port)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("Server gracefully stopped")
}
