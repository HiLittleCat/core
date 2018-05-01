package core

import (
	"strconv"
	"strings"
)

type IRouter interface {
	Handle(string, string, HandlerFunc)
	Get(string, HandlerFunc)
	Post(string, HandlerFunc)
	Delete(string, HandlerFunc)
	Patch(string, HandlerFunc)
	Put(string, HandlerFunc)
	Head(string, HandlerFunc)
	Options(string, HandlerFunc)
}

// RouterController route group
type RouterController struct {
	Name string
}

// Handle
func (g *RouterController) Handle(method string, router string, handler HandlerFunc) {
	path := strings.TrimSuffix(strings.TrimPrefix(router, "/"), "/")
	paths := strings.Split(path, "/")
	if len(paths) > PathMaxLength {
		panic("Max length of a router path is " + strconv.Itoa(PathMaxLength))
	} else if len(paths) < PathMinLength {
		panic("Min length of a router path is " + strconv.Itoa(PathMinLength))
	}
	g.generateRouter(method, paths, handler)
}

// Get define get method route
func (g *RouterController) Get(relativePath string, handler HandlerFunc) {
	g.Handle("GET", relativePath, handler)
}

// Post define post method route
func (g *RouterController) Post(relativePath string, handler HandlerFunc) {
	g.Handle("POST", relativePath, handler)
}

// Delete define delete method route
func (g *RouterController) Delete(relativePath string, handler HandlerFunc) {
	g.Handle("DELETE", relativePath, handler)
}

// Patch define patch method route
func (g *RouterController) Patch(relativePath string, handler HandlerFunc) {
	g.Handle("PATCH", relativePath, handler)
}

// Put define put method route
func (g *RouterController) Put(relativePath string, handler HandlerFunc) {
	g.Handle("PUT", relativePath, handler)
}

// Head define head method route
func (g *RouterController) Head(relativePath string, handler HandlerFunc) {
	g.Handle("HEAD", relativePath, handler)
}

// Options define options method route
func (g *RouterController) Options(relativePath string, handler HandlerFunc) {
	g.Handle("OPTIONS", relativePath, handler)
}

func (g *RouterController) generateRouter(method string, paths []string, handler HandlerFunc) {
	MaxLength := len(paths)
	var rNode *routerNode

	// set controller node to rNode
	for _, ctl := range routerTree.childs {
		if ctl.path == g.Name {
			rNode = ctl
			break
		}
	}

	// set method node to rNode
	for _, m := range rNode.childs {
		if m.path == method {
			rNode = m
			break
		}
	}

	// if rNode is nil, init this kind of method node
	if len(rNode.childs) == 0 {
		mNode := &routerNode{
			nType:     nMethod,
			path:      method,
			maxLength: int8(MaxLength),
			childs:    []*routerNode{},
		}
		rNode.childs = append(rNode.childs, mNode)
		rNode = mNode
	}

	// iterator paths, add a new router to the routerTree
	// remove controller from the request path array
	paths = paths[1:]
	pathLength := MaxLength - 1
walk:
	for i, path := range paths {
		node := &routerNode{
			nType:     static,
			path:      path,
			maxLength: int8(MaxLength),
			childs:    []*routerNode{},
		}

		if strings.HasPrefix(path, ":") {
			node.nType = param
		}

		hasOtherParamChild := false
		for _, _node := range rNode.childs {
			if _node.nType == param && _node.path != node.path {
				hasOtherParamChild = true
				break
			}
		}
		if node.nType == param && hasOtherParamChild {
			panic("Has multi param path in one path")
		}
		// path end, this node is a handler node
		if i == pathLength-1 {
			for _, _node := range rNode.childs {
				// node is exist
				if _node.path == path {
					if _node.handler != nil {
						panic("Has repeat routers")
					}
					_node.handler = handler
					break walk
				}
			}
			// node is not exist
			node.handler = handler
			rNode.childs = append(rNode.childs, node)
			break walk
		} else {
			for _, _node := range rNode.childs {
				// node is exist
				if _node.path == path {
					if _node.maxLength < node.maxLength {
						_node.maxLength = node.maxLength
					}
					rNode = _node
					continue walk
				}
			}
			// node is not exist
			rNode.childs = append(rNode.childs, node)
			rNode = node
		}
	}
}
