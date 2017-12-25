package core

// ServerError http.StatusInternalServerError
type ServerError struct {
	Message string
}

func (s *ServerError) Error() string {
	return s.Message
}

// ValidationError simple struct to store the Message & Key of a validation error
type ValidationError struct {
	Message string
	Key     string
}

func (v *ValidationError) Error() string {
	return v.Message
}

// NotFoundError simple struct to store the Message & Key of a validation error
type NotFoundError struct {
	Message string
	Key     string
}

func (n *NotFoundError) Error() string {
	return n.Message
}
