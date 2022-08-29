package error_handler

import (
	"errors"
	"fmt"
)

type Error struct {
	StatusCode int
	Err error
}

func (r *Error) Error() string {
	return fmt.Sprintf("status %d: err %v", r.StatusCode, r.Err)
}

//Error handler for unauthorized operation
func UnauthorizedHandler() error {
	return &Error{
		StatusCode: 401,
		Err:        errors.New("Unauthorized operation, please use select operation only"),
	}
}

//Error handler for null param handler
func ParamsNullHandler() error {
	return &Error{
		StatusCode: 400,
		Err:        errors.New("Incomplete parameters passed"),
	}
}

//Error handler for checking Db Driver
func IncorrectDbDriver() error {
	return &Error{
		StatusCode: 400,
		Err:        errors.New("Incorrect DB driver provided"),
	}
}
