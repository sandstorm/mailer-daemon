package worker

import "fmt"

type WorkerError struct {
	Message string
	Cause error
}

func (this *WorkerError) Error() string {
	if this.Cause == nil {
		return this.Message
	}    else {
		return fmt.Sprintf("%s: %s", this.Message, this.Cause.Error())
	}
}