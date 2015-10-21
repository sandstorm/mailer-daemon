package recipientsRepository

func getRecipientsRemainingKey(jobId string) string {
	return jobId + ".recipients.remaining"
}

func getRecipientsSendingKey(jobId string) string {
	return jobId + ".recipients.sending"
}

func getRecipientsFailedKey(jobId string) string {
	return jobId + ".recipients.sendingFailed"
}

// must be in sync with LUA script in function PopRemaining
func getRecipientLeaseKey(jobId, recipient string) string {
	return jobId + ".leases." + recipient
}

func getRecipientsLeaseKeyPrefix(jobId string) string {
	return jobId + ".leases."
}

func getJobStatusKey(jobId string) string {
	return getJobStateKey(jobId) + ".status"
}

func getJobTemplateKey(jobId string) string {
	return getJobStateKey(jobId) + ".htmlTemplate"
}

func getJobStatusMessageKey(jobId string) string {
	return getJobStateKey(jobId) + ".message"
}

func getJobNumberOfSentMailsKey(jobId string) string {
	return getJobStateKey(jobId) + ".numberOfSentMails"
}

func getJobNumberOfSendingFailuresKey(jobId string) string {
	return getJobStateKey(jobId) + ".numberOfSendingFailures"
}

func getJobNumberOfRecipientsKey(jobId string) string {
	return getJobStateKey(jobId) + ".numberOfRecipients"
}

func getJobStateKey(jobId string) string {
	return jobId + ".jobState"
}

func getOpenJobsKey() string {
	return "openJobs"
}

func getJobsKey() string {
	return "jobs"
}
