package email

type Message struct {
	FromName     string
	FromEmail    string
	ToName       string
	ToEmail      string
	TemplateName string
	TemplateData map[string]interface{}
}

type Service interface {
	Send(msg Message) error
}
