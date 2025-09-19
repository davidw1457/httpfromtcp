package headers

import (
	"bytes"
	"fmt"
	"strings"
)

type Headers map[string]string

func (h Headers) Parse(data []byte) (n int, done bool, err error) {
	kv, n, err := parseHeaderLine(data)
	if err != nil {
		return 0, false, fmt.Errorf("headers.Parse: %w", err)
	}
	if n == 0 {
		return n, false, nil
	}
	if kv == nil {
		return n, true, nil
	}

	h[kv[0]] = kv[1]
	return n, false, nil
}

func NewHeaders() Headers {
	return Headers(make(map[string]string))
}

func parseHeaderLine(input []byte) ([]string, int, error) {
	idx := bytes.Index(input, []byte("\r\n"))
	if idx == -1 {
		return nil, 0, nil
	}

	line := string(input[:idx])
	kv, err := headerLineFromString(line)
	if err != nil {
		return nil, idx + 2, fmt.Errorf("parseHeaderLine: %w", err)
	}

	return kv, idx + 2, nil
}

func headerLineFromString(line string) ([]string, error) {
	if len(line) == 0 {
		return nil, nil
	}

	idx := strings.Index(line, ":")
	if idx < 1 {
		return nil, fmt.Errorf(
			"headerLineFromString: improper header line: %s",
			line,
		)
	}

	key := strings.TrimSpace(line[:idx])
	if strings.TrimSpace(line[idx-1:idx]) == "" || key == "" {
		return nil, fmt.Errorf(
			"headerLineFromString: improper key value: %s",
			line,
		)
	}

	value := strings.TrimSpace(line[idx+1:])

	return []string{key, value}, nil
}
