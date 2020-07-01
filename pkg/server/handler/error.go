package handler

import (
	"fmt"
	"net/http"
	"net/url"

	"golang.org/x/xerrors"
)

type ClientError struct {
	Method string
	URL    *url.URL
	Header http.Header
	Err    error
	frame  xerrors.Frame
}

func NewClientError(r *http.Request, err error) *ClientError {
	return &ClientError{
		Method: r.Method,
		URL:    r.URL,
		Header: r.Header,
		Err:    err,
		frame:  xerrors.Caller(1),
	}
}

func (e *ClientError) Error() string {
	return fmt.Sprintf("client error in %s %s", e.Method, e.URL.String())
}

func (e *ClientError) Unwrap() error {
	return e.Err
}

func (e *ClientError) Format(f fmt.State, c rune) {
	xerrors.FormatError(e, f, c)
}

func (e *ClientError) FormatError(p xerrors.Printer) error {
	p.Print(e.Error())
	e.frame.Format(p)
	return e.Err
}

type ServerError struct {
	Method string
	URL    *url.URL
	Header http.Header
	Err    error
	frame  xerrors.Frame
}

func NewServerError(r *http.Request, err error) *ServerError {
	return &ServerError{
		Method: r.Method,
		URL:    r.URL,
		Header: r.Header,
		Err:    err,
		frame:  xerrors.Caller(1),
	}
}

func (e *ServerError) Error() string {
	return fmt.Sprintf("server error in %s %s", e.Method, e.URL.String())
}

func (e *ServerError) Unwrap() error {
	return e.Err
}

func (e *ServerError) Format(f fmt.State, c rune) {
	xerrors.FormatError(e, f, c)
}

func (e *ServerError) FormatError(p xerrors.Printer) error {
	p.Print(e.Error())
	e.frame.Format(p)
	return e.Err
}

type temporaryError struct {
	s string
}

func (te *temporaryError) Error() string {
	return te.s
}

func (te *temporaryError) Temporary() bool {
	return true
}
