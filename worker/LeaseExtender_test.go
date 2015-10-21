package worker

import (
	"testing"
	"time"
)

type stoppingRepository struct {
	MaximalNumberOfTick int
	ExtenderToStop *LeaseExtender
	T *testing.T
	numberOfTicks int
}

func (this *stoppingRepository) ExtendLease(jobId, recipient string) error {
	this.numberOfTicks++

	if this.MaximalNumberOfTick == this.numberOfTicks {
		this.ExtenderToStop.Stop()
	}
	if this.numberOfTicks > this.MaximalNumberOfTick {
		this.T.Errorf("too many tick %i: the lease extender should have already been stopped", this.numberOfTicks)
	}
	return nil
}

func TestStartAndStop(t *testing.T) {
	repository := stoppingRepository{
		MaximalNumberOfTick: 3,
		numberOfTicks: 0,
		T: t,
	}
	extender := LeaseExtender{
		JobId: "someJob",
		Receiver: "some@receiver",
		ExtensionInterval: 100 * time.Millisecond,
	}
	repository.ExtenderToStop = &extender
	extender.Repository = &repository

	extender.Start()
	time.Sleep(time.Second)
}