package network_client

type HTTPChannel struct {
	handler func(msg []byte)
}

func NewHTTPChannel() *HTTPChannel {
	return &HTTPChannel{
		handler: nil,
	}
}

func (c *HTTPChannel) Setup() error {
	return nil
}

func (c *HTTPChannel) Close() error {
	return nil
}

func (c *HTTPChannel) Send(msg []byte) error {
	return nil
}

func (c *HTTPChannel) SetMsgHandler(handler func(msg []byte)) {
	c.handler = handler
}
