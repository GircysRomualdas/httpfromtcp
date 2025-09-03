package main

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"github.com/GircysRomualdas/httpfromtcp/internal/headers"
	"github.com/GircysRomualdas/httpfromtcp/internal/request"
	"github.com/GircysRomualdas/httpfromtcp/internal/response"
	"github.com/GircysRomualdas/httpfromtcp/internal/server"
)

const port = 42069

func main() {
	server, err := server.Serve(port, handler)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
	defer server.Close()
	log.Println("Server started on port", port)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("Server gracefully stopped")
}

func handler(w *response.Writer, req *request.Request) {
	if req.RequestLine.RequestTarget == "/yourproblem" {
		handler400(w, req)
		return
	} else if req.RequestLine.RequestTarget == "/myproblem" {
		handler500(w, req)
		return
	} else if strings.HasPrefix(req.RequestLine.RequestTarget, "/httpbin/") {
		target := strings.TrimPrefix(req.RequestLine.RequestTarget, "/httpbin/")
		res, err := http.Get("https://httpbin.org/" + target)
		if err != nil {
			handler500(w, req)
			return
		}

		w.WriteStatusLine(response.OK)
		h := headers.NewHeaders()
		h.Set("Content-Type", "text/plain")
		h.Set("Transfer-Encoding", "chunked")
		h.Set("Trailer", "X-Content-SHA256")
		h.Set("Trailer", "X-Content-Length")
		w.WriteHeaders(h)

		buf := make([]byte, 1024)
		data := make([]byte, 0)
		for {
			n, err := res.Body.Read(buf)
			if n > 0 {

				if _, err = w.WriteChunkedBody(buf[:n]); err != nil {
					return
				}
				data = append(data, buf[:n]...)
			}
			if err == io.EOF {
				break
			}
			if err != nil {
				return
			}
		}
		w.WriteChunkedBodyDone()
		trailers := headers.NewHeaders()
		sum := sha256.Sum256(data)
		hash := hex.EncodeToString(sum[:])
		trailers.Set("X-Content-SHA256", hash)
		trailers.Set("X-Content-Length", strconv.Itoa(len(data)))
		w.WriteTrailers(trailers)
		return
	} else if req.RequestLine.RequestTarget == "/video" {
		w.WriteStatusLine(response.OK)
		data, err := os.ReadFile("assets/vim.mp4")
		if err != nil {
			handler500(w, req)
			return
		}
		h := response.GetDefaultHeaders(len(data))
		h.Override("Content-Type", "video/mp4")
		w.WriteHeaders(h)
		w.WriteBody(data)
	}
	handler200(w, req)
}

func handler400(w *response.Writer, _ *request.Request) {
	w.WriteStatusLine(response.BadRequest)
	body := []byte(`<html>
<head>
<title>400 Bad Request</title>
</head>
<body>
<h1>Bad Request</h1>
<p>Your request honestly kinda sucked.</p>
</body>
</html>
`)
	h := response.GetDefaultHeaders(len(body))
	h.Override("Content-Type", "text/html")
	w.WriteHeaders(h)
	w.WriteBody(body)
}

func handler500(w *response.Writer, _ *request.Request) {
	w.WriteStatusLine(response.InternalServerError)
	body := []byte(`<html>
<head>
<title>500 Internal Server Error</title>
</head>
<body>
<h1>Internal Server Error</h1>
<p>Okay, you know what? This one is on me.</p>
</body>
</html>
`)
	h := response.GetDefaultHeaders(len(body))
	h.Override("Content-Type", "text/html")
	w.WriteHeaders(h)
	w.WriteBody(body)
}

func handler200(w *response.Writer, _ *request.Request) {
	w.WriteStatusLine(response.OK)
	body := []byte(`<html>
<head>
<title>200 OK</title>
</head>
<body>
<h1>Success!</h1>
<p>Your request was an absolute banger.</p>
</body>
</html>
`)
	h := response.GetDefaultHeaders(len(body))
	h.Override("Content-Type", "text/html")
	w.WriteHeaders(h)
	w.WriteBody(body)
}
