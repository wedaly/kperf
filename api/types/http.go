package types

// HTTPError is used to render response for error.
type HTTPError struct {
	ErrorMessage string `json:"error"`
}

// Error implements error interface.
func (herr HTTPError) Error() string {
	return herr.ErrorMessage
}
