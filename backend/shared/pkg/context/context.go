package context

// ChatContext is similar to KernelArguments in C# - a flexible key-value store for passing parameters
type ChatContext struct {
	variables map[string]interface{}
}

// NewContext creates a new Context instance
func NewContext() *ChatContext {
	return &ChatContext{
		variables: make(map[string]interface{}),
	}
}

// Set sets a variable in the context (supports method chaining)
func (c *ChatContext) Set(key string, value interface{}) *ChatContext {
	c.variables[key] = value
	return c
}

// Get retrieves a variable from the context
func (c *ChatContext) Get(key string) (interface{}, bool) {
	val, exists := c.variables[key]
	return val, exists
}
