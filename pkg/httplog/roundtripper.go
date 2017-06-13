package httplog

import (
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
}

func logRequest(req *http.Request) {
	log.Printf("> %s %s %s", req.Method, req.URL, req.Proto)
	logHeader("> ", req.Header)
	log.Printf("> ...")
}

func logResponse(resp *http.Response) {
	log.Printf("< %s %s", resp.Proto, resp.Status)
	logHeader("< ", resp.Header)
	log.Printf("< ...")
}

type RoundTripper struct {
	http.RoundTripper
}

func (rt *RoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	transport := rt.RoundTripper
	if transport == nil {
		transport = http.DefaultTransport
	}

	logRequest(req)

	resp, err := transport.RoundTrip(req)
	if err != nil {
		return nil, err
	}

	logResponse(resp)

	return resp, err
}
