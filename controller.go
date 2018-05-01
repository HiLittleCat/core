package core

import (
	"strconv"
)

// IController 控制器接口定义
type IController interface {
	Register()
}

// Controller 控制器
type Controller struct {
	Validate *Validation
}

// Register this controller to the routers
func (c *Controller) Register() {
}

func (c *Controller) getRValue(ctx *Context, key string) string {
	value := ctx.Request.FormValue(key)
	if value == "" {
		value = ctx.PathValue[key]
	}
	return value
}

// IntMin  param must be a integer, and range is [n,]
func (c *Controller) IntMin(ctx *Context, key string, n int) int {
	value, err := strconv.Atoi(c.getRValue(ctx, key))
	if err != nil {
		panic((&ValidationError{}).New("Query param '" + key + "' must be a number."))
	}
	b := c.Validate.Min(value, n)
	if b == false {
		panic((&ValidationError{}).New("Query param '" + key + "' minimum is " + strconv.Itoa(n)))
	}
	return value
}

// IntMax  param must be a integer, and range is [,m]
func (c *Controller) IntMax(ctx *Context, key string, m int) int {
	value, err := strconv.Atoi(c.getRValue(ctx, key))
	if err != nil {
		panic((&ValidationError{}).New("Query param '" + key + "' must be a number."))
	}
	b := c.Validate.Max(value, m)
	if b == false {
		panic((&ValidationError{}).New("Query param '" + key + "' maximum is " + strconv.Itoa(m)))
	}
	return value
}

// IntRange  param must be a integer, and range is [n, m]
func (c *Controller) IntRange(ctx *Context, key string, n int, m int) int {
	value, err := strconv.Atoi(c.getRValue(ctx, key))
	if err != nil {
		panic((&ValidationError{}).New("Query param '" + key + "' must be a number."))
	}
	b := c.Validate.Range(value, n, m)
	if b == false {
		panic((&ValidationError{}).New("Query param '" + key + "' range is " + strconv.Itoa(n) + " to " + strconv.Itoa(m)))
	}
	return value
}

// StrLength param is a string, length must be n
func (c *Controller) StrLength(ctx *Context, key string, n int) string {
	value := c.getRValue(ctx, key)
	b := c.Validate.Length(value, n)
	if b == false {
		panic((&ValidationError{}).New("Query param '" + key + "' Required length is " + strconv.Itoa(n)))
	}
	return value
}

// StrLenRange param is a string, length range is [n,m]
func (c *Controller) StrLenRange(ctx *Context, key string, n int, m int) string {
	value := c.getRValue(ctx, key)
	if value == "" {
		panic((&ValidationError{}).New("Query param '" + key + "' is required. "))
	}
	return value
}

// StrGet get a string param
func (c *Controller) StrGet(ctx *Context, key string) string {
	value := c.getRValue(ctx, key)
	if len(value) > 100 {
		panic((&ValidationError{}).New("Query param '" + key + "' is too lang. "))
	}
	return value
}
