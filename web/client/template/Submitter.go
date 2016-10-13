package template

type (
	// Submitter must be implemented by types controlling forms.
	Submitter interface {
		// Submit will be called when the user presses the submit button.
		// values is a map of form input values.
		Submit(values map[string]interface{})
	}
)
