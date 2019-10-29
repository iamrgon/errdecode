package errdecode

// Option sets an optional parameter for decoders.
type Option func(*Decoder)

// Message is used to translate the message of a matched, decoded error value.
//
// This is particular useful for integrating transformations or performing
// key-based lookups, e.g., locale-based text or remote lookups.
func Message(t MessageTranslatorFunc) Option {
	return func(d *Decoder) { d.msgTranslator = t }
}

// Encoder is used to provide an error classifier.
//
// It is most useful in scenarios where errors need to be checked in a variety
// of ways, e.g., custom error wrapping.
func Encoder(enc EncoderFunc) Option {
	return func(d *Decoder) { d.encoder = enc }
}

// A mirror effect.
func defaultMessageTranslator(msg string) string {
	return msg
}

// Returns the default encoder, which performs the following checks:
//
//	1. Compare the error value to classified error values
//	2. Pass the error value to classified matchers
//
// If any are true, the classification code and message are returned.
//
// In the case of an unclassified error, the zero values are used.
func newDefaultEncoder(rs []Rule) EncoderFunc {
	idx := newRuleIndex(rs)

	return func(err error) (int, string) {
		if code, ok := idx.errToCode[err]; ok {
			return code, idx.codeToMessage[code]
		}
		for code, matcher := range idx.codeToMatcher {
			if isMatch := matcher(err); isMatch {
				return code, idx.codeToMessage[code]
			}
		}
		return 0, "" // unclassified error
	}
}

// ruleIndex represents various convenience maps derived from a rules slice.
// It provides constant-time lookups for fields of importance.
type ruleIndex struct {
	codeToMatcher map[int]MatcherFunc
	codeToMessage map[int]string
	errToCode     map[error]int
}

// newRuleIndex create indexes from the provided rules.
func newRuleIndex(rs []Rule) *ruleIndex {
	codeToMatcher := make(map[int]MatcherFunc)
	codeToMessage := make(map[int]string)
	errToCode := make(map[error]int)

	for _, rule := range rs {
		code := rule.Code

		codeToMessage[code] = rule.Message
		if rule.Match != nil {
			codeToMatcher[code] = rule.Match
		}

		for _, e := range rule.Errors {
			errToCode[e] = code
		}
	}

	return &ruleIndex{codeToMatcher, codeToMessage, errToCode}
}
