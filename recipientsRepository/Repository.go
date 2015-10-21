package recipientsRepository

type MailSendingFailure struct {
	Recipient string
	Cause string
}

type IndividualJobStatus struct {
	Status string
	Message string
	NumberOfRecipients int
	NumberOfSentMails int
	NumberOfSendingFailures int
	NumberOfCurrentlySendingMails int64
}

type JobStatus struct {
	Jobs map[string]IndividualJobStatus
	Summary IndividualJobStatus
}

type LeaseExtensionRepository interface {
	/*
	 * Extends the lease of the given already popped recipient.
	 */
	ExtendLease(jobId, recipient string) error
}

type Repository interface {
	LeaseExtensionRepository

	GetRandomOpenJob() (jobId string, err error)

	/*
	 * Creates a new job with the given jobId or returns an
	 * error if this job already exists.
	 */
	CreateJob(jobId string, templates string) error

	AbortAndRemoveJob(jobId string) error

	GetJobStatus(jobIds []string) (status JobStatus, err error)

	WriteSendingFailuresToFile(targetFile string, jobIds []string) (err error)

	/**
	 * marks the job as fully prepared and sets the total number of recipients
	 */
	FinishPreparation(jobId string, numberOfRecipients int) error

	MarkJobAsFailed(jobId, reason string) error

	AddRecipient(jobId, recipient string) error

	/*
	 * Pops the next remaining recipient for the given job id.
	 *
	 * see ExtendLease
	 * see RemoveRemaining
	 */
	GetNextOpenRecipient(jobId string) (recipient string, err error)

	MarkRecipientAsDone(jobId, recipient string) error

	MarkRecipientAsFailed(jobId, recipient, recipientId string, cause error) error

	GetTemplates(jobId string) (templates Templates, err error)
}



