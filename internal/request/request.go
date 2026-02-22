package request

import (
	"bytes"
	"fmt"
	"io"

	"github.com/mdjOoy/tcptohttp/internal/headers"
)

type parseState string

const (
	StateInit    parseState = "init"
	StateHeaders parseState = "headers"
	StateDone    parseState = "done"
	StateError   parseState = "error"
)

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

type Request struct {
	RequestLine RequestLine
	Headers     *headers.Headers
	State       parseState
}

func newRequest() *Request {
	return &Request{
		State:   StateInit,
		Headers: headers.NewHeaders(),
	}
}

var ErrorMalformedRequestLine = fmt.Errorf("malform start line")
var ErrorUnsupportedHttpVersion = fmt.Errorf("unsupported http version")
var ErrorRequestInErrorState = fmt.Errorf("request in error state")
var SP = []byte("\r\n")

func parseRequestLine(b []byte) (*RequestLine, int, error) {
	idx := bytes.Index(b, SP)
	if idx == -1 {
		return nil, 0, nil
	}

	startLine := b[:idx]
	read := idx + len(SP)

	parts := bytes.Split(startLine, []byte(" "))
	if len(parts) != 3 {
		return nil, 0, ErrorMalformedRequestLine
	}

	// checking valid http version
	httpParts := bytes.Split(parts[2], []byte("/"))
	if len(httpParts) != 2 || string(httpParts[0]) != "HTTP" || string(httpParts[1]) != "1.1" {
		return nil, 0, ErrorUnsupportedHttpVersion
	}
	rl := RequestLine{
		Method:        string(parts[0]),
		RequestTarget: string(parts[1]),
		HttpVersion:   string(httpParts[1]),
	}
	return &rl, read, nil
}

func (r *Request) parse(data []byte) (int, error) {
	read := 0
outer:
	for {
		currentData := data[read:]

		switch r.State {
		case StateError:
			return 0, ErrorRequestInErrorState
		case StateInit:
			rl, n, err := parseRequestLine(currentData)
			if err != nil {
				return 0, err
			}
			if n == 0 {
				break outer
			}
			r.RequestLine = *rl
			read += n
			r.State = StateHeaders
		case StateHeaders:
			n, done, err := r.Headers.Parse(currentData)
			if err != nil {
				return 0, err
			}
			if n == 0 {
				break outer
			}
			read += n

			if done {
				r.State = StateDone
			}
		case StateDone:
			break outer
		default:
			panic("somehow we have programmed poorly")
		}
	}
	return read, nil
}

func (r *Request) done() bool {
	return r.State == StateDone
}
func (r *Request) error() bool {
	return r.State == StateError
}
func RequestFromReader(reader io.Reader) (*Request, error) {
	request := newRequest()

	buf := make([]byte, 1024)
	bufLen := 0
	for !request.done() && !request.error() {
		n, err := reader.Read(buf[bufLen:])
		if err != nil {
			return nil, err
		}
		bufLen += n
		readN, err := request.parse(buf[:bufLen])
		if err != nil {
			return nil, err
		}
		copy(buf, buf[readN:bufLen])
		bufLen -= readN
	}
	return request, nil
}
