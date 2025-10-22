package transport

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
)

type Response struct {
	StatusCode int
	Header     map[string]string
	Body       any
}

type ResponseWriterTo struct {
	io.ReadCloser
}

func (writer ResponseWriterTo) WriteTo(dst io.Writer) (int64, error) {
	return io.Copy(dst, writer.ReadCloser)
}

func UnmarshalResponseBody[T any](body any) (*T, error) {
	var responseBody T
	var err error

	switch typedBody := body.(type) {
	case io.ReadCloser:
		err = json.NewDecoder(typedBody).Decode(&responseBody)
		_ = typedBody.Close()
	case io.Reader:
		err = json.NewDecoder(typedBody).Decode(&responseBody)
	case []byte:
		err = json.NewDecoder(bytes.NewReader(typedBody)).Decode(&responseBody)
	default:
		return nil, fmt.Errorf("unsupported type: %T", typedBody)
	}

	if err != nil {
		return nil, err
	}

	return &responseBody, nil
}
