package headers

import (
	"bytes"
	"fmt"
)

type Headers map[string]string

var crlf = []byte("\r\n")

func NewHeaders() Headers {
	return map[string]string{}
}

func parseHeader(filedline []byte) (string, string, error) {
	parts := bytes.SplitN(filedline, []byte(":"), 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("malformed field line")
	}
	name := parts[0]
	value := bytes.TrimSpace(parts[1])

	if bytes.HasSuffix(name, []byte(" ")) || bytes.HasPrefix(name, []byte(" ")) {
		return "", "", fmt.Errorf("malformed field name")
	}
	return string(name), string(value), nil
}

func (h Headers) Parse(data []byte) (int, bool, error) {
	read := 0
	done := false

	for {
		idx := bytes.Index(data[read:], crlf)

		if idx == -1 {
			break
		}
		//empty header
		if idx == 0 {
			done = true
			read += len(crlf)
			break
		}
		name, value, err := parseHeader(data[read : read+idx])
		if err != nil {
			return 0, false, err
		}

		read += idx + len(crlf)
		h[name] = value
	}
	return read, done, nil
}
