package request

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"

	"github.com/davidw1457/httpfromtcp/internal/headers"
)

const bufferSize = 8

type Request struct {
	RequestLine RequestLine
	Headers     headers.Headers
	Body        []byte

	state          requestState
	bodyLengthRead int
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

type requestState int

const (
	requestStateInitialized requestState = iota
	requestStateParsingHeaders
	requestStateParsingBody
	requestStateDone
)

func RequestFromReader(reader io.Reader) (*Request, error) {
	request := Request{
		Headers: headers.NewHeaders(),
		Body:    make([]byte, 0),
		state:   requestStateInitialized,
	}
	buf := make([]byte, bufferSize)
	readToIndex := 0

	for request.state != requestStateDone {
		if readToIndex >= len(buf) {
			newBuf := make([]byte, len(buf)*2)
			copy(newBuf, buf)
			buf = newBuf
		}

		n, err := reader.Read(buf[readToIndex:])
		if err != nil {
			if errors.Is(err, io.EOF) {
				if request.state != requestStateDone {
					return nil, fmt.Errorf("incomplete request")
				}
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
		return nil, 0, fmt.Errorf("parseRequestLine: %w", err)
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
	bytesParsed := 0
	for r.state != requestStateDone {
		n, err := r.parseSingle(data[bytesParsed:])
		if err != nil {
			return n, fmt.Errorf("request.parse: %w", err)
		}
		bytesParsed += n
		if n == 0 {
			break
		}
	}
	return bytesParsed, nil
}

func (r *Request) parseSingle(data []byte) (int, error) {
	switch r.state {
	case requestStateDone:
		return 0, fmt.Errorf("error: trying to read data in a done state")
	case requestStateInitialized:
		requestLine, n, err := parseRequestLine(data)
		if err != nil {
			return 0, fmt.Errorf("request.parse: %w", err)
		}

		if n == 0 {
			return 0, nil
		}

		r.RequestLine = *requestLine
		r.state = requestStateParsingHeaders

		return n, nil
	case requestStateParsingHeaders:
		n, done, err := r.Headers.Parse(data)
		if err != nil {
			return 0, fmt.Errorf("request.parse: %w", err)
		}

		if done {
			r.state = requestStateParsingBody
		}
		return n, nil
	case requestStateParsingBody:
		contentLengthStr, ok := r.Headers.Get("Content-Length")
		if !ok {
			r.state = requestStateDone
			return len(data), nil
		}

		contentLength, err := strconv.Atoi(contentLengthStr)
		if err != nil {
			return 0, fmt.Errorf("request.parse:: %w", err)
		}

		r.Body = append(r.Body, data...)
		r.bodyLengthRead += len(data)
		if r.bodyLengthRead > contentLength {
			return 0, fmt.Errorf("too much body")
		} else if r.bodyLengthRead == contentLength {
			r.state = requestStateDone
		}
		return len(data), nil
	default:
		return 0, fmt.Errorf("invalid state")
	}
}
