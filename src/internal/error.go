package internal

type ErrorType int

const (
	InternalError ErrorType = 500
	NotFound ErrorType = 404
	BadRequest ErrorType = 400
)


type ServiceError struct {
	ErrorMessage string
	ErrorType ErrorType
}

func NewServiceError(errorType ErrorType, message string) *ServiceError {
	return &ServiceError{ErrorType: errorType, ErrorMessage: message}
}


func NewBadRequestError(message string) *ServiceError {
	return NewServiceError(BadRequest, message)
}

func NewNotFoundError(message string) *ServiceError {
	return NewServiceError(NotFound, message)
}

func (e *ServiceError) Error() string {
	return e.ErrorMessage
}