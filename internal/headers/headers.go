package headers

import (
	"bytes"
	"fmt"
	"regexp"
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
	if kv == nil && n == 2 {
		return n, true, nil
	}
	if old, ok := h[kv[0]]; ok {
		kv[1] = fmt.Sprintf("%s, %s", old, kv[1])
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
	if idx == 0 {
		return nil, 2, nil
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

	key := strings.ToLower(strings.TrimSpace(line[:idx]))
	isValid, err := regexp.MatchString("^[\\w!#$%&'*+-.^_`|~]+$", key)
	if err != nil {
		return nil, fmt.Errorf(
			"headerLineFromString: %w",
			err,
		)
	}
	if strings.TrimSpace(line[idx-1:idx]) == "" || key == "" || !isValid {
		return nil, fmt.Errorf(
			"headerLineFromString: improper key value: %s",
			line,
		)
	}

	value := strings.TrimSpace(line[idx+1:])

	return []string{key, value}, nil
}

func (h Headers) Get(key string) (string, bool) {
	key = strings.ToLower(key)
	val, ok := h[key]
	return val, ok
}

func (h Headers) Set(key string, val string) {
	key = strings.ToLower(key)
	h[key] = val
}

func (h Headers) Add(key string, val string) {
	key = strings.ToLower(key)
	if v, ok := h[key]; ok {
		val = fmt.Sprintf("%s, %s", v, val)
	}
	h[key] = val
}

func (h Headers) Delete(key string) {
	key = strings.ToLower(key)
	delete(h, key)
}
