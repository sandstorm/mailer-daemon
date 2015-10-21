package router

import (
	"bytes"
	"strings"
	"github.com/gocraft/web"
	"encoding/json"
	"sandstormmedia/project-webessentials-mailer/go/mailer/recipientsRepository"
	"os"
)

type SendRequestPayload struct {
	RecipientsList string
	Blacklist string
	Templates recipientsRepository.TemplatesString
}

func GetSendRequestPayloadFromRequest(request *web.Request) (payload SendRequestPayload, err error) {
	payload = SendRequestPayload{}
	err = getRequestPayloadFromJsonBody(request, &payload)
	if err != nil { return payload, &ServerError{"Failed to parse JSON payload: " + err.Error()} }

	_, err = payload.Templates.Parse()
	if err != nil { return payload, &ServerError{"Failed to compile templates: " + err.Error()} }

	if payload.Templates.SenderEmailTemplate == "" {
		return payload, &ServerError{"SenderEmailTemplate must not be empty"}
	}

	if payload.Templates.ReceiverEmailTemplate == "" {
		return payload, &ServerError{"ReceiverEmailTemplate must not be empty"}
	}

	_, err = os.Stat(payload.RecipientsList)
	if err != nil { return payload, &ServerError{"Failed to find recipient list: " + err.Error()} }

	if payload.Blacklist != "" {
		_, err = os.Stat(payload.Blacklist)
		if err != nil { return payload, &ServerError{"Failed to find blacklist: " + err.Error()} }
	}

	return
}

func getRequestPayloadFromJsonBody(request *web.Request, payload interface{}) error {
	buffer := new(bytes.Buffer)
	buffer.ReadFrom(request.Body)
	bytes := buffer.Bytes()
	return json.Unmarshal(bytes, payload)
}

type StatusRequestPayload struct {
	JobIds []string
}

func GetStatusRequestPayloadFromRequest(request *web.Request) (payload StatusRequestPayload, err error) {
	payload = StatusRequestPayload{}
	payload.JobIds, err = GetJobIdsFrom(request)
	return
}

func GetJobIdsFrom(request *web.Request) (jobIds []string, err error) {
	jobIdsString := request.FormValue("jobIds")
	if jobIdsString == "" {
		err = &ServerError{"Missing path parameter 'jobIds', e.g. ?jobIds=id1,id2,my-big-job"}
		return
	}
	jobIds = strings.Split(jobIdsString, ",")
	return
}