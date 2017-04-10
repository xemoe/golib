// Tideland Go Library - Errors
//
// Copyright (C) 2013-2017 Frank Mueller / Tideland / Oldenburg / Germany
//
// All rights reserved. Use of this source code is governed
// by the new BSD license.

package errors

//--------------------
// IMPORTS
//--------------------

import (
	"fmt"
	"path"
	"runtime"
	"strings"
)

//--------------------
// MESSAGES
//--------------------

// Messages contains the message strings for the the error codes.
type Messages map[int]string

// Format returns the formatted error message for code with the
// given arguments.
func (m Messages) Format(code int, args ...interface{}) string {
	if m == nil || m[code] == "" {
		if len(args) == 0 {
			return fmt.Sprintf("[ERRORS:999] invalid error code '%d'", code)
		}
		format := fmt.Sprintf("%v", args[0])
		return fmt.Sprintf(format, args[1:]...)
	}
	format := m[code]
	return fmt.Sprintf(format, args...)
}

//--------------------
// CONSTANTS
//--------------------

const (
	ErrInvalidErrorType = iota + 1
	ErrNotYetImplemented
	ErrDeprecated
)

var errorMessages = Messages{
	ErrInvalidErrorType:  "invalid error type: %T %q",
	ErrNotYetImplemented: "feature is not yet implemented: %q",
	ErrDeprecated:        "feature is deprecated: %q",
}

//--------------------
// ERROR
//--------------------

// errorBox encapsulates an error.
type errorBox struct {
	err  error
	code int
	msg  string
	info *callInfo
}

// newErrorBox creates an initialized error box.
func newErrorBox(err error, code int, msgs Messages, args ...interface{}) *errorBox {
	return &errorBox{
		err:  err,
		code: code,
		msg:  msgs.Format(code, args...),
		info: retrieveCallInfo(),
	}
}

// Error implements the error interface.
func (eb *errorBox) Error() string {
	if eb.err != nil {
		return fmt.Sprintf("[%s:%03d] %s: %v", eb.info.packagePart, eb.code, eb.msg, eb.err)
	}
	return fmt.Sprintf("[%s:%03d] %s", eb.info.packagePart, eb.code, eb.msg)
}

// errorCollection bundles multiple errors.
type errorCollection struct {
	errs []error
}

// Error implements the error interface.
func (ec *errorCollection) Error() string {
	errMsgs := make([]string, len(ec.errs))
	for i, err := range ec.errs {
		errMsgs[i] = err.Error()
	}
	return strings.Join(errMsgs, "\n")
}

// Annotate creates an error wrapping another one together with a
// a code.
func Annotate(err error, code int, msgs Messages, args ...interface{}) error {
	return newErrorBox(err, code, msgs, args...)
}

// New creates an error with the given code.
func New(code int, msgs Messages, args ...interface{}) error {
	return newErrorBox(nil, code, msgs, args...)
}

// Collect collects multiple errors into one.
func Collect(errs ...error) error {
	return &errorCollection{
		errs: errs,
	}
}

// Valid returns true if it is a valid error generated by
// this package.
func Valid(err error) bool {
	_, ok := err.(*errorBox)
	return ok
}

// IsError checks if an error is one created by this
// package and has the passed code
func IsError(err error, code int) bool {
	if e, ok := err.(*errorBox); ok {
		return e.code == code
	}
	return false
}

// Annotated returns the possibly annotated error. In case of
// a different error an invalid type error is returned.
func Annotated(err error) error {
	if e, ok := err.(*errorBox); ok {
		return e.err
	}
	return New(ErrInvalidErrorType, errorMessages, err, err)
}

// Location returns the package and the file name as well as the line
// number of the error.
func Location(err error) (string, string, int, error) {
	if e, ok := err.(*errorBox); ok {
		return e.info.packageName, e.info.fileName, e.info.line, nil
	}
	return "", "", 0, New(ErrInvalidErrorType, errorMessages, err, err)
}

// Stack returns a slice of errors down to the first
// non-errors error in case of annotated errors.
func Stack(err error) []error {
	if eb, ok := err.(*errorBox); ok {
		return append([]error{eb}, Stack(eb.err)...)
	}
	return []error{err}
}

// All returns a slice of errors in case of collected errors.
func All(err error) []error {
	if ec, ok := err.(*errorCollection); ok {
		all := make([]error, len(ec.errs))
		copy(all, ec.errs)
		return all
	}
	return []error{err}
}

// DoAll iterates the passed function over all stacked
// or collected errors or simply the one that's passed.
func DoAll(err error, f func(error)) {
	switch terr := err.(type) {
	case *errorBox:
		for _, serr := range Stack(err) {
			f(serr)
		}
	case *errorCollection:
		for _, aerr := range All(err) {
			f(aerr)
		}
	default:
		f(terr)
	}
}

// IsInvalidTypeError checks if an error signals an invalid
// type in case of testing for an annotated error.
func IsInvalidTypeError(err error) bool {
	return IsError(err, ErrInvalidErrorType)
}

// NotYetImplementedError returns the common error for a not yet
// implemented feature.
func NotYetImplementedError(feature string) error {
	return New(ErrNotYetImplemented, errorMessages, feature)
}

// IsNotYetImplementedError checks if an error signals a not yet
// implemented feature.
func IsNotYetImplementedError(err error) bool {
	return IsError(err, ErrNotYetImplemented)
}

// DeprecatedError returns the common error for a deprecated
// feature.
func DeprecatedError(feature string) error {
	return New(ErrDeprecated, errorMessages, feature)
}

// IsDeprecatedError checks if an error signals deprecated
// feature.
func IsDeprecatedError(err error) bool {
	return IsError(err, ErrDeprecated)
}

//--------------------
// PRIVATE HELPERS
//--------------------

// callInfo bundles the info about the call environment
// when a logging statement occurred.
type callInfo struct {
	packageName string
	packagePart string
	fileName    string
	funcName    string
	line        int
}

// retrieveCallInfo
func retrieveCallInfo() *callInfo {
	pc, file, line, _ := runtime.Caller(3)
	_, fileName := path.Split(file)
	parts := strings.Split(runtime.FuncForPC(pc).Name(), ".")
	pl := len(parts)
	packageName := ""
	funcName := parts[pl-1]

	if parts[pl-2][0] == '(' {
		funcName = parts[pl-2] + "." + funcName
		packageName = strings.Join(parts[0:pl-2], ".")
	} else {
		packageName = strings.Join(parts[0:pl-1], ".")
	}

	packageParts := strings.Split(packageName, "/")
	packagePart := strings.ToUpper(packageParts[len(packageParts)-1])

	return &callInfo{
		packageName: packageName,
		packagePart: packagePart,
		fileName:    fileName,
		funcName:    funcName,
		line:        line,
	}
}

// EOF
