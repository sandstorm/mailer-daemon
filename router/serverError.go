package router

type ServerError struct {
	message string
}

func (this *ServerError) Error() string {
	return this.message
}