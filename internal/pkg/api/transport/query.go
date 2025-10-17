package transport

import "net/http"

func parseQuery(request *http.Request) map[string]string {
	query := make(map[string]string)

	for key, values := range request.URL.Query() {
		if len(values) > 0 {
			query[key] = values[0]
		}
	}

	return query
}
