package recipientsRepository

import (
	"fmt"
	"github.com/garyburd/redigo/redis"
	"strconv"
	"encoding/json"
	"os"
	"bufio"
)

type RedisRepository struct {
	Pool *redis.Pool
}

func (this *RedisRepository) GetRandomOpenJob() (jobId string, err error) {
	connection := this.Pool.Get()
	defer connection.Close()
	response, err := connection.Do("SRANDMEMBER", getOpenJobsKey())
	jobId = this.responseToString(response)
	return
}

func (this *RedisRepository) CreateJob(jobId string, htmlTemplate string) error {
	connection := this.Pool.Get()
	defer connection.Close()

	luaScript := `
		local jobsKey = KEYS[1];
		local jobStatusKey = KEYS[2];
		local jobStateHtmlTemplateKey = KEYS[3];

		local jobId = ARGV[1];
		local htmlTemplate = ARGV[2];

		local jobExists = redis.call('EXISTS', jobStatusKey);
		if jobExists == 1 then
			error('job already exists');
		else
			redis.call('SADD', jobsKey, jobId);
			redis.call('SET', jobStatusKey, 'preparing');
			redis.call('SET', jobStateHtmlTemplateKey, htmlTemplate);
		end
	`
	_, err := connection.Do("EVAL", luaScript, 3,
		getJobsKey(),
		getJobStatusKey(jobId),
		getJobTemplateKey(jobId),
		jobId,
		htmlTemplate,
	)
	return err
}

func (this *RedisRepository) AbortAndRemoveJob(jobId string) error {
	connection := this.Pool.Get()
	defer connection.Close()

	luaScript := `
		local jobsKey = KEYS[1];
		local openJobsKey = KEYS[2];

		local jobId = ARGV[1];

		redis.call('SREM', openJobsKey, jobId);
		redis.call('SREM', jobsKey, jobId);

		for i = 3, #KEYS do
			redis.call('DEL', KEYS[i]);
		end
	`
	_, err := connection.Do("EVAL", luaScript, 12,
		getJobsKey(),
		getOpenJobsKey(),
		getRecipientsRemainingKey(jobId),
		getRecipientsSendingKey(jobId),
		getRecipientsFailedKey(jobId),
		getJobStatusKey(jobId),
		getJobTemplateKey(jobId),
		getJobStatusMessageKey(jobId),
		getJobNumberOfSentMailsKey(jobId),
		getJobNumberOfSendingFailuresKey(jobId),
		getJobNumberOfRecipientsKey(jobId),
		getJobStateKey(jobId),
		jobId,
	)
	return err
}

func (this *RedisRepository) GetJobStatus(jobIds []string) (status JobStatus, err error) {
	connection := this.Pool.Get()
	defer connection.Close()

	status = JobStatus{
		Jobs: map[string]IndividualJobStatus{},
		Summary: IndividualJobStatus{
			Status: "summary",
		},
	}
	err = nil

	// get status for individual jobs
	for _, jobId := range jobIds {
		_, exists := status.Jobs[jobId]
		if exists { continue }

		jobStatus := IndividualJobStatus{}
		var response interface{}

		response, err = connection.Do("GET", getJobStatusKey(jobId))
		jobStatus.Status = this.responseToString(response)
		if err != nil { return }

		response, err = connection.Do("GET", getJobStatusMessageKey(jobId))
		jobStatus.Message = this.responseToString(response)
		if err != nil { return }

		response, err = connection.Do("GET", getJobNumberOfRecipientsKey(jobId))
		if err != nil { return }
		jobStatus.NumberOfRecipients, err = this.responseToInt(response)
		if err != nil { return }

		response, err = connection.Do("GET", getJobNumberOfSentMailsKey(jobId))
		if err != nil { return }
		jobStatus.NumberOfSentMails, err = this.responseToInt(response)
		if err != nil { return }

		response, err = connection.Do("GET", getJobNumberOfSendingFailuresKey(jobId))
		if err != nil { return }
		jobStatus.NumberOfSendingFailures, err = this.responseToInt(response)
		if err != nil { return }

		response, err = connection.Do("SCARD", getRecipientsSendingKey(jobId))
		if err != nil { return }
		jobStatus.NumberOfCurrentlySendingMails = response.(int64)
		if err != nil { return }

		status.Jobs[jobId] = jobStatus
	}

	// sum up
	for _, jobStatus := range status.Jobs {
		status.Summary.NumberOfRecipients += jobStatus.NumberOfRecipients
		status.Summary.NumberOfSentMails += jobStatus.NumberOfSentMails
		status.Summary.NumberOfSendingFailures += jobStatus.NumberOfSendingFailures
		status.Summary.NumberOfCurrentlySendingMails += jobStatus.NumberOfCurrentlySendingMails
	}

	return
}

func (this *RedisRepository) WriteSendingFailuresToFile(targetFile string, jobIds []string) (err error) {
	file, err := os.Create(targetFile)
	if err != nil { return }
	defer file.Close()

	writer := bufio.NewWriter(file)
	defer writer.Flush()

	for _, jobId := range jobIds {
		err = this.appendSendingFailuresToFile(writer, jobId)
		if err != nil { return }
	}
	return
}

func (this *RedisRepository) appendSendingFailuresToFile(targetFile *bufio.Writer, jobId string) (err error) {
	size := 50000
	lastResult := size

	for offset := 0; lastResult == size; offset += size {
		failures, err := this.getSendingFailures(jobId, offset, size)
		if err != nil { return err }

		lastResult = len(failures)
		for _, failure := range failures {
			targetFile.WriteString(fmt.Sprintf("%s	%s\n", failure.Recipient, failure.Cause))
		}
	}

	return
}

func (this *RedisRepository) getSendingFailures(jobId string, offset, size int) (failures []MailSendingFailure, err error) {
	connection := this.Pool.Get()
	defer connection.Close()

	response, err := connection.Do("LRANGE", getRecipientsFailedKey(jobId), offset, offset + size - 1)
	if err != nil { return }

	failuresJson := this.responseToSlice(response)
	failures = make([]MailSendingFailure, len(failuresJson))
	for index, failureJson := range failuresJson {
		var failure MailSendingFailure
		err = json.Unmarshal(failureJson.([]byte), &failure)
		if err != nil { return }

		failures[index] = failure
	}
	return
}

func (this *RedisRepository) FinishPreparation(jobId string, numberOfRecipients int) error {
	connection := this.Pool.Get()
	defer connection.Close()

	luaScript := `
			local openJobsKey = KEYS[1];
			local jobStatusKey = KEYS[2];
			local numberOfRecipientsKey = KEYS[3];

			local jobId = ARGV[1];
			local totalCount = ARGV[2];

			if totalCount == '0' then
				redis.call('SET', jobStatusKey, 'done');
			else
				redis.call('SET', jobStatusKey, 'prepared');
				redis.call('SADD', openJobsKey, jobId);
			end
			redis.call('SET', numberOfRecipientsKey, totalCount);
		`
	_, err := connection.Do("EVAL", luaScript, 3,
		getOpenJobsKey(),
		getJobStatusKey(jobId),
		getJobNumberOfRecipientsKey(jobId),
		jobId,
		numberOfRecipients,
	)
	return err
}

func (this *RedisRepository) MarkJobAsFailed(jobId, reason string) error {
	connection := this.Pool.Get()
	defer connection.Close()

	luaScript := `
		local openJobsKey = KEYS[1];
		local jobStatusKey = KEYS[2];
		local jobStatusMessageKey = KEYS[3];

		local jobId = ARGV[1];
		local jobStatusMessage = ARGV[2];

		redis.call('SET', jobStatusKey, 'failed');
		redis.call('SET', jobStatusMessageKey, jobStatusMessage);
		redis.call('SREM', openJobsKey, jobId);
	`
	_, err := connection.Do("EVAL", luaScript, 3,
		getOpenJobsKey(),
		getJobStatusKey(jobId),
		getJobStatusMessageKey(jobId),
		jobId,
		reason,
	)
	return err
}

func (this *RedisRepository) AddRecipient(jobId, recipient string) error {
	connection := this.Pool.Get()
	defer connection.Close()

	_, err := connection.Do("RPUSH", getRecipientsRemainingKey(jobId), recipient)
	return err
}

func (this *RedisRepository) GetNextOpenRecipient(jobId string) (recipient string, err error) {
	connection := this.Pool.Get()
	defer connection.Close()

	luaScript := `
		local remainingKey = KEYS[1];
		local sendingKey = KEYS[2];
		local leaseKeyPrefix = KEYS[3];

		local recipient = redis.call('LPOP', remainingKey);
		if recipient then
			redis.call('SADD', sendingKey, recipient);
			redis.call('SETEX', leaseKeyPrefix .. recipient, 30, 1);
			return recipient
		else
			-- try to re-use a recipient if its lease is expired
			local sending = redis.call('SMEMBERS', sendingKey);
			for index, recipient in pairs(sending) do
				local lease = redis.call('GET', leaseKeyPrefix .. recipient);
				if not lease then
					redis.call('SETEX', leaseKeyPrefix .. recipient, 30, 1);
					return recipient;
				end
			end
		end
		return nil;
	`
	response, err := connection.Do("EVAL", luaScript, 3,
		getRecipientsRemainingKey(jobId),
		getRecipientsSendingKey(jobId),
		getRecipientsLeaseKeyPrefix(jobId),
	)
	recipient = this.responseToString(response)
	return
}

func (this *RedisRepository) ExtendLease(jobId, recipient string) error {
	connection := this.Pool.Get()
	defer connection.Close()

	_, err := connection.Do("SETEX", getRecipientLeaseKey(jobId, recipient), 30, 1)
	return err
}

func (this *RedisRepository) MarkRecipientAsDone(jobId, recipient string) error {
	return this.markRecipientAsProcessed(jobId, recipient, "")
}

func (this *RedisRepository) MarkRecipientAsFailed(jobId, recipient, recipientId string, cause error) error {
	sendingFailure := MailSendingFailure{
		Recipient: recipientId,
		Cause: cause.Error(),
	}
	sendingFailureJson, err := json.Marshal(sendingFailure)
	if err != nil { return err }

	return this.markRecipientAsProcessed(jobId, recipient, string(sendingFailureJson))
}

func (this *RedisRepository) markRecipientAsProcessed(jobId, recipient, sendingFailureJson string) error {
	connection := this.Pool.Get()
	defer connection.Close()

	luaScript := `
		local openJobsKey = KEYS[1];
		local jobStatusKey = KEYS[2];
		local sendingKey = KEYS[3];
		local leaseKey = KEYS[4];
		local numberOfSentKey = KEYS[5];
		local numberOfRecipientsKey = KEYS[6];
		local numberOfFailedKey = KEYS[7];
		local sendingFailuresKey = KEYS[8];

		local jobId = ARGV[1];
		local recipient = ARGV[2];
		local sendingFailure = ARGV[3];

		redis.call('SREM', sendingKey, recipient);
		redis.call('DEL', leaseKey);

		local sentDelta = 1;
		local failedDelta = 0;
		if sendingFailure ~= '' then
			sentDelta = 0;
			failedDelta = 1;
			redis.call('RPUSH', sendingFailuresKey, sendingFailure)
		end

		local sentCount =  redis.call('INCRBY', numberOfSentKey, sentDelta);
		local failedCount = redis.call('INCRBY', numberOfFailedKey, failedDelta);
		-- we have to get this value as an integer for comparison
		local totalCount = redis.call('INCRBY', numberOfRecipientsKey, 0);
		if totalCount == sentCount + failedCount then
			redis.call('SREM', openJobsKey, jobId);
			redis.call('SET', jobStatusKey, 'done');
		end
	`
	_, err := connection.Do("EVAL", luaScript, 8,
		getOpenJobsKey(),
		getJobStatusKey(jobId),
		getRecipientsSendingKey(jobId),
		getRecipientLeaseKey(jobId, recipient),
		getJobNumberOfSentMailsKey(jobId),
		getJobNumberOfRecipientsKey(jobId),
		getJobNumberOfSendingFailuresKey(jobId),
		getRecipientsFailedKey(jobId),
		jobId,
		recipient,
		sendingFailureJson,
	)
	return err
}

func (this *RedisRepository) GetTemplates(jobId string) (templates Templates, err error) {
	connection := this.Pool.Get()
	defer connection.Close()

	response, err := connection.Do("GET", getJobTemplateKey(jobId))
	if err != nil { return }
	templatesJson := fmt.Sprintf("%s", response)
	var templatesString TemplatesString
	err = json.Unmarshal([]byte(templatesJson), &templatesString)
	if err != nil { return }

	templates, err = templatesString.Parse()
	return
}

func (this *RedisRepository) responseToInt(response interface{}) (value int, err error) {
	if response == nil {
		return 0, nil
	} else {
		return strconv.Atoi(this.responseToString(response))
	}
}

func (this *RedisRepository) responseToString(response interface{}) string {
	if response == nil {
		return ""
	} else {
		return string(response.([]byte))
	}
}

func (this *RedisRepository) responseToSlice(response interface{}) (result []interface{}) {
	if response == nil {
		return []interface{}{}
	} else {
		return response.([]interface{})
	}
}
