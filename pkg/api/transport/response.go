package transport

type Response struct {
	StatusCode int
	Header     map[string]string
	Body       any
}
