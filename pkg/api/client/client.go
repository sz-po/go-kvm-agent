package client

type Client struct {
	remoteAddress string
	remotePort    int
}

func NewClient(remoteAddress string, remotePort int) *Client {
	return &Client{
		remoteAddress: remoteAddress,
		remotePort:    remotePort,
	}
}
