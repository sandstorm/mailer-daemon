package worker

import (
	"net/smtp"
	"gopkg.in/gomail.v1"
	"net/http"
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
)

type EmailGateway interface {
	Send(message RenderedTemplates) *EmailGatewayError
	Description() string
}

type EmailGatewayError struct {
	Cause error
	IsConnectionError bool
}

func (this *EmailGatewayError) Error() string {
	return this.Cause.Error()
}

type SmtpEmailGateway struct {
	SmtpUrl string
	SmtpAuth smtp.Auth
}

func (this *SmtpEmailGateway) Send(message RenderedTemplates) (*EmailGatewayError) {
	msg := gomail.NewMessage()
	msg.SetAddressHeader("From", message.SenderEmail, message.SenderName)
	msg.SetAddressHeader("To", message.ReceiverEmail, message.ReceiverName)
	msg.SetHeader("Reply-To", message.ReplyToEmail)
	msg.SetHeader("Subject", message.Subject)
	msg.SetBody("text/html", message.Body)

	mailer := gomail.NewCustomMailer(this.SmtpUrl, this.SmtpAuth)
	err := mailer.Send(msg)
	if err != nil {
		return &EmailGatewayError{
			Cause: err,
			IsConnectionError: strings.Contains(err.Error(), "connection refused"),
		}
	}
	return nil
}

func (this *SmtpEmailGateway) Description() string {
	return fmt.Sprintf("SMTP at '%s'", this.SmtpUrl)
}

type MandrillGateway struct {
	ApiKey string
}

type jsonObject map[string]interface{}

func (this *MandrillGateway) Send(message RenderedTemplates) (*EmailGatewayError) {
	payload := jsonObject{
		"key": this.ApiKey,
		"async": false,
		"message": jsonObject{
			"html": message.Body,
			"subject": message.Subject,
			"from_email": message.SenderEmail,
			"from_name": message.SenderEmail,
			"to": []jsonObject{
				jsonObject{
					"email": message.ReceiverEmail,
					"name": message.ReceiverName,
					"type": "to",
				},
			},
		},
	}
	err := this.sendJson(payload)
	if err != nil {
		return &EmailGatewayError{
			Cause: err,
			IsConnectionError: strings.Contains(err.Error(), "connection refused"),
		}
	}
	return nil
}

func (this *MandrillGateway) sendJson(payload jsonObject) (err error) {
	url := "https://mandrillapp.com/api/1.0/messages/send.json"

	json, err := json.Marshal(payload)
	if err != nil { return }

	request, err := http.NewRequest("POST", url, bytes.NewBuffer(json))
	if err != nil { return }
	request.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil { return }
	defer response.Body.Close()

	status, err := this.parseJsonBody(response)
	if err != nil { return }

	if response.StatusCode >= 300 || status["status"] != "sent" {
		err = &WorkerError{
			fmt.Sprintf("%+v", status),
			nil,
		}
	}
	return
}

func (this *MandrillGateway) parseJsonBody(response *http.Response) (value jsonObject, err error) {
	buffer := new(bytes.Buffer)
	buffer.ReadFrom(response.Body)
	bytes := buffer.Bytes()

	if response.StatusCode < 300 {
		// we get an array of JSON objects containing one object per
		// sent mail (in our case always one)
		var values []jsonObject
		err = json.Unmarshal(bytes, &values)
		if err != nil { return }
		value = values[0]
	} else {
		// if an error occurs we do not want to depend on any structure of the
		// response body as long as it is JSON
		var body interface{}
		err = json.Unmarshal(bytes, &body)
		if err != nil { return }
		value = jsonObject{
			"error": body,
		}
	}

	return
}

func (this *MandrillGateway) Description() string {
	return fmt.Sprintf("Mandrill")
}
