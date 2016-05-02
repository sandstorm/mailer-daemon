package worker

import (
	"testing"
	"github.com/sandstorm/mailer-daemon/recipientsRepository"
	"encoding/json"
)

type Test struct {
	test *testing.T
}

func (this *Test) AssertIsNil(message string, actual interface{}) {
	if actual != nil {
		this.test.Errorf(message + ": actual should be nil, but is '%+v'", actual)
	}
}

func (this *Test) AssertEquals(message string, expected, actual interface{}) {
	if actual != expected {
		this.test.Errorf(message + ": expected '%+v' is different from actual '%+v'", expected, actual)
	}
}

func (this *Test) AssertRenderedTemplatesEquals(message string, expected, actual RenderedTemplates) {
	this.AssertEquals(message + ": incorrect Subject", expected.Subject, actual.Subject)
	this.AssertEquals(message + ": incorrect Body", expected.Body, actual.Body)
	this.AssertEquals(message + ": incorrect ReceiverEmail", expected.ReceiverEmail, actual.ReceiverEmail)
	this.AssertEquals(message + ": incorrect ReceiverName", expected.ReceiverName, actual.ReceiverName)
	this.AssertEquals(message + ": incorrect SenderEmail", expected.SenderEmail, actual.SenderEmail)
	this.AssertEquals(message + ": incorrect SenderName", expected.SenderName, actual.SenderName)
}

func TestRenderTemplates(t *testing.T) {
	test := Test{t}

	recipient, err := dehydrateRecipient("{\"email\": \"frhuw@def.de\", \"firstName\": \"Test\", \"lastName\": \"Fun\", \"language\": \"en\"}")
	test.AssertIsNil("error parsing recipient", err)
	templates := recipientsRepository.TemplatesString{
		SubjectTemplate: "Subject {{.firstName}}",
		BodyTemplate: `Body {{.firstName}}
			<a {{.link1}}>link1</a>
			<a {{.link2}}>link2</a>
		`,
		ReceiverEmailTemplate: "ReceiverEmail {{.firstName}}",
		ReceiverNameTemplate: "ReceiverName {{.firstName}}",
		SenderEmailTemplate: "SenderEmail {{.firstName}}",
		SenderNameTemplate: "SenderName {{.firstName}}",
		LinkTemplates: map[string]recipientsRepository.LinkTemplate{
			"link1": recipientsRepository.LinkTemplate{
				EncryptionKey: "key",
				BaseLink: "http://www.example.net/link1",
				Parameters: []string{"firstName", "lastName"},
			},
			"link2": recipientsRepository.LinkTemplate{
				EncryptionKey: "key",
				BaseLink: "http://www.example.net/link2?someParam=someValue",
				Parameters: []string{"email"},
			},
		},
	}
	parsedTemplate, err := templates.Parse()
	test.AssertIsNil("error during parsing", err)
	actual, err := RenderTemplates(parsedTemplate, recipient)
	test.AssertIsNil("error during rendering", err)

	expected := RenderedTemplates{
		Subject: "Subject Test",
		Body: `Body Test
			<a href="http://www.example.net/link1?firstName=Test&lastName=Fun&hmac=887be55bb72974423e601eb114fd0680858ca997">link1</a>
			<a href="http://www.example.net/link2?someParam=someValue&email=frhuw@def.de&hmac=8e614ea8d4411db4475ebc75900c2173a5c7deb3">link2</a>
		`,
		ReceiverEmail: "ReceiverEmail Test",
		ReceiverName: "ReceiverName Test",
		SenderEmail: "SenderEmail Test",
		SenderName: "SenderName Test",
	}
	test.AssertRenderedTemplatesEquals("incorrect result", expected, actual)
}

func dehydrateRecipient(recipient string) (dehydratedRecipient map[string]interface{}, err error) {
	err = json.Unmarshal([]byte(recipient), &dehydratedRecipient)
	return
}