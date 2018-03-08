package core

// ServerError http.StatusInternalServerError
type ServerError struct {
	Message string
}

func (s *ServerError) Error() string {
	return s.Message
}

func (v *ServerError) Code() int {
	return 500
}

// BusinessError http.StatusInternalServerError
type BusinessError struct {
	Message string
}

func (s *BusinessError) Error() string {
	return s.Message
}

func (v *BusinessError) Code() int {
	return 500
}

// ValidationError simple struct to store the Message & Key of a validation error
type ValidationError struct {
	Message string
}

func (v *ValidationError) Error() string {
	return v.Message
}

func (v *ValidationError) Code() int {
	return 400
}

// NotFoundError simple struct to store the Message & Key of a validation error
type NotFoundError struct {
	Message string
}

func (n *NotFoundError) Error() string {
	return n.Message
}

func (v *NotFoundError) Code() int {
	return 404
}
