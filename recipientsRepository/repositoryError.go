package recipientsRepository

type RepositoryError struct {
	message string
}

func (this *RepositoryError) Error() string {
	return this.message
}