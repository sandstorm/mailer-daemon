package worker

import (
	"sandstormmedia/project-webessentials-mailer/go/mailer/recipientsRepository"
	"bytes"
	"html/template"
	"fmt"
	"strings"
	"sandstormmedia/project-webessentials-mailer/go/mailer/hmacs"
)

type RenderedTemplates struct {
	Subject string
	Body string

	ReceiverEmail string
	ReceiverName string

	SenderEmail string
	SenderName string

	ReplyToEmail string
}

func RenderTemplates(templates recipientsRepository.Templates, recipient map[string]interface{}) (result RenderedTemplates, err error) {
	result = RenderedTemplates{}

	for placeholder, t := range templates.LinkTemplates {
		link := renderLinkTemplate(&t, &recipient)
		recipient[placeholder] = template.HTMLAttr(fmt.Sprintf("href=\"%s\"", link))
	}

	result.Subject, err = renderHtmlTemplate(templates.SubjectTemplate, &recipient)
	if err != nil { return }

	result.ReceiverName, err = renderHtmlTemplate(templates.ReceiverNameTemplate, &recipient)
	if err != nil { return }

	result.ReceiverEmail, err = renderHtmlTemplate(templates.ReceiverEmailTemplate, &recipient)
	if err != nil { return }

	result.SenderName, err = renderHtmlTemplate(templates.SenderNameTemplate, &recipient)
	if err != nil { return }

	result.SenderEmail, err = renderHtmlTemplate(templates.SenderEmailTemplate, &recipient)
	if err != nil { return }

	result.ReplyToEmail, err = renderHtmlTemplate(templates.ReplyToEmailTemplate, &recipient)
	if err != nil { return }


	result.Body, err = renderHtmlTemplate(templates.BodyTemplate, &recipient)
	if err != nil { return }

	return
}

func renderHtmlTemplate(template *template.Template, recipient *map[string]interface{}) (result string, err error) {
	buf := new(bytes.Buffer)
	err = template.Execute(buf, recipient)
	if err != nil { return }

	result = buf.String()
	return
}

func renderLinkTemplate(template *recipientsRepository.LinkTemplate, recipient *map[string]interface{}) (result string) {
	// collect parameters
	parameters := make([]string, len(template.Parameters))
	for index, parameter := range template.Parameters {
		parameters[index] = fmt.Sprintf("%s=%s", parameter, (*recipient)[parameter])
	}

	// generate link
	var linkWithoutHmac string
	if strings.Contains(template.BaseLink, "?") {
		linkWithoutHmac = fmt.Sprintf("%s&%s", template.BaseLink, strings.Join(parameters, "&"))
	} else {
		linkWithoutHmac = fmt.Sprintf("%s?%s", template.BaseLink, strings.Join(parameters, "&"))
	}

	// protect link from manipulation
	hmacGenerator := hmacs.HashHmacGenerator{
		EncryptionKey: []byte(template.EncryptionKey),
	}
	return fmt.Sprintf("%s&hmac=%s", linkWithoutHmac, hmacGenerator.Sha1String(linkWithoutHmac))
}