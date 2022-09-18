package preview

type Container struct {
	Headers map[string]string
	Body    []byte
}

func NewContainer() *Container {
	return &Container{
		Headers: make(map[string]string),
	}
}

func (c *Container) SetHeader(name, value string) {
	c.Headers[name] = value
}

func (c *Container) Len() int {
	return len(c.Body)
}
