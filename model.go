package core

import (
	"fmt"
	"reflect"
	"strconv"

	"gopkg.in/go-playground/validator.v9"
)

var validate *validator.Validate

// IModel model interface
type IModel interface {
	Err(string, string) error
}

// Model model struct
type Model struct {
}

var validationError *ValidationError

// Err return a controller error
func (m *Model) Err(errno int, message string) error {
	return validationError.New(message)
}

//Check 检查model
func (m *Model) Check() error {
	validate = validate.New()
	err := validate.Struct(m)
	return err
}

//SetDefault 设置默认值，支持类型string和数字
func (m *Model) SetDefault() error {
	cType := reflect.TypeOf(m).Elem()
	cValue := reflect.ValueOf(m).Elem()
	structLen := cValue.NumField()
	for i := 0; i < structLen; i++ {
		field := cType.Field(i)
		defaultValue := field.Tag.Get("default")
		if defaultValue == "" {
			continue
		}
		var v interface{}
		switch cValue.FieldByName(field.Name).Kind() {
		case reflect.String:
			cValue.FieldByName(field.Name).Set(reflect.ValueOf(defaultValue))
		case reflect.Int8:
			v, err := strconv.ParseInt(defaultValue, 10, 8)
			if err != nil {
				return validationError.New(fmt.Printf("model: %s, field: %s, the type of default data is incorrect.", cType, field.Name))
			}
			cValue.FieldByName(field.Name).Set(reflect.ValueOf(int8(v)))
		case reflect.Int16:
			v, err := strconv.ParseInt(defaultValue, 10, 16)
			if err != nil {
				return validationError.New(fmt.Printf("model: %s, field: %s, the type of default data is incorrect.", cType, field.Name))
			}
			cValue.FieldByName(field.Name).Set(reflect.ValueOf(int16(v)))
		case reflect.Int:
			v, err := strconv.ParseInt(defaultValue, 10, 32)
			if err != nil {
				return validationError.New(fmt.Printf("model: %s, field: %s, the type of default data is incorrect.", cType, field.Name))
			}
			cValue.FieldByName(field.Name).Set(reflect.ValueOf(int32(v)))
		case reflect.Int64:
			v, err := strconv.ParseInt(defaultValue, 10, 64)
			if err != nil {
				return validationError.New(fmt.Printf("model: %s, field: %s, the type of default data is incorrect.", cType, field.Name))
			}
			cValue.FieldByName(field.Name).Set(reflect.ValueOf(int64(v)))
		case reflect.Uint8:
			v, err := strconv.ParseUint(defaultValue, 10, 8)
			if err != nil {
				return validationError.New(fmt.Printf("model: %s, field: %s, the type of default data is incorrect.", cType, field.Name))
			}
			cValue.FieldByName(field.Name).Set(reflect.ValueOf(uint8(v)))
		case reflect.Uint16:
			v, err := strconv.ParseUint(defaultValue, 10, 16)
			if err != nil {
				return validationError.New(fmt.Printf("model: %s, field: %s, the type of default data is incorrect.", cType, field.Name))
			}
			cValue.FieldByName(field.Name).Set(reflect.ValueOf(uint16(v)))
		case reflect.Uint32:
			v, err := strconv.ParseUint(defaultValue, 10, 32)
			if err != nil {
				return validationError.New(fmt.Printf("model: %s, field: %s, the type of default data is incorrect.", cType, field.Name))
			}
			cValue.FieldByName(field.Name).Set(reflect.ValueOf(uint32(v)))
		case reflect.Uint64:
			v, err := strconv.ParseUint(defaultValue, 10, 64)
			if err != nil {
				return validationError.New(fmt.Printf("model: %s, field: %s, the type of default data is incorrect.", cType, field.Name))
			}
			cValue.FieldByName(field.Name).Set(reflect.ValueOf(uint64(v)))
		case reflect.Float32:
			v, err := strconv.ParseFloat(defaultValue, 32)
			if err != nil {
				return validationError.New(fmt.Printf("model: %s, field: %s, the type of default data is incorrect.", cType, field.Name))
			}
			cValue.FieldByName(field.Name).Set(reflect.ValueOf(float32(v)))
		case reflect.Float64:
			v, err := strconv.ParseFloat(defaultValue, 64)
			if err != nil {
				return validationError.New(fmt.Printf("model: %s, field: %s, the type of default data is incorrect.", cType, field.Name))
			}
			cValue.FieldByName(field.Name).Set(reflect.ValueOf(float64(v)))
		}
	}
	return nil
}
