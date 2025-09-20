package response

import (
	"fmt"
	"io"
	"strconv"

	"github.com/davidw1457/httpfromtcp/internal/headers"
)

type StatusCode int

const (
	OK          StatusCode = 200
	BADREQUEST  StatusCode = 400
	SERVERERROR StatusCode = 500
)

func WriteStatusLine(w io.Writer, statusCode StatusCode) error {
	var phrase string
	switch statusCode {
	case OK:
		phrase = "OK"
	case BADREQUEST:
		phrase = "Bad Request"
	case SERVERERROR:
		phrase = "Internal Server Error"
	default:
		phrase = ""
	}

	statusLine := []byte(fmt.Sprintf("HTTP/1.1 %d %s\r\n", statusCode, phrase))

	_, err := w.Write(statusLine)
	if err != nil {
		return fmt.Errorf("writeStatusLine: %w", err)
	}

	return nil
}

func GetDefaultHeaders(contentLen int) headers.Headers {
	header := headers.NewHeaders()

	header["Content-Length"] = strconv.Itoa(contentLen)
	header["Connection"] = "close"
	header["Content-Type"] = "text/plain"

	return header
}

func WriteHeaders(w io.Writer, headers headers.Headers) error {
	for k, v := range headers {
		headerLine := []byte(fmt.Sprintf("%s: %s\r\n", k, v))
		_, err := w.Write(headerLine)
		if err != nil {
			return fmt.Errorf("WriteHeaders: %w", err)
		}
	}

	_, err := w.Write([]byte("\r\n"))
	if err != nil {
		return fmt.Errorf("WriteHeaders: %w", err)
	}

	return nil
}
