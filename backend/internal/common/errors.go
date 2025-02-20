package common

type AppError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

func (e AppError) Error() string {
	return e.Message
}

func NewNotFoundError(details string) AppError {
	return AppError{
		Code:    "NOT_FOUND",
		Message: "Resource not found",
		Details: details,
	}
}

func NewUnauthorizedError(details string) AppError {
	return AppError{
		Code:    "UNAUTHORIZED",
		Message: "Unauthorized access",
		Details: details,
	}
}

func NewInvalidInputError(details string) AppError {
	return AppError{
		Code:    "INVALID_INPUT",
		Message: "Invalid input provided",
		Details: details,
	}
}

func NewInternalServerError(details string) AppError {
	return AppError{
		Code:    "INTERNAL_ERROR",
		Message: "Internal server error",
		Details: details,
	}
}
