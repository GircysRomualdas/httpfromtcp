package response

import (
	"fmt"
	"io"

	"github.com/GircysRomualdas/httpfromtcp/internal/headers"
)

type Writer struct {
	writer io.Writer
}

func (w *Writer) WriteStatusLine(statusCode StatusCode) error {
	statusLine := []byte{}

	switch statusCode {
	case OK:
		statusLine = []byte("HTTP/1.1 200 OK\r\n")
	case BadRequest:
		statusLine = []byte("HTTP/1.1 400 Bad Request\r\n")
	case InternalServerError:
		statusLine = []byte("HTTP/1.1 500 Internal Server Error\r\n")
	default:
		return fmt.Errorf("unknown status code: %d", statusCode)
	}

	_, err := w.writer.Write(statusLine)
	return err
}

func NewWriter(writer io.Writer) *Writer {
	return &Writer{writer: writer}
}

func (w *Writer) WriteHeaders(headers headers.Headers) error {
	for key, value := range headers {
		_, err := w.writer.Write([]byte(key + ": " + value + "\r\n"))
		if err != nil {
			return err
		}
	}
	_, err := w.writer.Write([]byte("\r\n"))
	return err
}

func (w *Writer) WriteBody(p []byte) (int, error) {
	n, err := w.writer.Write(p)
	return n, err
}

type StatusCode int

const (
	OK                  StatusCode = 200
	BadRequest          StatusCode = 400
	InternalServerError StatusCode = 500
)

func GetDefaultHeaders(contentLen int) headers.Headers {
	return headers.Headers{
		"content-length": fmt.Sprintf("%d", contentLen),
		"connection":     "close",
		"content-type":   "text/plain",
	}
}

func (w *Writer) WriteChunkedBody(p []byte) (int, error) {
	if len(p) == 0 {
		return 0, nil
	}

	hdr := []byte(fmt.Sprintf("%x\r\n", len(p)))
	buf := make([]byte, 0, len(hdr)+len(p)+2)
	buf = append(buf, hdr...)
	buf = append(buf, p...)
	buf = append(buf, '\r', '\n')
	return w.writeAll(buf)
}

func (w *Writer) writeAll(b []byte) (int, error) {
	total := 0
	for total < len(b) {
		n, err := w.writer.Write(b[total:])
		total += n
		if err != nil {
			return total, err
		}
	}
	return total, nil
}

func (w *Writer) WriteChunkedBodyDone() (int, error) {
	return w.writeAll([]byte("0\r\n"))
}

func (w *Writer) WriteTrailers(h headers.Headers) error {
	for key, value := range h {
		_, err := w.writer.Write([]byte(key + ": " + value + "\r\n"))
		if err != nil {
			return err
		}
	}
	_, err := w.writer.Write([]byte("\r\n"))
	return err
}
