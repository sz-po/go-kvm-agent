package transport

import "net/http"

func parseHeader(request *http.Request) map[string]string {
	header := make(map[string]string)

	for key, values := range request.Header {
		if len(values) > 0 {
			header[key] = values[0]
		}
	}

	return header
}

func writeHeader(w http.ResponseWriter, headers map[string]string) {
	for key, value := range headers {
		w.Header().Set(key, value)
	}
}
