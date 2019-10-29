# errdecode [![godoc](https://godoc.org/github.com/iamrgon/errdecode?status.svg)](https://godoc.org/github.com/iamrgon/errdecode) [![Build Status](https://travis-ci.org/iamrgon/errdecode.svg?branch=master)](https://travis-ci.org/iamrgon/errdecode) [![coverage](https://coveralls.io/repos/github/iamrgon/errdecode/badge.svg)](https://coveralls.io/github/iamrgon/errdecode)

Package errdecode provides a way to represent rules for classifying error
values returned by your application.

## Example Usage

```go
import "github.com/iamrgon/errdecode"

var ErrInvalidToken = errors.New("invalid token")
var ErrMissingEmail = errors.New("missing email")
var ErrMissingPassword = errors.New("missing password")

func myAppDomainFunc() error {
  // Do some work, then return error...
  return ErrInvalidToken
}

func main() {
  rules := []errdecode.Rule{
  	{
  		Code: 1001,
  		Message: "The provided token is not valid.",
  		Errors: []error{ErrInvalidToken},
  	},
  	{
  		Code: 1002,
  		Message: "An email and password is required.",
  		Errors: []error{ErrMissingEmail, ErrMissingPassword},
  	},
  }

	decoder := errdecode.New(rules)
	if err := myAppDomainFunc(); err != nil {
		log.Fatal(decoder.Translate(err))
	}
}
```

## License
Apache


