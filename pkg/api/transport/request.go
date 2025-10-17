package transport

import "io"

type Request struct {
	Path   Path
	Query  map[string]string
	Header map[string]string
	Body   io.Reader
}
