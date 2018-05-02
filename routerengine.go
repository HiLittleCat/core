package core

// Routers create router instance
var Routers = create()

// Engine each group has a router engine
type Engine struct {
	RouterGroup
	allNoRoute  RouterHandlerChain
	allNoMethod RouterHandlerChain
	noRoute     RouterHandlerChain
	noMethod    RouterHandlerChain
	trees       methodTrees
}

func (engine *Engine) addRoute(method, path string, handlers RouterHandlerChain) {
	assert1(path[0] == '/', "path must begin with '/'")
	assert1(method != "", "HTTP method can not be empty")
	assert1(len(handlers) > 0, "there must be at least one handler")

	root := engine.trees.get(method)
	if root == nil {
		root = new(node)
		engine.trees = append(engine.trees, methodTree{method: method, root: root})
	}
	root.addRoute(path, handlers)
}

// create returns a new blank Engine instance without any middleware attached.
func create() *Engine {
	engine := &Engine{
		RouterGroup: RouterGroup{
			Handlers: nil,
			basePath: "/",
			root:     true,
		},
		trees: make(methodTrees, 0, 9),
	}
	engine.RouterGroup.engine = engine
	return engine
}

func (engine *Engine) handlers(ctx *Context) {
	httpMethod := ctx.Request.Method
	path := ctx.Request.URL.Path
	unescape := false

	// Find root of the tree for the given HTTP method
	t := engine.trees
	for i, tl := 0, len(t); i < tl; i++ {
		if t[i].method == httpMethod {
			root := t[i].root
			// Find route in tree
			handlers, params, _ := root.getValue(path, ctx.Params, unescape)
			if handlers != nil {
				ctx.Params = params
				engine.exeHandlers(ctx, handlers)
				return
			}
			break
		}
	}
}

func (engine *Engine) exeHandlers(ctx *Context, handlers RouterHandlerChain) {
	for _, h := range handlers {
		ctx.handlersStack.Use(h)
	}
	ctx.Next()
}
