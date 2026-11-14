package service

// ValidationError represents invalid prediction input
// or insufficient historical match data.
type ValidationError struct {
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}

// RepositoryError represents an error when retrieving
// historical matches from PostgreSQL.
type RepositoryError struct {
	Err error
}

func (e *RepositoryError) Error() string {
	return e.Err.Error()
}

func (e *RepositoryError) Unwrap() error {
	return e.Err
}
