// Package errdecode provides a way to represent rules for classifying error
// values returned by your application.
//
// Typically, errors in the application domain are either:
//
// a. returned as-is
//
//	if _, err := hex.DecodeString("48656c6c6f20476f7068657221"); err != nil {
//		return err
//	}
//
// b. annotated, e.g., using the %w verb (go 1.13) or similar means
//
//	var ErrInvalidToken = errors.New("invalid token")
//	if _, err := hex.DecodeString("48656..."); err != nil {
//		return fmt.Errorf("%w: could not decode token", ErrInvalidToken)
//	}
//
// It is preferable to follow approach (b) near the top of the call stack,
// so errors can be explained to an end user.
//
// These error classification rules can be composed as such:
//
//	rules := []errdecode.Rule{
//		{
//			Code: 1001,
//			Message: "The provided token is not valid.",
//			Errors: []error{ErrInvalidToken},
//		},
//		{
//			Code: 1002,
//			Message: "An email and password is required.",
//			Errors: []error{ErrMissingEmail, ErrMissingPassword},
//		},
//	}
//
// Once classified, New returns a decoder that can be used to classify and
// decode encountered error values into pre-defined messages.
//
//	func myAppDomainFunc() error {
//		// do some work, then return error...
//		return ErrInvalidToken
//	}
//
//	decoder := errdecode.New(rules)
//	if err := myAppDomainFunc(); err != nil {
//		log.Fatal(decoder.Translate(err))
//	}
//
// The end user sees the message for the classifed error. In this case,
// "The provided token is not valid." rather than the terse, "invalid token".
//
// In the case that an error type needs to be matched, a MatcherFunc can be
// provided when declaring the rule. This is most relevant when creating rules
// to match custom types, e.g., if following approach (a) without any
// top-of-the-call stack handling.
//
//	rules := []errdecode.Rule{
//		{
//			Code: 1001,
//			Message: "The provided token is not valid.",
//			Match: func(err error) bool {
//				switch err {
// 				case hex.ErrLength:
// 					return true
// 				}
//				var errInvalidByte *hex.InvalidByteError
// 				if errors.As(err, &errInvalidByte) {
// 					return true
// 				}
// 				return false
//			},
//		},
//	}
//	decoder := errdecode.New(rules)
//	if _, err := hex.DecodeString("48656..."); err != nil {
//		log.Fatal(decoder.Translate(err))
//	}
//
// This approach can be further extended by defining an optional encoder that
// can classify all encountered errors.
//
//	const (
//		codeGeneralError = iota + 1000
//		codeAuthError
//	)
//
//	set := map[int]string{
//		codeGeneralError: "An unknown error occurred.",
//		codeAuthError: "The provided token is not valid.",
//	}
//
//	dec := errdecode.New(nil, errdecode.Encoder(func(err error) (int, string) {
//		classify := func(code int) (int, string) { return code, set[code] }
//
//	       	switch err {
// 	       	case hex.ErrLength, ErrInvalidToken:
//			return classify(codeAuthError)
// 	       	}
//	       	var errInvalidByte *hex.InvalidByteError
// 	       	if errors.As(err, &errInvalidByte) {
//			return classify(codeAuthError)
// 	       	}
//
// 	       	return classify(codeGeneralError) // catch-all
//	}))
//
// When there are scenarios where a classified error will need to undergo
// further transformations, as is the case when integrating localized text
// lookups, an optional message translator can integrate the key-based lookup.
//
//	dec := errdecode.New(
// 		[]errdecode.Rule{{
// 			Code:    1001,
// 			Message: "error.auth_error",
// 			Errors:  []error{ErrInvalidToken},
// 		}},
// 		errdecode.Message(func(msg string) string {
// 			switch msg {
// 			case "error.auth_error":
//				// Or lookup locale key...
// 				return "Translated authentication error."
// 			default:
// 				return msg
// 			}
// 		}),
// 	)
//
package errdecode
