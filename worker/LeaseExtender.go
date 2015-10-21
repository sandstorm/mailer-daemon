package worker

import (
	"sandstormmedia/project-webessentials-mailer/go/mailer/recipientsRepository"
	"time"
	"log"
)

type LeaseExtender struct {
	Repository recipientsRepository.LeaseExtensionRepository
	JobId string
	Receiver string
	ExtensionInterval time.Duration
	ticker *time.Ticker
}

func (this *LeaseExtender) Start() {
	this.ticker = time.NewTicker(this.ExtensionInterval)
	go func() {
		// ignore first tick to limit load
		// it seems to be triggered at the ticker start-up with little detail
		<- this.ticker.C
		for range this.ticker.C {
			this.extendLease()
		}
	}()
}

// TODO: rename to StopAndRevoke and if stopped revoke lease (?)
func (this *LeaseExtender) Stop() {
	this.ticker.Stop()
}

func (this *LeaseExtender) extendLease() {
	err := this.Repository.ExtendLease(this.JobId, this.Receiver)
	if err != nil {
		log.Println("Failed to extend lease", err)
	}
}
