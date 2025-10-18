package transport

import (
	"net/http"
	"strings"
)

func parseHeaders(request *http.Request) map[string]string {
	header := make(map[string]string)

	for key, values := range request.Header {
		if len(values) > 0 {
			header[strings.ToLower(key)] = values[0]
		}
	}

	return header
}

func writeHeader(w http.ResponseWriter, headers map[string]string) {
	for key, value := range headers {
		w.Header().Set(key, value)
	}
}
