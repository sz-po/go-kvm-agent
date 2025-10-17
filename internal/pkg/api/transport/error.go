package transport

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/szymonpodeszwa/go-kvm-agent/pkg/api/transport"
)

func HandleError(responseWriter http.ResponseWriter, request *http.Request, err error) {
	slog.Warn("Error while handling API request.", slog.String("error", err.Error()))

	var writeErr error

	switch typedErr := err.(type) {
	default:
		responseWriter.WriteHeader(http.StatusInternalServerError)
		writeErr = json.NewEncoder(responseWriter).Encode(transport.InternalServerError{
			Message: typedErr.Error(),
		})
	}

	if writeErr != nil {
		slog.Warn("Error while writing API error response.", slog.String("error", err.Error()))
		return
	}
}
