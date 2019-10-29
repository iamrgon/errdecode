package errdecode_test

import (
	"errors"
	"fmt"
	"testing"

	"iamrgon.com/pkg/errdecode"
)

const (
	codeCatchAll = iota + 1000
	codeClientError
	codeCustomError
	codeWrappedError
)

var errClient1 = errors.New("client error 1")
var errClient2 = errors.New("client error 2")
var errWrappedError = fmt.Errorf("%w", errors.New("wrapped error"))
var errUnclassified = errors.New("unclassified error")

type CustomError string

func (e CustomError) Error() string { return string(e) }

func newCustomError(msg string) error {
	err := CustomError(msg)
	return &err
}

func newDecoder() *errdecode.Decoder {
	rules := []errdecode.Rule{
		{
			Code:    codeClientError,
			Message: "error.client",
			Errors:  []error{errClient1, errClient2},
		},
		{
			Code:    codeCustomError,
			Message: "error.custom",
			Match: errdecode.MatcherFunc(func(err error) bool {
				var errCustom *CustomError
				if errors.As(err, &errCustom) {
					return true
				}
				return false
			}),
		},
		{
			Code:    codeWrappedError,
			Message: "error.wrapped",
			Match: func(err error) bool {
				if errors.Is(err, errWrappedError) {
					return true
				}
				return false
			},
		},
	}

	return errdecode.New(rules)
}

func TestDecoderTranslate(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		wantCode int
		wantMsg  string
	}{
		{"unclassified value is returned", errors.New("unmatched"), 0, "unmatched"},
		{"error value match", errClient1, codeClientError, "error.client"},
		{"second error value in group is matched", errClient2, codeClientError, "error.client"},
		{"wrapped error value match", errWrappedError, codeWrappedError, "error.wrapped"},
		{"custom error type match", newCustomError("custom error"), codeCustomError, "error.custom"},
	}

	dec := newDecoder()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := dec.Translate(tt.err)
			if msg := err.Error(); msg != tt.wantMsg {
				t.Fatalf("unexpected message: got='%s' wantMsg='%s'", msg, tt.wantMsg)
			}

			// Only classified errors can be code checked.
			if ce, ok := err.(errdecode.ClassifiedError); ok {
				if code := ce.Code(); code != tt.wantCode {
					t.Fatalf("unexpected code: got=%d want=%d", code, tt.wantCode)
				}
			}
		})
	}
}

func TestUnclassifiedHandling(t *testing.T) {
	errUnclassified = errors.New("unclassified error")
	dec := errdecode.New(nil, errdecode.Encoder(func(err error) (int, string) { return 0, "" }))

	err := dec.Translate(errUnclassified)
	if err != errUnclassified {
		t.Fatalf("expected unclassified error value to be returned")
	}
}

func TestConfiguredCatchAll(t *testing.T) {
	errUnclassified = errors.New("unclassified error")
	dec := errdecode.New([]errdecode.Rule{{
		Code:    codeCatchAll,
		Message: "error.catchall",
		Match:   func(_ error) bool { return true },
	}})

	err := dec.Translate(errUnclassified)
	ce, ok := err.(errdecode.ClassifiedError)
	if !ok {
		t.Fatalf("expected catch-all to be classified")
	}
	if ce.Code() != codeCatchAll {
		t.Fatalf("expected configured catch-all code got=%d", ce.Code())
	}
	if ce.Error() != "error.catchall" {
		t.Fatalf("expected configured catch-all message got=%s", ce.Error())
	}
}

func TestClassifiedErrorUnwrapping(t *testing.T) {
	dec := newDecoder()
	err := dec.Translate(errClient1) // pre-configured in setup

	ce, ok := err.(errdecode.ClassifiedError)
	if !ok {
		t.Fatalf("expected error to be classified")
	}
	if !errors.Is(err, errClient1) {
		t.Fatalf("decoded error does not wrap encountered error: got=%v want=%v", ce.Unwrap(), errClient1)
	}
}

func TestEncoderOption(t *testing.T) {
	const (
		codeGeneralError = iota + 1000
		codeAppDomainError1
	)

	set := map[int]string{
		codeGeneralError:    "An unknown error occurred.",
		codeAppDomainError1: "Description for application domain error.",
	}

	dec := errdecode.New(nil, errdecode.Encoder(func(err error) (int, string) {
		classify := func(code int) (int, string) { return code, set[code] }

		switch err {
		case errClient1:
			return classify(codeAppDomainError1)
		}

		return classify(codeGeneralError) // catch-all
	}))

	tests := []struct {
		name string
		err  error
		want string
	}{
		{"unclassified is handled with catch-all", errors.New("unclassified"), set[codeGeneralError]},
		{"classified returns configured nessage", errClient1, set[codeAppDomainError1]},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := dec.Translate(tt.err)
			if msg := err.Error(); msg != tt.want {
				t.Fatalf("unexpected message: got='%s' want='%s'", msg, tt.want)
			}
		})
	}
}

func TestMessageTranslatorOption(t *testing.T) {
	errExample := errors.New("example error")
	dec := errdecode.New(
		[]errdecode.Rule{{
			Code:    codeCustomError,
			Message: "error.custom_defined",
			Errors:  []error{errExample},
		}},
		errdecode.Message(func(msg string) string {
			switch msg {
			case "error.custom_defined":
				return "Translated custom error."
			default:
				return msg
			}
		}),
	)

	tests := []struct {
		name string
		err  error
		want string
	}{
		{"unclassified returns encountered message", errors.New("unclassified"), "unclassified"},
		{"classified returns configured translation", errExample, "Translated custom error."},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := dec.Translate(tt.err)
			if msg := err.Error(); msg != tt.want {
				t.Fatalf("unexpected message: got='%s' want='%s'", msg, tt.want)
			}
		})
	}
}
