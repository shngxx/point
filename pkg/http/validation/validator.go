package validation

// Validator defines the interface for request validation
type Validator interface {
	// Validate validates the given value and returns an error if validation fails
	Validate(v any) error
}
