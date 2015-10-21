package worker
import (
	"log"
	"sandstormmedia/project-webessentials-mailer/go/mailer/recipientsRepository"
	"time"
	"fmt"
	"encoding/json"
)

type EmailSender struct {
	Gateway EmailGateway
	Repository recipientsRepository.Repository
	workerStopChannel chan bool
}

func (this *EmailSender) Start() {
	this.workerStopChannel = make(chan bool)
	go this.worker()
}

func (this *EmailSender) Stop() {
	this.workerStopChannel <- true
}

func (this *EmailSender) worker() {
	for {
		select {
		case <-this.workerStopChannel:
			return
		default:
			this.getAndProcessNextJob()
		}
	}
}

func (this *EmailSender) getAndProcessNextJob() error {
	jobId, err := this.getNextOpenJob()
	if (jobId == "" || err != nil) { return this.waitAndLog(err) }

	templates, err := this.getTemplates(jobId)
	if (err != nil) { return this.markJobAsFailed(jobId, err) }

	return this.processEntireJob(jobId, templates)
}

func (this *EmailSender) markJobAsFailed(jobId string, cause error) error {
	log.Printf("Job '%s' failed permanently: %s", jobId, cause.Error())
	err := this.Repository.MarkJobAsFailed(jobId, cause.Error())
	if err != nil {
		return this.waitAndLog(&WorkerError{
			Message: fmt.Sprintf("Failed to mark job '%s' as failed", jobId),
			Cause: err,
		})
	}
	return cause
}

func (this *EmailSender) waitAndLog(fatal error) error {
	if fatal != nil {
		log.Println("FATAL error in EmailSender. Going to wait for longer time: ", fatal)
		// wait long if error occurred while trying to access redis or such
		time.Sleep(10 * time.Second)
	} else {
		// wait short if there is not work
		time.Sleep(time.Second)
	}
	return fatal
}

func (this *EmailSender) getNextOpenJob() (jobId string, err error) {
	jobId, err = this.Repository.GetRandomOpenJob()
	if (err != nil) {
		err = &WorkerError{
			Message: "Failed to get next open job",
			Cause: err,
		}
	}
	return
}

func (this *EmailSender) getTemplates(jobId string) (templates recipientsRepository.Templates, err error) {
	templates, err = this.Repository.GetTemplates(jobId)
	if (err != nil) {
		err = &WorkerError{
			Message: fmt.Sprintf("Failed to load templates of job '%s'", jobId),
			Cause: err,
		}
	}
	return
}

func (this *EmailSender) processEntireJob(jobId string, templates recipientsRepository.Templates) (err error) {
	for recipient, err := this.getNextRecipient(jobId); recipient != ""; recipient, err = this.getNextRecipient(jobId) {
		// fatal error while loading recipient: abort job execution
		if (err != nil) { return this.waitAndLog(err) }

		// fatal error while processing recipient: abort job execution
		err = this.processRecipient(jobId, recipient, templates)
		if (err != nil) { return this.waitAndLog(err) }
	}
	return
}

func (this *EmailSender) getNextRecipient(jobId string) (recipient string, err error) {
	recipient, err = this.Repository.GetNextOpenRecipient(jobId)
	if (err != nil) {
		err = &WorkerError{
			Message: "Failed to get next receiver",
			Cause: err,
		}
	}
	return
}

func (this *EmailSender) processRecipient(jobId, recipient string, templates recipientsRepository.Templates) (fatal error) {
	leaseExtender := this.startLeaseExtender(jobId, recipient)
	defer leaseExtender.Stop()

	dehydratedRecipient, err := this.dehydrateRecipient(recipient)
	if (err != nil) { return this.markRecipientAsFailed(jobId, recipient, recipient, err) }
	email, ok := dehydratedRecipient["email"].(string)
	if (!ok) {
		return this.markRecipientAsFailed(jobId, recipient, recipient, &WorkerError{
			Message: "email of recipient is not a string",
		})
	}

	rendered, err := this.renderTemplates(dehydratedRecipient, templates)
	if (err != nil) { return this.markRecipientAsFailed(jobId, recipient, email, err) }

	gatewayErr := this.Gateway.Send(rendered)
	if (gatewayErr != nil) {
		if gatewayErr.IsConnectionError {
			return gatewayErr
		} else {
			return this.markRecipientAsFailed(jobId, recipient, email, gatewayErr)
		}
	}

	return this.markRecipientAsDone(jobId, recipient)
}

func (this *EmailSender) startLeaseExtender(jobId, recipient string) *LeaseExtender {
	leaseExtender := LeaseExtender{
		Repository: this.Repository,
		JobId: jobId,
		Receiver: recipient,
		ExtensionInterval: 15 * time.Second,
	}
	leaseExtender.Start()
	return &leaseExtender
}

func (this *EmailSender) dehydrateRecipient(recipient string) (dehydratedRecipient map[string]interface{}, err error) {
	err = json.Unmarshal([]byte(recipient), &dehydratedRecipient)
	if err != nil {
		err = &WorkerError{
			Message: "failed to unmarshal recipient",
			Cause: err,
		}
		return
	}

	if dehydratedRecipient["email"] == nil || dehydratedRecipient["email"] == "" {
		err = &WorkerError{
			Message: "recipient has no 'email'",
			Cause: nil,
		}
	}

	return
}

func (this *EmailSender) renderTemplates(recipient map[string]interface{}, templates recipientsRepository.Templates) (rendered RenderedTemplates, err error) {
	rendered, err = RenderTemplates(templates, recipient)
	if err != nil {
		err = &WorkerError{
			Message: fmt.Sprintf("Failed to render templates for '%s'", recipient),
			Cause: err,
		}
	}
	return
}

func (this *EmailSender) markRecipientAsDone(jobId, recipient string) (err error) {
	err = this.Repository.MarkRecipientAsDone(jobId, recipient)
	if (err != nil) {
		err = &WorkerError{
			Message: fmt.Sprintf("Failed to mark recipient '%s' as done", recipient),
			Cause: err,
		}
	}
	return
}

func (this *EmailSender) markRecipientAsFailed(jobId, recipient, recipientId string, cause error) (err error) {
	err = this.Repository.MarkRecipientAsFailed(jobId, recipient, recipientId, cause)
	if (err != nil) {
		err = &WorkerError{
			Message: "Failed to mark recipient as failed",
			Cause: err,
		}
	}
	return
}
