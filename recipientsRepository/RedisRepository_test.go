package recipientsRepository

import (
	"testing"
	"github.com/garyburd/redigo/redis"
	"time"
	"os/exec"
	"log"
	"strconv"
	"fmt"
	"io/ioutil"
	"os"
)

type Test struct {
	test *testing.T
}

func (this *Test) AssertIsNil(message string, actual interface{}) {
	if actual != nil {
		this.test.Errorf(message + ": actual should be nil, but is '%v'", actual)
	}
}

func (this *Test) AssertIsNotNil(message string, actual interface{}) {
	if actual == nil {
		this.test.Errorf(message + ": actual should not be nil")
	}
}

func (this *Test) AssertStringIsEmpty(message string, actual string) {
	if actual != "" {
		this.test.Errorf(message + ": actual should be an empty string, but is '%v'", actual)
	}
}

func (this *Test) AssertEquals(message string, expected, actual interface{}) {
	if actual != expected {
		this.test.Errorf(message + ": expected '%v' is different from actual '%v'  ", expected, actual)
	}
}

func (this *Test) AssertJobStatusEquals(message string, expected, actual JobStatus) {
	this.AssertIndividualJobStatusEquals(message + ": incorrect Summary", expected.Summary, actual.Summary)
	this.AssertEquals(message + ": incorrect number of individual jobs", len(expected.Jobs), len(actual.Jobs))
	for jobId, expectedStatus := range expected.Jobs {
		this.AssertIndividualJobStatusEquals(message + ": incorrect status for individual job '" + jobId + "'", expectedStatus, actual.Jobs[jobId])
	}
}

func (this *Test) AssertIndividualJobStatusEquals(message string, expected, actual IndividualJobStatus) {
	this.AssertEquals(message + ": incorrect Status", expected.Status, actual.Status)
	this.AssertEquals(message + ": incorrect Message", expected.Message, actual.Message)
	this.AssertEquals(message + ": incorrect NumberOfRecipients", expected.NumberOfRecipients, actual.NumberOfRecipients)
	this.AssertEquals(message + ": incorrect NumberOfSentMails", expected.NumberOfSentMails, actual.NumberOfSentMails)
	this.AssertEquals(message + ": incorrect NumberOfSendingFailures", expected.NumberOfSendingFailures, actual.NumberOfSendingFailures)
	this.AssertEquals(message + ": incorrect NumberOfCurrentlySendingMails", expected.NumberOfCurrentlySendingMails, actual.NumberOfCurrentlySendingMails)
}

func (this *Test) AssertMailSendingFailureSliceEquals(message string, expected, actual []MailSendingFailure) {
	this.AssertEquals("incorrect number of sending failures", len(expected), len(actual))
	for index, expectedFailure := range expected {
		actualFailure := actual[index]
		this.AssertEquals(message + ": incorrect Recipient", expectedFailure.Recipient, actualFailure.Recipient)
		this.AssertEquals(message + ": incorrect Cause", expectedFailure.Cause, actualFailure.Cause)
	}
}






func TestFullJobPreparation(t *testing.T) {
	test := Test{t}
	recipient1 := `{"email": "foo@bar.de", "firstName": "Egon", "lastName": "Björk", "additional": "Test"}`
	recipient2 := `{"email": "frhuw@def.de", "firstName": "Test", "lastName": "Fun"}`
	jobId := "miniJob"

	testRedis := createAndStartTestRedis()
	defer testRedis.Terminate()
	repository := createRepository(testRedis)

	// preparation phase
	test.AssertIsNil("error while creating job", repository.CreateJob(jobId, "HTML"))
	test.AssertIsNil("error while adding recipient1", repository.AddRecipient(jobId, recipient1))
	test.AssertIsNil("error while adding recipient2", repository.AddRecipient(jobId, recipient2))
	test.AssertIsNil("error while finishing preparation", repository.FinishPreparation(jobId, 2))

	// validate content of redis
	connection := repository.Pool.Get()
	defer connection.Close()

	status, err := connection.Do("GET", getJobStatusKey(jobId))
	test.AssertIsNil("Failed to get status", err)
	test.AssertEquals("incorrect job status", "prepared", string(status.([]byte)))

	count, err := connection.Do("GET", getJobNumberOfRecipientsKey(jobId))
	test.AssertIsNil("Failed to get numberOfRecipients", err)
	test.AssertEquals("incorrect job status", "2", string(count.([]byte)))

	openJobs, err := connection.Do("SMEMBERS", getOpenJobsKey())
	test.AssertIsNil("Failed to get list of open jobs", err)
	test.AssertEquals("Incorrect list of open jobs", "[miniJob]", fmt.Sprintf("%s", openJobs))

	jobs, err := connection.Do("SMEMBERS", getJobsKey())
	test.AssertIsNil("Failed to get list of jobs", err)
	test.AssertEquals("Incorrect list of jobs", "[miniJob]", fmt.Sprintf("%s", jobs))

	remaining, err := connection.Do("LRANGE", getRecipientsRemainingKey(jobId), 0, 1000)
	test.AssertIsNil("Failed to get list of remaining", err)
	test.AssertEquals("Incorrect list of remaining", fmt.Sprintf("[%s %s]", recipient1, recipient2), fmt.Sprintf("%s", remaining))
}

func TestFullJobPreparationOfJobWithoutRecipients(t *testing.T) {
	test := Test{t}
	jobId := "emptyJob"

	testRedis := createAndStartTestRedis()
	defer testRedis.Terminate()
	repository := createRepository(testRedis)

	// preparation phase
	test.AssertIsNil("error while creating job", repository.CreateJob(jobId, "HTML"))
	test.AssertIsNil("error while finishing preparation", repository.FinishPreparation(jobId, 0))

	// validate content of redis
	connection := repository.Pool.Get()
	defer connection.Close()

	status, err := connection.Do("GET", getJobStatusKey(jobId))
	test.AssertIsNil("Failed to get status", err)
	test.AssertEquals("incorrect job status", "done", string(status.([]byte)))

	count, err := connection.Do("GET", getJobNumberOfRecipientsKey(jobId))
	test.AssertIsNil("Failed to get numberOfRecipients", err)
	test.AssertEquals("incorrect job status", "0", string(count.([]byte)))

	openJobs, err := connection.Do("SMEMBERS", getOpenJobsKey())
	test.AssertIsNil("Failed to get list of open jobs", err)
	test.AssertEquals("Incorrect list of open jobs", "[]", fmt.Sprintf("%s", openJobs))

	jobs, err := connection.Do("SMEMBERS", getJobsKey())
	test.AssertIsNil("Failed to get list of jobs", err)
	test.AssertEquals("Incorrect list of jobs", "[emptyJob]", fmt.Sprintf("%s", jobs))

	remaining, err := connection.Do("LRANGE", getRecipientsRemainingKey(jobId), 0, 1000)
	test.AssertIsNil("Failed to get list of remaining", err)
	test.AssertEquals("Incorrect list of remaining", "[]", fmt.Sprintf("%s", remaining))
}

func TestJobExecutionByEmailSenders(t *testing.T) {
	test := Test{t}
	recipient := `{"email": "foo@bar.de", "firstName": "Egon", "lastName": "Björk", "additional": "Test"}`
	jobId := "miniJob"

	testRedis := createAndStartTestRedis()
	defer testRedis.Terminate()
	repository := createRepository(testRedis)

	// preparation phase
	test.AssertIsNil("error while creating job", repository.CreateJob(jobId, "HTML"))
	test.AssertIsNil("error while adding recipient1", repository.AddRecipient(jobId, recipient))
	test.AssertIsNil("error while finishing preparation", repository.FinishPreparation(jobId, 1))

	// execution phase
	actualRecipient, err := repository.GetNextOpenRecipient(jobId)
	test.AssertIsNil("error while getting next recipient", err)
	test.AssertEquals("incorrect recipient", recipient, fmt.Sprintf("%s", actualRecipient))

	actualRecipient2, err := repository.GetNextOpenRecipient(jobId)
	test.AssertIsNil("error while getting next recipient 2", err)
	test.AssertStringIsEmpty("incorrect recipient 2", fmt.Sprintf("%s", actualRecipient2))

	err = repository.ExtendLease(jobId, actualRecipient)
	test.AssertIsNil("error while extending lease", err)

	// validate content of redis
	connection := repository.Pool.Get()
	defer connection.Close()

	remaining, err := connection.Do("LRANGE", getRecipientsRemainingKey(jobId), 0, 1000)
	test.AssertIsNil("Failed to get remaining", err)
	test.AssertEquals("Incorrect get remaining", "[]", fmt.Sprintf("%s", remaining))

	sending, err := connection.Do("SMEMBERS", getRecipientsSendingKey(jobId))
	test.AssertIsNil("Failed to get sending", err)
	test.AssertEquals("Incorrect sending", fmt.Sprintf("[%s]", recipient), fmt.Sprintf("%s", sending))

	lease, err := connection.Do("GET", getRecipientLeaseKey(jobId, recipient))
	test.AssertIsNil("Failed to get lease", err)
	test.AssertEquals("Incorrect lease", "1", fmt.Sprintf("%s", lease))

	template, err := connection.Do("GET", getJobTemplateKey(jobId))
	test.AssertIsNil("Failed to get template", err)
	test.AssertEquals("Incorrect template", "HTML", fmt.Sprintf("%s", template))
}

func TestJobIsMarkedAsDoneIfLastEmailHasBeenSentAndSentCounterAndFailureCounterAreCorrectAsWellAsSendingFailureList(t *testing.T) {
	test := Test{t}
	recipientSuccess := `{"email": "foo@bar.de", "firstName": "Egon", "lastName": "Björk", "additional": "Test"}`
	recipientFailure := `{"email": ""}`
	recipientFailureId := `recipient1`
	recipientFailure2 := `{"email": "1233453564"}`
	recipientFailure2Id := `recipient2`
	failureCause := &RepositoryError{"failed by test design"}
	jobId := "miniJob"

	testRedis := createAndStartTestRedis()
	defer testRedis.Terminate()
	repository := createRepository(testRedis)

	// preparation phase
	test.AssertIsNil("error while creating job", repository.CreateJob(jobId, "HTML"))
	test.AssertIsNil("error while adding recipient1", repository.AddRecipient(jobId, recipientSuccess))
	test.AssertIsNil("error while adding recipient2", repository.AddRecipient(jobId, recipientFailure))
	test.AssertIsNil("error while adding recipient3", repository.AddRecipient(jobId, recipientFailure2))
	test.AssertIsNil("error while finishing preparation", repository.FinishPreparation(jobId, 3))
	_, err := repository.GetNextOpenRecipient(jobId)
	test.AssertIsNil("error while getting next recipient", err)
	test.AssertIsNil("error while finishing recipient", repository.MarkRecipientAsDone(jobId, recipientSuccess))
	_, err = repository.GetNextOpenRecipient(jobId)
	test.AssertIsNil("error while getting next recipient", err)
	test.AssertIsNil("error while failing recipient", repository.MarkRecipientAsFailed(jobId, recipientFailure, recipientFailureId, failureCause))
	_, err = repository.GetNextOpenRecipient(jobId)
	test.AssertIsNil("error while getting next recipient", err)
	test.AssertIsNil("error while failing recipient 2", repository.MarkRecipientAsFailed(jobId, recipientFailure2, recipientFailure2Id, failureCause))

	// validate content of redis
	connection := repository.Pool.Get()
	defer connection.Close()

	status, err := connection.Do("GET", getJobStatusKey(jobId))
	test.AssertIsNil("Failed to get status", err)
	test.AssertEquals("Incorrect status", "done", fmt.Sprintf("%s", status))

	doneCount, err := connection.Do("GET", getJobNumberOfSentMailsKey(jobId))
	test.AssertIsNil("Failed to get doneCount", err)
	test.AssertEquals("Incorrect doneCount", "1", fmt.Sprintf("%s", doneCount))

	failedCount, err := connection.Do("GET", getJobNumberOfSendingFailuresKey(jobId))
	test.AssertIsNil("Failed to get failedCount", err)
	test.AssertEquals("Incorrect failedCount", "2", fmt.Sprintf("%s", failedCount))

	failures, err := repository.getSendingFailures(jobId, 0, 2)
	test.AssertIsNil("Failed to get failures", err)
	test.AssertMailSendingFailureSliceEquals("Incorrect failures", []MailSendingFailure{
		MailSendingFailure{
			Recipient: recipientFailureId,
			Cause: failureCause.Error(),
		},
		MailSendingFailure{
			Recipient: recipientFailure2Id,
			Cause: failureCause.Error(),
		},
	}, failures)

	openJob, err := repository.GetRandomOpenJob()
	test.AssertIsNil("error while getting open job", err)
	test.AssertStringIsEmpty("job should not be open any more", openJob)
}

func TestJobDeletionRemovesEverything(t *testing.T) {
	test := Test{t}
	jobToDeleteId := "job-to-delete"
	jobToKeepId := "job-to-keep"
	testRedis := createAndStartTestRedis()
	defer testRedis.Terminate()
	repository := createRepository(testRedis)

	// preparation phase
	test.AssertIsNil("error while creating job", repository.CreateJob(jobToKeepId, "HTML"))
	test.AssertIsNil("error while finishing preparation", repository.FinishPreparation(jobToKeepId, 3))

	test.AssertIsNil("error while creating job", repository.CreateJob(jobToDeleteId, "HTML"))
	test.AssertIsNil("error while adding recipient1", repository.AddRecipient(jobToDeleteId, "recipient1"))
	test.AssertIsNil("error while adding recipient2", repository.AddRecipient(jobToDeleteId, "recipient2"))
	test.AssertIsNil("error while adding recipient3", repository.AddRecipient(jobToDeleteId, "recipient3"))
	test.AssertIsNil("error while finishing preparation", repository.FinishPreparation(jobToDeleteId, 3))
	_, err := repository.GetNextOpenRecipient(jobToDeleteId)
	test.AssertIsNil("error while getting next recipient", err)
	test.AssertIsNil("error while finishing recipient", repository.MarkRecipientAsDone(jobToDeleteId, "recipient1"))
	_, err = repository.GetNextOpenRecipient(jobToDeleteId)
	test.AssertIsNil("error while getting next recipient", err)
	test.AssertIsNil("error while failing recipient", repository.MarkRecipientAsFailed(jobToDeleteId, "recipient2", "r2", &RepositoryError{"failed by test design"}))
	test.AssertIsNil("error while deleting job", repository.AbortAndRemoveJob(jobToDeleteId))

	// validate content of redis
	connection := repository.Pool.Get()
	defer connection.Close()

	keys, err := connection.Do("KEYS", jobToDeleteId + "*")
	test.AssertEquals("incorrect KEYs left in redis", "[]", fmt.Sprintf("%s", keys))

	jobs, err := connection.Do("SMEMBERS", getJobsKey())
	test.AssertIsNil("Failed to get list of jobs", err)
	test.AssertEquals("Incorrect list of jobs", "[" + jobToKeepId + "]", fmt.Sprintf("%s", jobs))

	openJobs, err := connection.Do("SMEMBERS", getOpenJobsKey())
	test.AssertIsNil("Failed to get list of open jobs", err)
	test.AssertEquals("Incorrect list of jobs", "[" + jobToKeepId + "]", fmt.Sprintf("%s", openJobs))
}

func TestJobCannotBeCreatedTwice(t *testing.T) {
	test := Test{t}
	jobId := "someJob"

	testRedis := createAndStartTestRedis()
	defer testRedis.Terminate()
	repository := createRepository(testRedis)

	test.AssertIsNil("error while creating job", repository.CreateJob(jobId, "HTML"))
	test.AssertIsNotNil("job should not be created again", repository.CreateJob(jobId, "HTML"))
}

func TestJobIsNotOpenUnlessPreparationIsFinished(t *testing.T) {
	test := Test{t}
	jobId := "someJob"

	testRedis := createAndStartTestRedis()
	defer testRedis.Terminate()
	repository := createRepository(testRedis)

	test.AssertIsNil("error while creating job", repository.CreateJob(jobId, "HTML"))
	openJob, err := repository.GetRandomOpenJob()
	test.AssertIsNil("error while getting open job", err)
	test.AssertStringIsEmpty("job should not be open yet", openJob)
}

func TestExtendLease(t *testing.T) {
	test := Test{t}
	jobId := "someJob"
	recipient := "some recipient"

	testRedis := createAndStartTestRedis()
	defer testRedis.Terminate()
	repository := createRepository(testRedis)

	test.AssertIsNil("error while creating job", repository.ExtendLease(jobId, recipient))

	// validate content of redis
	connection := repository.Pool.Get()
	defer connection.Close()

	lease, err := connection.Do("GET", getRecipientLeaseKey(jobId, recipient))
	test.AssertIsNil("error while getting lease", err)
	test.AssertEquals("incorrect lease", "1", fmt.Sprintf("%s", lease))
}

func TestMarkJobAsFailed(t *testing.T) {
	test := Test{t}
	jobId := "someJob"
	failureReason := "failure by test design"

	testRedis := createAndStartTestRedis()
	defer testRedis.Terminate()
	repository := createRepository(testRedis)

	test.AssertIsNil("error while creating job", repository.CreateJob(jobId, "HTML"))
	test.AssertIsNil("error while finishing job preparation", repository.FinishPreparation(jobId, 100))
	test.AssertIsNil("error while maring job as failed", repository.MarkJobAsFailed(jobId, failureReason))

	// validate content of redis

	// job is no longer marked as open
	openJob, err := repository.GetRandomOpenJob()
	test.AssertIsNil("error while getting open job", err)
	test.AssertStringIsEmpty("job should not be open yet", openJob)

	// check job state
	connection := repository.Pool.Get()
	defer connection.Close()

	status, err := connection.Do("GET", getJobStatusKey(jobId))
	test.AssertIsNil("error while getting job status", err)
	test.AssertEquals("incorrect status", "failed", fmt.Sprintf("%s", status))

	reason, err := connection.Do("GET", getJobStatusMessageKey(jobId))
	test.AssertIsNil("error while getting job failure reason", err)
	test.AssertEquals("incorrect failure reason", failureReason, fmt.Sprintf("%s", reason))
}

func TestReuseRecipientsFromSendingWhenLeaseHasExpired(t *testing.T) {
	test := Test{t}
	jobId := "someJob"
	recipient1 := "some recipient"
	recipient2 := "some other recipient"

	testRedis := createAndStartTestRedis()
	defer testRedis.Terminate()
	repository := createRepository(testRedis)
	connection := repository.Pool.Get()
	defer connection.Close()

	// setup the job and start sending the one recipient
	test.AssertIsNil("error while creating job", repository.CreateJob(jobId, "HTML"))
	test.AssertIsNil("error while adding recipient", repository.AddRecipient(jobId, recipient1))
	test.AssertIsNil("error while adding recipient", repository.AddRecipient(jobId, recipient2))
	test.AssertIsNil("error while finishing job preparation", repository.FinishPreparation(jobId, 2))

	// get all recipients put of remaining
	actualRecipient, err := repository.GetNextOpenRecipient(jobId)
	test.AssertIsNil("error while getting recipient 1", err)
	test.AssertEquals("retrieved recipient 1 is incorrect", recipient1, fmt.Sprintf("%s", actualRecipient))
	actualRecipient, err = repository.GetNextOpenRecipient(jobId)
	test.AssertIsNil("error while getting recipient 2", err)
	test.AssertEquals("retrieved recipient 2 is incorrect", recipient2, fmt.Sprintf("%s", actualRecipient))
	actualRecipient, err = repository.GetNextOpenRecipient(jobId)
	test.AssertIsNil("error while getting another recipient", err)
	test.AssertStringIsEmpty("there should not be any more recipients", fmt.Sprintf("%s", actualRecipient))

	// manually expire lease of recipient2 by removing key
	_, err = connection.Do("DEL", getRecipientLeaseKey(jobId, recipient2))
	test.AssertIsNil("failed to expire lease", err)

	// re-use recipient2 with expired lease
	actualRecipient, err = repository.GetNextOpenRecipient(jobId)
	test.AssertIsNil("error while getting recipient", err)
	test.AssertEquals("retrieved recipient is incorrect", recipient2, fmt.Sprintf("%s", actualRecipient))

	// make sure that recipient1 is not re-used and that recipient2 is not re-used again
	actualRecipient, err = repository.GetNextOpenRecipient(jobId)
	test.AssertIsNil("error while getting another recipient", err)
	test.AssertStringIsEmpty("there should not be any more recipients", fmt.Sprintf("%s", actualRecipient))
}

func TestGetJobStatusProvidesStatusAndIgnoresDuplicateJobIds(t *testing.T) {
	test := Test{t}
	jobIds := []string{"id1", "id2", "id3"}

	testRedis := createAndStartTestRedis()
	defer testRedis.Terminate()
	repository := createRepository(testRedis)
	connection := repository.Pool.Get()
	defer connection.Close()

	/*
	 * insert mock data into redis directly
	 *
	 * jobId | status | message   | #recipients | #sentMails | #sendFailed | currently sending mails
	 * ---------------------------------------------------------------------------------------------
	 * id1   | id1    | message 1 | 5           | 0          | 1           | mail1@id1.job, ..., mail3@id1.job
	 * id2   | id2    | message 2 [ 6           | 1          | 2           | mail1@id2.job, ..., mail3@id2.job
	 * id3   | id3    | message 3 | 7           | 2          | 3           | mail1@id3.job, ..., mail3@id3.job
	 */
	for index, jobId := range (jobIds) {
		numberOfSentMails := index
		numberOfFailures := index + 1
		numberOfRecipients := index + 5
		test.AssertIsNil("failed to insert test data into redis", connection.Send("SET", getJobStatusKey(jobId), jobId))
		test.AssertIsNil("failed to insert test data into redis", connection.Send("SET", getJobStatusMessageKey(jobId), fmt.Sprintf("message %d", index + 1)))
		test.AssertIsNil("failed to insert test data into redis", connection.Send("SET", getJobNumberOfSentMailsKey(jobId), numberOfSentMails))
		test.AssertIsNil("failed to insert test data into redis", connection.Send("SET", getJobNumberOfSendingFailuresKey(jobId), numberOfFailures))
		test.AssertIsNil("failed to insert test data into redis", connection.Send("SET", getJobNumberOfRecipientsKey(jobId), numberOfRecipients))
		for i := 0; i < 3; i++ {
			recipient := fmt.Sprintf("mail%s@%s.job", i, jobId)
			test.AssertIsNil("failed to insert test data into redis", connection.Send("SADD", getRecipientsSendingKey(jobId), recipient))
		}
	}
	connection.Flush()

	status, err := repository.GetJobStatus(append(jobIds, "id3"))
	test.AssertIsNil("error while retrieving job status", err)
	test.AssertJobStatusEquals("incorrect status", JobStatus{
		Jobs: map[string]IndividualJobStatus{
			"id1": IndividualJobStatus{
				Status: "id1",
				Message: "message 1",
				NumberOfRecipients: 5,
				NumberOfSentMails: 0,
				NumberOfSendingFailures: 1,
				NumberOfCurrentlySendingMails: 3,
			},
			"id2": IndividualJobStatus{
				Status: "id2",
				Message: "message 2",
				NumberOfRecipients: 6,
				NumberOfSentMails: 1,
				NumberOfSendingFailures: 2,
				NumberOfCurrentlySendingMails: 3,
			},
			"id3": IndividualJobStatus{
				Status: "id3",
				Message: "message 3",
				NumberOfRecipients: 7,
				NumberOfSentMails: 2,
				NumberOfSendingFailures: 3,
				NumberOfCurrentlySendingMails: 3,
			},
		},
		Summary: IndividualJobStatus{
			Status: "summary",
			NumberOfRecipients: 5 + 6 + 7,
			NumberOfSentMails: 0 + 1 + 2,
			NumberOfSendingFailures: 1 + 2 + 3,
			NumberOfCurrentlySendingMails: 3 + 3 + 3,
		},
	}, status)
}

func TestWriteSendingFailuresToFile(t *testing.T) {
	test := Test{t}
	jobIds := []string{"id1", "id2", "id3"}

	testRedis := createAndStartTestRedis()
	defer testRedis.Terminate()
	repository := createRepository(testRedis)

	for _, jobId := range jobIds {
		err := repository.CreateJob(jobId, "some html template")
		test.AssertIsNil("error while creating job", err)
		email := `failed@job.` + jobId
		recipient := `{ "email": "` + email + `"}`
		cause := &RepositoryError{"trouble with job " + jobId}

		err = repository.MarkRecipientAsFailed(jobId, recipient, email, cause)
		test.AssertIsNil("error while marking recipient as failed", err)
	}

	targetFile := "TestWriteSendingFailuresToFile.csv"
	err := repository.WriteSendingFailuresToFile(targetFile, jobIds)
	test.AssertIsNil("failed to write failed recipients file", err)
	defer os.Remove(targetFile)

	actual, err := ioutil.ReadFile(targetFile)
	test.AssertIsNil("failed to read file containing failed recipients", err)
	expected := `failed@job.id1	trouble with job id1
failed@job.id2	trouble with job id2
failed@job.id3	trouble with job id3
`
	test.AssertEquals("incorrect content of file containing failed recipients", expected, string(actual))
}

func createRepository(testRedis *testRedis) *RedisRepository {
	return &RedisRepository{
		Pool: testRedis.CreateRedisPool(),
	}
}

type testRedis struct {
	port string
	redisServer *exec.Cmd
}

var testIndex = 0

func createAndStartTestRedis() *testRedis {
	// change port for better test isolation
	port := "505" + strconv.Itoa(testIndex)
	testIndex++
	command := exec.Command("/usr/local/bin/redis-server", "--port", port)
	//command.Stdout = os.Stdout
	//command.Stderr = os.Stderr
	go func() {
		command.Run()
	}()
	testRedis := &testRedis{
		port: port,
		redisServer: command,
	}
	testRedis.WaitForRedisServer()
	return testRedis
}

func (this *testRedis) WaitForRedisServer() {
	pool := this.CreateRedisPool()
	maxTries := 10
	for err := this.Ping(pool); err != nil; err = this.Ping(pool) {
		maxTries--
		if maxTries < 5 {
			log.Printf("Waiting for redis-server... %v", err)
		}
		if maxTries == 0 {
			log.Fatalf("Failed to execute tests since failed to connect to redis-server: %v", err)
		}

		time.Sleep(250 * time.Millisecond)
	}
}

func (this *testRedis) Ping(pool *redis.Pool) error {
	connection := pool.Get()
	defer connection.Close()

	_, err := connection.Do("PING")
	return err
}


func (this *testRedis) Terminate() {
	this.redisServer.Process.Kill()
}

func (this *testRedis) CreateRedisPool() *redis.Pool {
	pool := &redis.Pool{
		MaxActive: 10,
		Wait: true,
		MaxIdle: 3,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			connection, err := redis.Dial("tcp", "localhost:" + this.port)
			if err != nil {
				return nil, err
			}
			return connection, err
		},
		TestOnBorrow: func(connection redis.Conn, t time.Time) error {
			_, err := connection.Do("PING")
			return err
		},
	}

	return pool
}