package errors

import "fmt"

type (
	ValidationError struct {
		Field   string
		Value   any
		Message string
	}

	ProcessingError struct {
		EventID string
		Cause   error
		Message string
	}

	DatabaseError struct {
		Operation string
		Cause     error
	}
)

func (e ValidationError) Error() string {
	return fmt.Sprintf("validation error for field '%s': %s", e.Field, e.Message)
}

func (e ProcessingError) Error() string {
	if e.EventID != "" {
		return fmt.Sprintf("processing error for event '%s': %s", e.EventID, e.Message)
	}
	return fmt.Sprintf("processing error: %s", e.Message)
}

func (e DatabaseError) Error() string {
	return fmt.Sprintf("database error during %s: %v", e.Operation, e.Cause)
}

func NewValidationError(field string, value any, message string) ValidationError {
	return ValidationError{
		Field:   field,
		Value:   value,
		Message: message,
	}
}

func NewProcessingError(eventID, message string, cause error) ProcessingError {
	return ProcessingError{
		EventID: eventID,
		Message: message,
		Cause:   cause,
	}
}

func NewDatabaseError(operation string, cause error) DatabaseError {
	return DatabaseError{
		Operation: operation,
		Cause:     cause,
	}
}

func IsValidationError(err error, target **ValidationError) bool {
	if val, ok := err.(ValidationError); ok {
		*target = &val
		return true
	}
	return false
}

func IsProcessingError(err error, target **ProcessingError) bool {
	if proc, ok := err.(ProcessingError); ok {
		*target = &proc
		return true
	}
	return false
}

func IsDatabaseError(err error, target **DatabaseError) bool {
	if db, ok := err.(DatabaseError); ok {
		*target = &db
		return true
	}
	return false
}
