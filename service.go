package core

// IService service interface
type IService interface {
	Err(int, string) error
}

// Service service struct
type Service struct {
	Validate *Validation
}
