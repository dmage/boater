package httplog

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"sort"
)

func logHeader(prefix string, header http.Header) {
	var keys []string
	for k := range header {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		for _, v := range header[k] {
			log.Printf("%s%s: %s", prefix, k, v)
		}
	}
	log.Printf("%s", prefix)
}

type logReader struct {
	prefix  string
	rc      io.ReadCloser
	n       int
	toLog   []byte
	trunc   bool
	printed bool
}

func newLogReader(prefix string, rc io.ReadCloser, n int) io.ReadCloser {
	if rc == nil {
		return nil
	}
	return &logReader{
		prefix: prefix,
		rc:     rc,
		n:      n,
	}
}

func (r *logReader) Read(b []byte) (int, error) {
	n, err := r.rc.Read(b)
	if n > 0 {
		if len(r.toLog) < r.n {
			i := n
			if cap := r.n - len(r.toLog); i > cap {
				r.trunc = true
				i = cap
			}
			r.toLog = append(r.toLog, b[:i]...)
			if r.trunc {
				r.print()
			}
		} else {
			r.trunc = true
			r.print()
		}
	}
	return n, err
}

func (r *logReader) print() {
	if r.printed {
		return
	}
	for offset := 0; offset < len(r.toLog); offset += 0x10 {
		chunk := r.toLog[offset:]
		if len(chunk) > 0x10 {
			chunk = chunk[:0x10]
		}

		chunkHex := fmt.Sprintf("% 02x", chunk[:8])
		if len(chunk) > 8 {
			chunkHex += "  " + fmt.Sprintf("% 02x", chunk[8:])
		}

		for i, b := range chunk {
			if b < 32 || b > 126 {
				chunk[i] = '.'
			}
		}

		log.Printf("%s%08x  %-48s  |%s|", r.prefix, offset, chunkHex, chunk)
	}
	if r.trunc {
		log.Printf("%s...", r.prefix)
	}
	r.printed = true
}

func (r *logReader) Close() error {
	r.print()
	return r.rc.Close()
}

func logRequest(req *http.Request, n int) {
	log.Printf("-> %s %s %s", req.Method, req.URL, req.Proto)
	logHeader("-> ", req.Header)
	req.Body = newLogReader("-> ", req.Body, n)
}

func logResponse(resp *http.Response, n int) {
	log.Printf("<- %s %s", resp.Proto, resp.Status)
	logHeader("<- ", resp.Header)
	resp.Body = newLogReader("<- ", resp.Body, n)
}

type RoundTripper struct {
	http.RoundTripper
	N int // log N first bytes
}

func (rt *RoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	transport := rt.RoundTripper
	if transport == nil {
		transport = http.DefaultTransport
	}

	logRequest(req, rt.N)

	resp, err := transport.RoundTrip(req)
	if err != nil {
		return nil, err
	}

	logResponse(resp, rt.N)

	return resp, err
}
