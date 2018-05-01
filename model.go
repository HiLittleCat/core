package core

// IModel model interface
type IModel interface {
	Err(string, string) error
}

// Model model struct
type Model struct {
	Validate *Validation
}
