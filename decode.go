package errdecode

// ClassifiedError describes the wrapped error value matched by the
// classification rule set.
type ClassifiedError interface {
	error

	// Code returns the classification for the matched error.
	Code() int

	// Unwrap returns the underlying error.
	Unwrap() error
}

// Rule represents criteria for matching error values.
type Rule struct {
	// Code is an identifier for a class of errors.
	Code int

	// Message describes the error class, e.g., a friendly explanation or
	// a string identifier for key-based lookups.
	Message string

	// Errors are values that fall under this classification.
	Errors []error

	// Match is a func that returns true if a given error is a match.
	// It can be used to check error types by using a closure.
	Match MatcherFunc
}

// MatcherFunc describes an error matcher.
type MatcherFunc func(err error) (isMatch bool)

// Decoder wraps a set of error translation rules, on which it provides
// classication and translation of error values.
type Decoder struct {
	encoder       EncoderFunc
	msgTranslator MessageTranslatorFunc
}

// EncoderFunc describes an error classifier, i.e., a function that converts
// an error value into a recognized code and message.
type EncoderFunc func(err error) (code int, message string)

// MessageTranslatorFunc describes further transformations for decoded errors.
type MessageTranslatorFunc func(decoded string) (translated string)

// New returns a configured error decoder.
func New(rs []Rule, options ...Option) *Decoder {
	d := &Decoder{
		encoder:       newDefaultEncoder(rs),
		msgTranslator: defaultMessageTranslator,
	}
	for _, option := range options {
		option(d)
	}
	return d
}

// Translate decodes an error value into a configured encoded mapping.
// If the error cannot be classified, it is returned as-is.
func (d *Decoder) Translate(err error) error {
	code, msg := d.encoder(err)
	if code == 0 {
		return err
	}
	return &matchedError{code, err, d.msgTranslator(msg)}
}

// Compile-time check.
var _ ClassifiedError = (*matchedError)(nil)

// Represents an error matched by the encoder.
type matchedError struct {
	code int
	err  error
	msg  string
}

// Code satisfies ClassifiedError interface.
func (e *matchedError) Code() int { return e.code }

// Unwrap satisfies ClassifiedError interface.
func (e *matchedError) Unwrap() error { return e.err }

// Error satisties the error interface.
func (e *matchedError) Error() string { return e.msg }
