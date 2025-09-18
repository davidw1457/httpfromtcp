package request

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"regexp"
	"strings"
)

const bufferSize = 8

type Request struct {
	RequestLine RequestLine
	state requestState // 0 = init; 1 = done
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

type requestState int

const (
    requestStateInitialized requestState = iota
    requestStateDone
)

func RequestFromReader(reader io.Reader) (*Request, error) {
	request := Request{state: requestStateInitialized}
	buf := make([]byte, bufferSize)
	readToIndex := 0

	for request.state != requestStateDone {
		if readToIndex == len(buf) {
			newBuf := make([]byte, len(buf)*2)
			copy(newBuf, buf)
			buf = newBuf
		}

		n, err := reader.Read(buf[readToIndex:])
		if err != nil {
			if errors.Is(err, io.EOF) {
				request.state = requestStateDone
				break
			}
			return nil, fmt.Errorf("RequestFromReader: %w", err)
		}

		readToIndex += n
		n, err = request.parse(buf[:readToIndex])
		if err != nil {
			return nil, fmt.Errorf("RequestFromReader: %w", err)
		}

		copy(buf, buf[n:])
		readToIndex -= n
	}

	return &request, nil
}

func parseRequestLine(input []byte) (*RequestLine, int, error) {
	idx := bytes.Index(input, []byte("\r\n"))
	if idx == -1 {
		return nil, 0, nil
	}
	requestLineText := string(input[:idx])
	requestLine, err := requestLineFromString(requestLineText)
	if err != nil {
		return nil, idx, fmt.Errorf("parseRequestLine: %w", err)
	}

	return requestLine, idx + 2, nil
}

func requestLineFromString(line string) (*RequestLine, error) {
	fields := strings.Fields(line)

	if len(fields) != 3 {
		return nil, fmt.Errorf(
			"requestLineFromString: request-line missing fields %s",
			line,
		)
	}

	validMethod, err := regexp.Match("[A-Z]+$", []byte(fields[0]))
	if err != nil {
		return nil, fmt.Errorf("parseRequestLine: %w", err)
	} else if !validMethod {
		return nil, fmt.Errorf(
			"parseRequestLine: invalid method: %s",
			fields[0],
		)
	}

	if !strings.Contains(fields[2], "HTTP/1.1") {
		return nil, fmt.Errorf(
			"parseRequestLine: invalid http version: %s",
			fields[2],
		)
	}

	parsedRL := RequestLine{
		HttpVersion:   "1.1",
		RequestTarget: fields[1],
		Method:        fields[0],
	}

	return &parsedRL, nil
}

func (r *Request) parse(data []byte) (int, error) {
	if r.state == requestStateDone {
		return 0, fmt.Errorf("error: trying to read data in a done state")
	}

	requestLine, n, err := parseRequestLine(data)
	if err != nil {
		return n, fmt.Errorf("request.parse: %w", err)
	}

	if requestLine != nil {
		r.RequestLine = *requestLine
		r.state = requestStateDone
	}

	return n, nil
}
