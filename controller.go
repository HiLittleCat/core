package core

import (
	"strconv"
)

// IController 控制器接口定义
type IController interface {
	Init(ctx *Context)
}

// Controller 控制器
type Controller struct {
	Ctx      *Context
	Validate *Validation
}

// Init 控制器初始化方法
func (c *Controller) Init(ctx *Context) {
	c.Ctx = ctx
}

// ParamMin  Validation FormValue
func (c *Controller) ParamMin(key string, n int) int {
	value, err := strconv.Atoi(c.Ctx.Request.FormValue(key))
	if err != nil {
		panic(ValidationError{Message: "Query param " + key + " must be a number."})
	}
	b := c.Validate.Min(value, n)
	if b == false {
		panic(ValidationError{Message: "Query param " + key + " Minimum is " + strconv.Itoa(n)})
	}
	return value
}

// ParamLength  param length validate
func (c *Controller) ParamLength(key string, n int) string {
	value := c.Ctx.Request.FormValue(key)
	b := c.Validate.Length(value, n)
	if b == false {
		panic(ValidationError{Message: "Query param " + key + " Required length is " + strconv.Itoa(n)})
	}
	return value
}

// ParamRange  param length validate
func (c *Controller) ParamRange(key string, n int, m int) int {
	value, err := strconv.Atoi(c.Ctx.Request.FormValue(key))
	if err != nil {
		panic(ValidationError{Message: "Query param " + key + " must be a number."})
	}
	b := c.Validate.Range(value, n, m)
	if b == false {
		panic(ValidationError{Message: "Query param " + key + " range is " + strconv.Itoa(n) + " to " + strconv.Itoa(m)})
	}
	return value
}

// ParamGet  return param
func (c *Controller) ParamGet(key string) string {
	value := c.Ctx.Request.FormValue(key)
	return value
}

// ParamRequire param must not be ""
func (c *Controller) ParamRequire(key string) string {
	value := c.Ctx.Request.FormValue(key)
	if value == "" {
		panic(ValidationError{Message: "Query param " + key + " is required. "})
	}
	return value
}
