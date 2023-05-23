package registry

// Message Context Implementation

// DefaultContext is a global variable that serves as default message context (it commonly contains the DefaultRegistry
// as Registry and a default configuration of its settings).
// If you want to use a different configuration, you have manually create a respective Context object.
var DefaultContext = &Context{
	MessageTypeRegistry: DefaultMessageRegistry,
	PrintSettings: &PrintSettings{
		Uppercase: false,
	},
}

// PrintSettings or rather instances of PrintSettings, are accessed within the Print method of Messages and influence
// the method's behavior (print the message with uppercase letters only if Uppercase is set to true).
type PrintSettings struct {
	Uppercase bool
}

// A Context object bundles the Settings (here, PrintSettings) and the extension point implementations (the types
// registered at the MessageTypeRegistry)
type Context struct {
	MessageTypeRegistry MessageTypeRegistry
	PrintSettings       *PrintSettings
}

func (c *Context) MessageForSpec(spec MessageSpec) Message {
	return spec.Message(c)
}

func (c *Context) MessageSpecForConfig(data []byte) (MessageSpec, error) {
	messageSpec, err := c.MessageTypeRegistry.DecodeMessage(data)
	if err != nil {
		return nil, err
	}

	return messageSpec, nil
}

func (c *Context) MessageForConfig(data []byte) (Message, error) {
	spec, err := c.MessageSpecForConfig(data)
	if err != nil {
		return nil, err
	}
	return spec.Message(c), nil
}
