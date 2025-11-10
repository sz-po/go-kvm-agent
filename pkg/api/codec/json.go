package codec

import (
	"encoding/json"
	"io"

	"github.com/szymonpodeszwa/go-kvm-agent/pkg/api"
)

type Json struct {
	encoder *json.Encoder
	decoder *json.Decoder
}

var _ api.Codec = (*Json)(nil)

func NewJsonCodec(stream io.ReadWriter) *Json {
	encoder := json.NewEncoder(stream)
	decoder := json.NewDecoder(stream)

	return &Json{
		encoder: encoder,
		decoder: decoder,
	}
}

func (codec *Json) Encode(value any) error {
	return codec.encoder.Encode(value)
}

func (codec *Json) Decode(value any) error {
	return codec.decoder.Decode(value)
}
