package transport

import "github.com/elnormous/contenttype"

type MediaType contenttype.MediaType

var ApplicationJsonMediaType = MustCreateMediaType("application/json")
var ApplicationXRgb24MediaType = MustCreateMediaType("application/x-rgb24")

func (mediaType MediaType) String() string {
	ct := contenttype.MediaType(mediaType)
	return ct.String()
}

func NewMediaType(mediaTypeQuery string) (*MediaType, error) {
	contentType, err := contenttype.ParseMediaType(mediaTypeQuery)
	if err != nil {
		return nil, err
	}

	mediaType := MediaType(contentType)

	return &mediaType, nil
}

func MustCreateMediaType(mediaTypeQuery string) MediaType {
	mediaType, err := NewMediaType(mediaTypeQuery)
	if err != nil {
		panic(err)
	}

	return *mediaType
}

func NegotiateAcceptedMediaType(header string, acceptableMediaTypes []MediaType) (MediaType, error) {
	var mediaTypes []contenttype.MediaType

	for _, acceptableMediaType := range acceptableMediaTypes {
		mediaTypes = append(mediaTypes, contenttype.MediaType(acceptableMediaType))
	}

	acceptedMediaType, _, err := contenttype.GetAcceptableMediaTypeFromHeader(header, mediaTypes)

	return MediaType(acceptedMediaType), err
}
