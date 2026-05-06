package apierror

type Error struct {
	StatusCode int
	Code       string
	Message    string
}

func (e *Error) Error() string {
	return e.Message
}

func New(statusCode int, code string, message string) *Error {
	return &Error{
		StatusCode: statusCode,
		Code:       code,
		Message:    message,
	}
}
