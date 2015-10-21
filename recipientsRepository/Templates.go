package recipientsRepository

import (
	"html/template"
)

type Templates struct {
	SubjectTemplate *template.Template
	BodyTemplate *template.Template

	ReceiverEmailTemplate *template.Template
	ReceiverNameTemplate *template.Template

	SenderEmailTemplate *template.Template
	SenderNameTemplate *template.Template

	ReplyToEmailTemplate *template.Template

	LinkTemplates map[string]LinkTemplate
}

type LinkTemplate struct {
	EncryptionKey string
	BaseLink string
	Parameters []string
}

type TemplatesString struct {
	SubjectTemplate string
	BodyTemplate string

	ReceiverEmailTemplate string
	ReceiverNameTemplate string

	SenderEmailTemplate string
	SenderNameTemplate string

	ReplyToEmailTemplate string

	LinkTemplates map[string]LinkTemplate
}

func (this *TemplatesString) Parse() (templates Templates, err error) {
	templates = Templates{}

	templates.SubjectTemplate, err = template.New("SubjectTemplate").Parse(this.SubjectTemplate)
	if err != nil { return }

	templates.BodyTemplate, err = template.New("BodyTemplate").Parse(this.BodyTemplate)
	if err != nil { return }

	templates.ReceiverEmailTemplate, err = template.New("ReceiverEmailTemplate").Parse(this.ReceiverEmailTemplate)
	if err != nil { return }

	templates.ReceiverNameTemplate, err = template.New("ReceiverNameTemplate").Parse(this.ReceiverNameTemplate)
	if err != nil { return }

	templates.SenderEmailTemplate, err = template.New("SenderEmailTemplate").Parse(this.SenderEmailTemplate)
	if err != nil { return }

	templates.SenderNameTemplate, err = template.New("SenderNameTemplate").Parse(this.SenderNameTemplate)
	if err != nil { return }

	templates.ReplyToEmailTemplate, err = template.New("ReplyToEmailTemplate").Parse(this.ReplyToEmailTemplate)
	if err != nil { return }

	templates.LinkTemplates = this.LinkTemplates

	return
}