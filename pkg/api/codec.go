package api

type Codec interface {
	Encode(value any) error
	Decode(value any) error
}
