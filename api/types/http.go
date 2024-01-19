package types

// HTTPError is used to render response for error.
type HTTPError struct {
	Error string `json:"error"`
}
