package worker

import (
	"testing"
	"github.com/sandstorm/mailer-daemon/recipientsRepository"
	"encoding/json"
	"strings"
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

func TestRenderTemplatesWithComments(t *testing.T) {
	test := Test{t}

	recipient, err := dehydrateRecipient("{\"email\": \"frhuw@def.de\", \"firstName\": \"Test\", \"lastName\": \"Fun\", \"language\": \"en\", \"bookingNr\": \"Event-123\"}")

	var mjmlBodyTemplateWithoutPlaceholders = `
	<!doctype html>
<html xmlns="http://www.w3.org/1999/xhtml" xmlns:v="urn:schemas-microsoft-com:vml" xmlns:o="urn:schemas-microsoft-com:office:office">

<head>
  <title></title>
  <!--[if !mso]><!-- {{.firstName}} -->
  <meta http-equiv="X-UA-Compatible" content="IE=edge">
  <!--<![endif]-->
  <meta http-equiv="Content-Type" content="text/html; charset=UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <style type="text/css">
    #outlook a {
      padding: 0;
    }

    .ReadMsgBody {
      width: 100%;
    }

    .ExternalClass {
      width: 100%;
    }

    .ExternalClass * {
      line-height: 100%;
    }

    body {
      margin: 0;
      padding: 0;
      -webkit-text-size-adjust: 100%;
      -ms-text-size-adjust: 100%;
    }

    table,
    td {
      border-collapse: collapse;
      mso-table-lspace: 0pt;
      mso-table-rspace: 0pt;
    }

    img {
      border: 0;
      height: auto;
      line-height: 100%;
      outline: none;
      text-decoration: none;
      -ms-interpolation-mode: bicubic;
    }

    p {
      display: block;
      margin: 13px 0;
    }
  </style>
  <!--[if !mso]><!-->
  <style type="text/css">
    @media only screen and (max-width:480px) {
      @-ms-viewport {
        width: 320px;
      }
      @viewport {
        width: 320px;
      }
    }
  </style>
  <!--<![endif]-->
  <!--[if mso]>
<xml>
  <o:OfficeDocumentSettings>
    <o:AllowPNG/>
    <o:PixelsPerInch>96</o:PixelsPerInch>
  </o:OfficeDocumentSettings>
</xml>
<![endif]-->
  <!--[if lte mso 11]>
<style type="text/css">
  .outlook-group-fix {
    width:100% !important;
  }
</style>
<![endif]-->
  <style type="text/css">
    @media only screen and (min-width:480px) {
      .mj-column-per-100 {
        width: 100%!important;
      }
    }
  </style>
</head>

<body>

  <div class="mj-container">
    <!--[if mso | IE]>
      <table role="presentation" border="0" cellpadding="0" cellspacing="0" width="600" align="center" style="width:600px;">
        <tr>
          <td style="line-height:0px;font-size:0px;mso-line-height-rule:exactly;">
      <![endif]-->
    <div style="margin:0px auto;max-width:600px;">
      <table role="presentation" cellpadding="0" cellspacing="0" style="font-size:0px;width:100%;" align="center" border="0">
        <tbody>
          <tr>
            <td style="text-align:center;vertical-align:top;direction:ltr;font-size:0px;padding:20px 0px;">
              <!--[if mso | IE]>
      <table role="presentation" border="0" cellpadding="0" cellspacing="0">
        <tr>
          <td style="vertical-align:top;width:600px;">
      <![endif]-->
              <div class="mj-column-per-100 outlook-group-fix" style="vertical-align:top;display:inline-block;direction:ltr;font-size:13px;text-align:left;width:100%;">
                <table role="presentation" cellpadding="0" cellspacing="0" width="100%" border="0">
                  <tbody>
                    <tr>
                      <td style="word-wrap:break-word;font-size:0px;padding:10px 25px;" align="center">
                        <table role="presentation" cellpadding="0" cellspacing="0" style="border-collapse:collapse;border-spacing:0px;" align="center" border="0">
                          <tbody>
                            <tr>
                              <td style="width:100px;"><img alt="" title="" height="auto" src="/assets/img/logo-small.png" style="border:none;border-radius:0px;display:block;font-size:13px;outline:none;text-decoration:none;width:100%;height:auto;" width="100"></td>
                            </tr>
                          </tbody>
                        </table>
                      </td>
                    </tr>
                    <tr>
                      <td style="word-wrap:break-word;font-size:0px;padding:10px 25px;">
                        <p style="font-size:1px;margin:0px auto;border-top:4px solid #F45E43;width:100%;"></p>
                        <!--[if mso | IE]><table role="presentation" align="center" border="0" cellpadding="0" cellspacing="0" style="font-size:1px;margin:0px auto;border-top:4px solid #F45E43;width:100%;" width="600"><tr><td style="height:0;line-height:0;"> </td></tr></table><![endif]-->
                      </td>
                    </tr>
                    <tr>
                      <td style="word-wrap:break-word;font-size:0px;padding:10px 25px;" align="left">
                        <div style="cursor:auto;color:#F45E43;font-family:helvetica;font-size:20px;line-height:22px;text-align:left;">Hello World</div>
                      </td>
                    </tr>
                  </tbody>
                </table>
              </div>
              <!--[if mso | IE]>
      </td></tr></table>
      <![endif]-->
            </td>
          </tr>
        </tbody>
      </table>
    </div>
    <!--[if mso | IE]>
      </td></tr></table>
      <![endif]-->
  </div>
</body>

</html>`;

	test.AssertIsNil("error parsing recipient", err)
	templates := recipientsRepository.TemplatesString{
		SubjectTemplate: "Subject {{.firstName}}",
		BodyTemplate: mjmlBodyTemplateWithoutPlaceholders,
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
		Body: strings.Replace(mjmlBodyTemplateWithoutPlaceholders,"{{.firstName}}", "Test", 1),
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