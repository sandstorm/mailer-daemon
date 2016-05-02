package router

import (
	"log"
	"fmt"
	"os"
	"bufio"
	"github.com/sandstorm/mailer-daemon/recipientsRepository"
	"encoding/json"
)

type RecipientListImporterJob struct {
	Target recipientsRepository.Repository
	SourceRecipientsList string
	Key string
	Templates recipientsRepository.TemplatesString
	Blacklist Blacklist
	isPrepared bool
}

func (this *RecipientListImporterJob) Prepare() error {
	templatesAsJson, err := json.Marshal(this.Templates)
	if err != nil {
		return err
	}

	err = this.Target.CreateJob(this.Key, string(templatesAsJson[:]))
	if err == nil {
		this.isPrepared = true
	}
	return err
}

func (this *RecipientListImporterJob) ExecuteAndLogErrors() {
	if jobErr := this.Execute(); jobErr != nil {
		log.Println("Failed to import recipients:", jobErr)
	}
}

func (this *RecipientListImporterJob) Execute() error {
	if !this.isPrepared {
		return &ServerError{"you must call 'Prepare' before 'Execute'"}
	}
	numberOfRecipients, err := this.importRecipientsFromFile()
	if (err == nil) {
		return this.Target.FinishPreparation(this.Key, numberOfRecipients)
	} else {
		return this.Target.MarkJobAsFailed(this.Key, "preparation failed: " + err.Error())
	}
	return err
}

func (this *RecipientListImporterJob) importRecipientsFromFile() (numberOfRecipients int, err error) {
	recipientsFile, err := os.Open(this.SourceRecipientsList)
	if err != nil {
		return 0, &ServerError{
			fmt.Sprintf("Failed to open '%s': %s", this.SourceRecipientsList, err),
		}
	}
	defer recipientsFile.Close()

	numberOfRecipients = 0
	recipientsScanner := bufio.NewScanner(recipientsFile)
	for recipientsScanner.Scan() {
		recipient := recipientsScanner.Text()
		if this.Blacklist.Contains(recipient) { continue }

		if err := this.Target.AddRecipient(this.Key, recipient); err != nil {
			return numberOfRecipients, &ServerError{
				fmt.Sprintf("Failed to store recipient due to storage failure: %s", err),
			}
		}
		numberOfRecipients++
	}

	if err := recipientsScanner.Err(); err != nil {
		return numberOfRecipients, &ServerError{
			fmt.Sprintf("Failed to entirely scan '%s': %s", this.SourceRecipientsList, err),
		}
	}
	return numberOfRecipients, nil
}
