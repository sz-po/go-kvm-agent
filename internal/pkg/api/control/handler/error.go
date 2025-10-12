package handler

import (
	"encoding/json"
	"net/http"
)

func handleJsonSyntaxError(response http.ResponseWriter, jsonSyntaxError json.SyntaxError) error {
	response.Header().Set("Content-Type", "application/json")
	response.WriteHeader(http.StatusBadRequest)

	responseBody := struct {
		Message  string `json:"message"`
		Position int64  `json:"position"`
	}{
		Message:  jsonSyntaxError.Error(),
		Position: jsonSyntaxError.Offset,
	}

	return json.NewEncoder(response).Encode(responseBody)
}

func handleUnknownError(response http.ResponseWriter, err error) error {
	response.Header().Set("Content-Type", "application/json")
	response.WriteHeader(http.StatusInternalServerError)

	responseBody := struct {
		Message string `json:"message"`
	}{
		Message: err.Error(),
	}

	return json.NewEncoder(response).Encode(responseBody)
}

func handleError(response http.ResponseWriter, err error) error {
	switch targetErr := err.(type) {
	case *json.SyntaxError:
		return handleJsonSyntaxError(response, *targetErr)
	default:
		return handleUnknownError(response, err)
	}
}

func mustHandleError(response http.ResponseWriter, err error) {
	_ = handleError(response, err)
}
