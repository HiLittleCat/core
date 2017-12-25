package core

import (
	"net/http"
	"reflect"
	"strconv"
	"strings"
)

type IController interface {
	Init(ctx *Context)
}

type Controller struct {
	Ctx      *Context
	Validate *Validation
}

func (c *Controller) Init(ctx *Context) {
	c.Ctx = ctx
}

// ParamMin  Validation FormValue
func (c *Controller) ParamMin(key string, n int) int {
	value, err := strconv.Atoi(c.Ctx.Request.FormValue(key))
	if err != nil {
		panic(ValidationError{Message: "Query param " + key + " Minimum is " + strconv.Itoa(n)})
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

// ParamGet  param get
func (c *Controller) ParamGet(key string) string {
	value := c.Ctx.Request.FormValue(key)
	return value
}

var mRoutering map[string]reflect.Type = make(map[string]reflect.Type)

// AutoController register controller
func AutoController(controller IController) {
	refV := reflect.ValueOf(controller)
	refT := reflect.Indirect(refV).Type()
	refName := strings.ToLower(refT.Name())
	mRoutering[strings.ToLower(refName)] = refT
}

const (
	defController = "index"
	defMethod     = "Index"
)

func findControllerInfo(r *http.Request) (string, string) {
	path := r.URL.Path
	path = strings.TrimSuffix(path, "/")
	pathInfo := strings.Split(path, "/")

	controllerName := defController
	if len(pathInfo) > 0 {
		controllerName = strings.ToLower(pathInfo[1])
	}
	methodName := defMethod
	if len(pathInfo) == 2 {
		methodName = strings.Title(strings.ToLower(r.Method))
	} else if len(pathInfo) == 3 {
		methodName = strings.Title(strings.ToLower(r.Method) + pathInfo[2])
	}
	return controllerName, methodName
}

// AutoRouter route middleware
func AutoRouter(ctx *Context) {
	r := ctx.Request
	controllerName, methodName := findControllerInfo(r)
	controllerT, ok := mRoutering[controllerName]
	if !ok {
		ctx.Fail(http.StatusNotFound, &NotFoundError{Message: "Module Not Found"})
		return
	}
	refV := reflect.New(controllerT)
	method := refV.MethodByName(methodName)
	if !method.IsValid() {
		ctx.Fail(http.StatusNotFound, &NotFoundError{Message: "Method Not Found"})
		return
	}
	ctx.Data["Controller"] = controllerName
	ctx.Data["Method"] = methodName
	controller := refV.Interface().(IController)
	controller.Init(ctx)
	v := method.Call(nil)
	value := v[0].Interface()
	err := v[1].Interface()
	if err != nil {
		ctx.Fail(http.StatusInternalServerError, &ServerError{Message: err.(error).Error()})
		return
	}
	//ctx.Data["ResData"] = value
	ctx.Ok(http.StatusOK, value)
}
