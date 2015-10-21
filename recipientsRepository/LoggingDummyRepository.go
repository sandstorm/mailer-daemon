package recipientsRepository

import (
	"log"
)

type LoggingDummyRepository struct {
	// storage to use after logging
	Repository Repository
}

func (this *LoggingDummyRepository) GetRandomOpenJob() (jobId string, err error) {
	log.Println("GetRandomOpenJob()")
	return this.Repository.GetRandomOpenJob()
}

func (this *LoggingDummyRepository) CreateJob(jobId string, htmlTemplate string) error {
	log.Println("CreateJob(", jobId, ", htmlTemplate length = ", len(htmlTemplate), ")")
	return this.Repository.CreateJob(jobId, htmlTemplate);
}

func (this *LoggingDummyRepository) AbortAndRemoveJob(jobId string) error {
	log.Printf("AbortAndDeleteJob(%s)", jobId)
	return this.Repository.AbortAndRemoveJob(jobId);
}

func (this *LoggingDummyRepository) GetJobStatus(jobIds []string) (status JobStatus, err error) {
	log.Printf("GetJobStatus(%s)", jobIds)
	return this.Repository.GetJobStatus(jobIds);
}

func (this *LoggingDummyRepository) WriteSendingFailuresToFile(targetFile string, jobIds []string) (err error) {
	log.Println("WriteSendingFailuresToFile(", targetFile, ",", jobIds, ")")
	return this.Repository.WriteSendingFailuresToFile(targetFile, jobIds);
}

func (this *LoggingDummyRepository) FinishPreparation(jobId string, numberOfRecipients int) error {
	log.Println("FinishPreparation(", jobId, ",", numberOfRecipients, ")")
	return this.Repository.FinishPreparation(jobId, numberOfRecipients);
}

func (this *LoggingDummyRepository) MarkJobAsFailed(jobId, reason string) error {
	log.Println("MarkJobAsFailed(", jobId, ",", reason, ")")
	return this.Repository.MarkJobAsFailed(jobId, reason);
}

func (this *LoggingDummyRepository) AddRecipient(jobId, recipient string) error {
	log.Println("AddRecipient(", jobId, ",", recipient, ")")
	return this.Repository.AddRecipient(jobId, recipient);
}

func (this *LoggingDummyRepository) GetNextOpenRecipient(jobId string) (recipient string, err error) {
	log.Println("GetNextOpenRecipient(", jobId, ")")
	return this.Repository.GetNextOpenRecipient(jobId);
}

func (this *LoggingDummyRepository) ExtendLease(jobId, recipient string) error {
	log.Println("ExtendLease(", jobId, ",", recipient, ")")
	return this.Repository.ExtendLease(jobId, recipient);
}

func (this *LoggingDummyRepository) MarkRecipientAsDone(jobId, recipient string) error {
	log.Println("MarkRecipientAsDone(", jobId, ",", recipient, ")")
	return this.Repository.MarkRecipientAsDone(jobId, recipient);
}

func (this *LoggingDummyRepository) MarkRecipientAsFailed(jobId, recipient, recipientId string, cause error) error {
	log.Println("MarkRecipientAsFailed(", jobId, ",", recipient, ",", recipientId, ",", cause, ")")
	return this.Repository.MarkRecipientAsFailed(jobId, recipient, recipientId, cause);
}

func (this *LoggingDummyRepository) GetTemplates(jobId string) (templates Templates, err error) {
	log.Println("GetTemplates(", jobId, ")")
	return this.Repository.GetTemplates(jobId);
}