package base

import (
	"encoding/json"
	"fmt"
	"net/http"
)

var errorTable = map[int]*Error{}

// StatefulError describes states of errors.
type StatefulError interface {
	error
	Status() int
	Code() int
	Details() interface{}
}

type Error struct {
	status  int
	code    int
	message string
	details interface{}
}

func (r *Error) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"code": r.code,
		"errors": map[string]interface{}{
			"message": r.message,
			"details": r.details,
		},
	})
}

func (r *Error) Error() string {
	return fmt.Sprintf("%d//%d: %s", r.status, r.code, r.message)
}

func (r *Error) Status() int {
	return r.status
}

func (r *Error) Code() int {
	return r.code
}

func (r *Error) Details() interface{} {
	return r.details
}

func (r *Error) SetDetails(details interface{}) *Error {
	ret := *r

	ret.details = details
	return &ret
}

func DefineError(status, code int, message string) *Error {
	ret := &Error{
		status:  status,
		code:    code,
		message: message,
	}

	errorTable[code] = ret

	return ret
}

var (
	ErrUnknown        = DefineError(http.StatusInternalServerError, 1, "unknown error")
	ErrNotImplemented = DefineError(http.StatusBadRequest, 2, "not implemented")
	ErrParseRequest   = DefineError(http.StatusBadRequest, 3, "failed to parse request")
)
