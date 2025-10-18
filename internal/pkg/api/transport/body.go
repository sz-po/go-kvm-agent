package transport

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
)

func writeBody(responseWriter http.ResponseWriter, body any) {
	if body == nil {
		return
	}

	var err error

	switch typedBody := body.(type) {
	case error:
		_, err = responseWriter.Write([]byte(typedBody.Error()))
	case io.Reader:
		_, err = io.Copy(responseWriter, typedBody)
	case io.WriterTo:
		_, err = typedBody.WriteTo(responseWriter)
	case string:
		_, err = responseWriter.Write([]byte(typedBody))
	default:
		err = json.NewEncoder(responseWriter).Encode(body)
	}
	if err != nil {
		slog.Warn("Error while writing API response.", slog.String("error", err.Error()))
		return
	}
}
