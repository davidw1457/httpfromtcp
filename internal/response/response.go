package response

import (
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/davidw1457/httpfromtcp/internal/headers"
)

type StatusCode int

const (
	OK          StatusCode = 200
	BADREQUEST  StatusCode = 400
	SERVERERROR StatusCode = 500
)

type Writer struct {
	writer io.Writer
	state  writerState
}

type writerState int

const (
	writerStateStatusLine writerState = iota
	writerStateHeaders
	writerStateBody
	writerStateTrailers
)

func NewWriter(w io.Writer) *Writer {
	return &Writer{
		writer: w,
		state:  writerStateStatusLine,
	}
}

func (w *Writer) WriteStatusLine(statusCode StatusCode) error {
	if w.state != writerStateStatusLine {
		return fmt.Errorf("writer not in writerStateStatusLine state")
	}

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

	_, err := w.writer.Write(statusLine)
	if err != nil {
		return fmt.Errorf("writeStatusLine: %w", err)
	}

	w.state = writerStateHeaders

	return nil
}

func GetDefaultHeaders(contentLen int) headers.Headers {
	header := headers.NewHeaders()

	header.Set("Content-Length", strconv.Itoa(contentLen))
	header.Set("Connection", "close")
	header.Set("content-type", "text/plain")

	return header
}

func (w *Writer) WriteHeaders(headers headers.Headers) error {
	if w.state != writerStateHeaders {
		return fmt.Errorf("writer not in writerStateHeaders")
	}
	for k, v := range headers {
		headerLine := []byte(fmt.Sprintf("%s: %s\r\n", k, v))
		_, err := w.writer.Write(headerLine)
		if err != nil {
			return fmt.Errorf("WriteHeaders: %w", err)
		}
	}

	_, err := w.writer.Write([]byte("\r\n"))
	if err != nil {
		return fmt.Errorf("WriteHeaders: %w", err)
	}

	w.state = writerStateBody

	return nil
}

func (w *Writer) WriteBody(p []byte) (int, error) {
	if w.state != writerStateBody {
		return 0, fmt.Errorf("writer not in writerStatebody")
	}
	n, err := w.writer.Write(p)
	if err != nil {
		return 0, fmt.Errorf("writer.WriteBody: %w", err)
	}

	return n, nil
}

func (w *Writer) WriteChunkedBody(p []byte) (int, error) {
	length := strconv.FormatInt(int64(len(p)), 16)

	bytesWritten := 0

	n, err := w.WriteBody([]byte(fmt.Sprintf("%s\r\n", length)))
	if err != nil {
		return 0, fmt.Errorf("writer.WriteChunkedBody: %w", err)
	}
	bytesWritten += n

	n, err = w.WriteBody(p)
	if err != nil {
		return 0, fmt.Errorf("writer.WriteChunkedBody: %w", err)
	}
	bytesWritten += n

	n, err = w.WriteBody([]byte("\r\n"))
	if err != nil {
		return 0, fmt.Errorf("writer.WriteChunkedBody: %w", err)
	}
	bytesWritten += n

	return bytesWritten, nil
}

func (w *Writer) WriteChunkedBodyDone() (int, error) {
	n, err := w.WriteBody([]byte("0\r\n"))
	if err != nil {
		return 0, fmt.Errorf("writer.WriteChunkedBodyDone: %w", err)
	}

	w.state = writerStateTrailers

	return n, nil
}

func (w *Writer) WriteTrailers(h headers.Headers) error {
	trailers, _ := h.Get("Trailer")

	trailerKeys := strings.Split(trailers, ",")
	for _, k := range trailerKeys {
		k = strings.TrimSpace(k)
		v, ok := h.Get(k)
		if !ok {
			continue
		}

		trailerLine := []byte(fmt.Sprintf("%s: %s\r\n", k, v))
		_, err := w.writer.Write(trailerLine)
		if err != nil {
			return fmt.Errorf("writer.WriteTrailers: %w", err)
		}
	}

	_, err := w.writer.Write([]byte("\r\n"))
	if err != nil {
		return fmt.Errorf("writer.WriteTrailers: %w", err)
	}

	return nil
}
