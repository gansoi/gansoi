package database

type (
	// Validator is a type that can validate itself.
	Validator interface {
		Validate(Reader) error
	}
)
