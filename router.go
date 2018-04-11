package core

import (
	"net/http"
	"reflect"
	"strings"
)

var mRoutering = make(map[string]reflect.Type)

// AddController register controller
func AddController(controller IController) {
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

// Router route middleware
func Router(ctx *Context) {
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
	if len(v) == 0 {
		return
	}
	value := v[0].Interface()
	err := v[1].Interface()
	if err != nil {
		ctx.Fail(http.StatusInternalServerError, &ServerError{Message: err.(error).Error()})
		return
	}
	//ctx.Data["ResData"] = value
	ctx.Ok(http.StatusOK, value)
}
