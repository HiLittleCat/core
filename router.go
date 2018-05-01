package core

import (
	"net/http"
	"reflect"
	"strings"
)

// PathMinLength set path min length
const PathMinLength = 2

// PathMaxLength set path max length
const PathMaxLength = 8

var methods = [...]string{"GET", "POST", "DELETE", "PATCH", "PUT", "HEAD", "OPTIONS"}

// Routers init the Router
var Routers = &Router{}

// init route map
type nodeType int8

// HandlerFunc http handler
type HandlerFunc func(*Context) (interface{}, error)

const (
	static nodeType = iota // default
	ctl
	nMethod
	root
	param
)

var routerTree = routerNode{
	nType:  root,
	childs: []*routerNode{},
}

type routerNode struct {
	nType     nodeType
	path      string
	maxLength int8
	handler   HandlerFunc
	childs    []*routerNode
}

// Router router struct
type Router struct{}

// AddController add controller to router
func (r *Router) AddController(controller IController) IRouter {
	refV := reflect.ValueOf(controller)
	refT := reflect.Indirect(refV).Type()
	refName := strings.ToLower(refT.Name())
	b := r.isCtlRepeat(routerTree.childs, refName)
	if b == true {
		panic("Controller repeat")
	}
	node := &routerNode{
		nType:  ctl,
		path:   refName,
		childs: []*routerNode{},
	}
	routerTree.childs = append(routerTree.childs, node)
	return &RouterController{Name: refName}
}

// RouteHandler check router, and call handler func
func (r *Router) RouteHandler(ctx *Context) {
	paths, ctlName := r.dealPath(ctx.Request.URL.Path)
	pLength := len(paths)

	method := ctx.Request.Method

	var ctlNode *routerNode
	for _, n := range routerTree.childs {
		if n.path == ctlName {
			ctlNode = n
			break
		}
	}

	if ctlNode == nil {
		ctx.Fail(http.StatusNotFound, (&NotFoundError{}).New("Controller not found"))
		return
	}
	var methodNode *routerNode
	for _, n := range ctlNode.childs {
		if n.path == method {
			methodNode = n
			break
		}
	}

	if methodNode == nil {
		ctx.Fail(http.StatusNotFound, (&NotFoundError{}).New("Method not found"))
		return
	}

	node := methodNode
	if len(node.childs) == 0 {
		ctx.Fail(http.StatusNotFound, (&NotFoundError{}).New("Path not found"))
		return
	}
walk:
	for i, path := range paths {
		// path end, the node is a handler node
		if i == pLength-1 {
			for _, n := range node.childs {
				if n.handler == nil {
					ctx.Fail(http.StatusNotFound, (&NotFoundError{}).New("Path not found"))
					return
				}
				if n.nType == static {
					if n.path == path {
						data, err := n.handler(ctx)
						if err != nil {
							ctx.Fail(http.StatusInternalServerError, err)
						} else {
							ctx.Ok(http.StatusOK, data)
						}
						break walk
					}
				} else if n.nType == param {
					if len(n.childs) == 0 {
						ctx.PathValue[strings.TrimPrefix(n.path, ":")] = path
						data, err := n.handler(ctx)
						if err != nil {
							ctx.Fail(http.StatusInternalServerError, err)
						} else {
							ctx.Ok(http.StatusOK, data)
						}
						break walk
					}
				}
			}
			ctx.Fail(http.StatusNotFound, (&NotFoundError{}).New("Path not found"))
			break walk
		} else {
			// the node must be has childs
			var pNode *routerNode
			for _, n := range node.childs {
				if len(node.childs) == 0 || n.maxLength < int8(pLength) {
					continue
				}
				if n.nType == static {
					if n.path == path {
						if n.maxLength < int8(pLength) {
							ctx.Fail(http.StatusNotFound, (&NotFoundError{}).New("Path not found"))
							return
						}
						node = n
						continue walk
					}
				} else if n.nType == param {
					pNode = n
					ctx.PathValue[strings.TrimPrefix(n.path, ":")] = path
					node = n
				}
			}
			if pNode == nil {
				ctx.Fail(http.StatusNotFound, (&NotFoundError{}).New("Path not found"))
				return
			}
		}
	}
}

func (r *Router) dealPath(path string) ([]string, string) {
	path = strings.TrimPrefix(path, "/")
	path = strings.TrimSuffix(path, "/")
	paths := strings.Split(path, "/")
	ctl := paths[0]
	paths = paths[1:]
	return paths, ctl
}

func (r *Router) isCtlRepeat(nodes []*routerNode, path string) bool {
	b := false
	for _, node := range nodes {
		if node.path == path {
			b = true
			break
		}
	}
	return b
}
