package core

import (
	"net/http"
)

// ICoreError core error interface
type ICoreError interface {
	Error() string
	GetHTTPCode() int
	GetErrno() int
}

// coreError core error define
type coreError struct {
	HTTPCode int
	Errno    int
	Message  string
}

// New http.StatusInternalServerError
func (e *coreError) New(errno int, message string) *coreError {
	e.HTTPCode = http.StatusInternalServerError
	e.Errno = errno
	e.Message = message
	return e
}

// Error get error message
func (e *coreError) Error() string {
	return e.Message
}

// GetHTTPCode get error HTTPCode
func (e *coreError) GetHTTPCode() int {
	return e.HTTPCode
}

// GetErrno get error Errno
func (e *coreError) GetErrno() int {
	return e.Errno
}

// ServerError http.StatusInternalServerError
type ServerError struct {
	coreError
}

// New ServerError.New
func (e *ServerError) New(message string) *ServerError {
	e.HTTPCode = http.StatusInternalServerError
	e.Errno = 0
	e.Message = message
	return e
}

// BusinessError http.StatusInternalServerError
type BusinessError struct {
	coreError
}

// New http.StatusInternalServerError
func (e *BusinessError) New(errno int, message string) *BusinessError {
	e.HTTPCode = http.StatusBadRequest
	e.Errno = errno
	e.Message = message
	return e
}

// DBError http.StatusInternalServerError
type DBError struct {
	coreError
	DBName string
}

// New DBError.New
func (e *DBError) New(dbName string, message string) *DBError {
	e.HTTPCode = http.StatusInternalServerError
	e.Errno = 0
	e.Message = message
	e.DBName = dbName
	return e
}

// ValidationError simple struct to store the Message & Key of a validation error
type ValidationError struct {
	coreError
}

// New ValidationError.New
func (e *ValidationError) New(message string) *ValidationError {
	e.HTTPCode = http.StatusBadRequest
	e.Errno = 0
	e.Message = message
	return e
}

// NotFoundError route not found.
type NotFoundError struct {
	coreError
}

// New NotFoundError.New
func (e *NotFoundError) New(message string) *NotFoundError {
	e.HTTPCode = http.StatusNotFound
	e.Errno = 0
	e.Message = message
	return e
}
