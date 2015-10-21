package router

import (
	"fmt"
	"net/http"
	"github.com/gocraft/web"
	"sandstormmedia/project-webessentials-mailer/go/mailer/recipientsRepository"
	"encoding/json"
)

type RequestHandler struct {
	Repository recipientsRepository.Repository
	ServerConfiguration ServerConfiguration
}

func (this *RequestHandler) HandleStatus(response web.ResponseWriter, request *web.Request) {
	this.handleRequestWithPossibleErrors(response, request, this.getJobStatus)
}

func (this *RequestHandler) HandleSendingFailures(response web.ResponseWriter, request *web.Request) {
	this.handleRequestWithPossibleErrors(response, request, this.writeSendingFailuresToFile)
}

func (this *RequestHandler) HandleSend(response web.ResponseWriter, request *web.Request) {
	this.handleRequestWithPossibleErrors(response, request, this.startRecipientsImportJob)
}

func (this *RequestHandler) HandleAbortAndRemove(response web.ResponseWriter, request *web.Request) {
	this.handleRequestWithPossibleErrors(response, request, this.abortAndRemove)
}

func (this *RequestHandler) HandleServerConfiguration(response web.ResponseWriter, request *web.Request) {
	this.handleRequestWithPossibleErrors(response, request, this.getServerConfiguration)
}

func (this *RequestHandler) handleRequestWithPossibleErrors(response web.ResponseWriter, request *web.Request, handler func(web.ResponseWriter, *web.Request) error) {
	response.Header().Add("Access-Control-Allow-Origin", "*")
	err := handler(response, request)
	if (err != nil) {
		response.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(response, err.Error())
	}
}

func (this *RequestHandler) getJobStatus(response web.ResponseWriter, request *web.Request) error {
	payload, err := GetStatusRequestPayloadFromRequest(request)
	if err != nil {
		return err
	}
	status, err := this.Repository.GetJobStatus(payload.JobIds)
	if err != nil {
		return err
	}
	json, err := json.Marshal(status)
	if err != nil {
		return err
	}
	response.Write(json)
	return nil
}

func (this *RequestHandler) writeSendingFailuresToFile(response web.ResponseWriter, request *web.Request) (err error) {
	targetFile := request.FormValue("targetFile")
	if targetFile == "" { return &ServerError{"targetFile must not be empty"} }

	jobIds, err := GetJobIdsFrom(request)
	if err != nil { return }

	err = this.Repository.WriteSendingFailuresToFile(targetFile, jobIds)
	return
}

func (this *RequestHandler) startRecipientsImportJob(response web.ResponseWriter, request *web.Request) error {
	payload, err := GetSendRequestPayloadFromRequest(request)
	if (err == nil) {
		blacklist, err := CreateBlacklistFromFile(payload.Blacklist)
		if err != nil { return err }
		job := RecipientListImporterJob{
			Target: this.Repository,
			SourceRecipientsList: payload.RecipientsList,
			Key: request.PathParams["id"],
			Templates: payload.Templates,
			Blacklist: blacklist,
		}
		err = job.Prepare()
		if err != nil { return err }
		go job.ExecuteAndLogErrors()
	} else {
		return err
	}
	return nil
}

func (this *RequestHandler) abortAndRemove(response web.ResponseWriter, request *web.Request) error {
	jobId := request.PathParams["id"]
	return this.Repository.AbortAndRemoveJob(jobId)
}

func (this *RequestHandler) getServerConfiguration(response web.ResponseWriter, request *web.Request) error {
	json, err := json.Marshal(this.ServerConfiguration)
	if err != nil { return err }
	response.Write(json)
	return nil
}

