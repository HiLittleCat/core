package core

import (
	"net/http"
	"reflect"
	"strings"
)

type IController interface {
	Init(ctx *Context)
}

type Controller struct {
	Ctx *Context
}

func (controller *Controller) Init(ctx *Context) {
	controller.Ctx = ctx
}

var openAutoController = false
var mRoutering map[string]reflect.Type = make(map[string]reflect.Type)

func AutoController(controller IController) {
	openAutoController = true
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

func controller(ctx *Context) {
	w := ctx.ResponseWriter
	r := ctx.Request
	controllerName, methodName := findControllerInfo(r)
	controllerT, ok := mRoutering[controllerName]
	if !ok {
		http.NotFound(w, r)
		return
	}
	refV := reflect.New(controllerT)
	method := refV.MethodByName(methodName)
	if !method.IsValid() {
		http.NotFound(w, r)
		return
	}

	controller := refV.Interface().(IController)
	controller.Init(ctx)
	v := method.Call(nil)
	ctx.ResData = v[0].Interface()
}
