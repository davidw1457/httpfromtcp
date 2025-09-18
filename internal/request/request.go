package request

import (
	"fmt"
	"io"
	"regexp"
	"strings"
)

type Request struct {
	RequestLine RequestLine
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	input, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("RequestFromReader: %w", err)
	}

	lines := strings.Split(string(input), "\r\n")
	rl, err := parseRequestLine(lines[0])
	if err != nil {
		return nil, fmt.Errorf("RequestFromReader: %w", err)
	}

	rq := Request{RequestLine: *rl}

	return &rq, nil
}

func parseRequestLine(rl string) (*RequestLine, error) {
	rlFields := strings.Fields(rl)

	if len(rlFields) != 3 {
		return nil, fmt.Errorf(
			"parseRequestLine: request-line missing fields %s",
			rl,
		)
	}

	validMethod, err := regexp.Match("[A-Z]+$", []byte(rlFields[0]))
	if err != nil {
		return nil, fmt.Errorf("parseRequestLine: %w", err)
	} else if !validMethod {
		return nil, fmt.Errorf(
			"parseRequestLine: invalid method: %s",
			rlFields[0],
		)
	}

	if !strings.Contains(rlFields[2], "HTTP/1.1") {
		return nil, fmt.Errorf(
			"parseRequestLine: Invalid HTTP version: %s",
			rlFields[2],
		)
	}

	parsedRL := RequestLine{
		HttpVersion:   "1.1",
		RequestTarget: rlFields[1],
		Method:        rlFields[0],
	}

	return &parsedRL, nil
}
