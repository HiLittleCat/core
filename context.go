package core

import (
	"archive/zip"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"

	jsoniter "github.com/json-iterator/go"
)

// Context contains all the data needed during the serving flow, including the standard http.ResponseWriter and *http.Request.
//
// The Data field can be used to pass all kind of data through the handlers stack.
type Context struct {
	ResponseWriter http.ResponseWriter
	Request        *http.Request
	index          int                    // Keeps the actual handler index.
	handlersStack  HandlersStack          // Keeps the reference to the actual handlers stack.
	written        bool                   // A flag to know if the response has been written.
	Params         Params                 // Path Value
	Data           map[string]interface{} // Custom Data
	BodyJSON       map[string]interface{} // body json data
}

// ResFormat response data
type ResFormat struct {
	Ok      bool        `json:"ok"`
	Data    interface{} `json:"data"`
	Message string      `json:"message"`
	Errno   int         `json:"errno"`
}

// Ok Response json
func (ctx *Context) Ok(data interface{}) {
	if ctx.written == true {
		log.WithFields(log.Fields{"path": ctx.Request.URL.Path}).Warnln("Context.Success: request has been writed")
		return
	}
	ctx.written = true
	var json = jsoniter.ConfigCompatibleWithStandardLibrary
	b, _ := json.Marshal(&ResFormat{Ok: true, Data: data})
	ctx.ResponseWriter.WriteHeader(http.StatusOK)
	_, err := ctx.ResponseWriter.Write(b)
	if err != nil {
		log.WithFields(log.Fields{"path": ctx.Request.URL.Path}).Warnln(err.Error())
	}
}

// Fail Response fail
func (ctx *Context) Fail(err error) {
	if err == nil {
		log.WithFields(log.Fields{"path": ctx.Request.URL.Path}).Warnln("Context.Fail: err is nil")
		ctx.ResponseWriter.WriteHeader(err.(*ServerError).HTTPCode)
		_, err = ctx.ResponseWriter.Write(nil)
		return
	}

	if ctx.written == true {
		log.WithFields(log.Fields{"path": ctx.Request.URL.Path}).Warnln("Context.Fail: request has been writed")
		return
	}

	// //判断错误类型，ValidationError，NotFoundError，BusinessError 返回具体的错误信息，其他类型的错误返回服务器错误
	// _, ok := err.(*BusinessError)
	// _, ok = err.(*NotFoundError)
	// _, ok = err.(*BusinessError)
	// message := http.StatusText(http.StatusInternalServerError)
	// if ok == true {
	// 	message = err.Error()
	// }

	errno := 0
	errCore, ok := err.(ICoreError)
	if ok == true {
		errno = errCore.GetErrno()
	}
	ctx.written = true
	if Production == false {
		log.WithFields(log.Fields{"path": ctx.Request.URL.Path}).Warnln(err.Error())
	} else if _, ok := err.(*ServerError); ok == true {
		log.WithFields(log.Fields{"path": ctx.Request.URL.Path}).Warnln(err.Error())
	}

	var json = jsoniter.ConfigCompatibleWithStandardLibrary
	b, _ := json.Marshal(&ResFormat{Ok: false, Message: err.Error(), Errno: errno})

	coreErr, ok := err.(ICoreError)
	if ok == true {
		ctx.ResponseWriter.WriteHeader(coreErr.GetHTTPCode())
	} else {
		ctx.ResponseWriter.WriteHeader(http.StatusInternalServerError)
	}

	_, err = ctx.ResponseWriter.Write(b)
	if err != nil {
		log.WithFields(log.Fields{"path": ctx.Request.URL.Path}).Warnln(err.Error())
	}
}

//ZipHandler 响应下载文件请求，返回zip文件
func (ctx *Context) ZipHandler(fileName string, file []byte) {
	zipName := fileName + ".zip"
	rw := ctx.ResponseWriter
	// 设置header信息中的ctontent-type，对于zip可选以下两种
	// rw.Header().Set("Content-Type", "application/octet-stream")
	rw.Header().Set("Content-Type", "application/zip")
	// 设置header信息中的Content-Disposition为attachment类型
	rw.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", zipName))
	// 向rw中写入zip文件
	// 创建zip.Writer
	zipW := zip.NewWriter(rw)
	defer zipW.Close()

	// 向zip中添加文件
	f, err := zipW.Create(fileName)
	if err != nil {
		log.WithFields(log.Fields{"path": ctx.Request.URL.Path}).Warnln(err.Error())
	}
	// 向文件中写入文件内容
	_, err = f.Write(file)
	if err != nil {
		log.WithFields(log.Fields{"path": ctx.Request.URL.Path}).Warnln(err.Error())
	}
}

// ResFree Response json
func (ctx *Context) ResFree(data interface{}) {
	if ctx.written == true {
		log.WithFields(log.Fields{"path": ctx.Request.URL.Path}).Warnln("Context.Success: request has been writed")
		return
	}
	ctx.written = true
	var json = jsoniter.ConfigCompatibleWithStandardLibrary
	b, _ := json.Marshal(data)
	ctx.ResponseWriter.WriteHeader(http.StatusOK)
	_, err := ctx.ResponseWriter.Write(b)
	if err != nil {
		log.WithFields(log.Fields{"path": ctx.Request.URL.Path}).Warnln(err.Error())
	}
}

// ResStatus Response status code, use http.StatusText to write the response.
func (ctx *Context) ResStatus(code int) (int, error) {
	if ctx.written == true {
		return 0, errors.New("Context.ResStatus: request has been writed")
	}
	ctx.written = true
	ctx.ResponseWriter.WriteHeader(code)
	return fmt.Fprint(ctx.ResponseWriter, http.StatusText(code))
}

// Written tells if the response has been written.
func (ctx *Context) Written() bool {
	return ctx.written
}

// Next calls the next handler in the stack, but only if the response isn't already written.
func (ctx *Context) Next() {
	// Call the next handler only if there is one and the response hasn't been written.
	if !ctx.Written() && ctx.index < len(ctx.handlersStack.Handlers)-1 {
		ctx.index++
		ctx.handlersStack.Handlers[ctx.index](ctx)
	}
}

// Param returns the value of the URL param.
// It is a shortcut for c.Params.ByName(key)
//     router.GET("/user/:id", func(c *gin.Context) {
//         // a GET request to /user/john
//         id := c.Param("id") // id == "john"
//     })
func (ctx *Context) Param(key string) string {
	return ctx.Params.ByName(key)
}

// GetSession get session
func (ctx *Context) GetSession() IStore {
	store := ctx.Data["session"]
	if store == nil {
		return nil
	}
	st, ok := store.(IStore)
	if ok == false {
		return nil
	}
	return st
}

// GetBodyJSON return a json from body
func (ctx *Context) GetBodyJSON() {
	var reqJSON map[string]interface{}
	body, _ := ioutil.ReadAll(ctx.Request.Body)
	defer ctx.Request.Body.Close()
	cType := ctx.Request.Header.Get("Content-Type")
	a := strings.Split(cType, ";")
	if a[0] == "application/x-www-form-urlencoded" {
		reqJSON = make(map[string]interface{})
		reqStr := string(body)
		reqArr := strings.Split(reqStr, "&")
		for _, v := range reqArr {
			param := strings.Split(v, "=")
			reqJSON[param[0]], _ = url.QueryUnescape(param[1])
		}
	} else {
		json.Unmarshal(body, &reqJSON)
	}
	ctx.BodyJSON = reqJSON
}

// SetSession set session
func (ctx *Context) SetSession(key string, values map[string]string) error {
	sid := ctx.genSid(key)
	values["Sid"] = sid
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	token := ctx.genSid(key + timestamp)
	values["Token"] = token
	store, err := provider.Set(sid, values)
	if err != nil {
		return err
	}
	cookie := httpCookie
	cookie.Value = sid
	ctx.Data["session"] = store

	respCookie := ctx.ResponseWriter.Header().Get("Set-Cookie")
	if strings.HasPrefix(respCookie, cookie.Name) {
		ctx.ResponseWriter.Header().Del("Set-Cookie")
	}
	http.SetCookie(ctx.ResponseWriter, &cookie)
	return nil
}

// FreshSession set session
func (ctx *Context) FreshSession(key string) error {
	err := provider.UpExpire(key)
	if err != nil {
		return err
	}
	return nil
}

// DeleteSession delete session
func (ctx *Context) DeleteSession() error {
	sid := ctx.Data["Sid"].(string)
	ctx.Data["session"] = nil
	provider.Destroy(sid)
	cookie := httpCookie
	cookie.MaxAge = -1
	http.SetCookie(ctx.ResponseWriter, &cookie)
	return nil
}

//GetSid 获取sid
func (ctx *Context) GetSid() string {
	sid := ctx.Data["Sid"]
	if sid == nil {
		return ""
	}
	return sid.(string)

}

func (ctx *Context) genSid(key string) string {
	h := md5.New()
	h.Write([]byte(key))
	cipherStr := h.Sum(nil)
	return hex.EncodeToString(cipherStr)
}

// Recover recovers form panics.
// It logs the stack and uses the PanicHandler (or a classic Internal Server Error) to write the response.
//
// Usage:
//
//	defer c.Recover()
func (ctx *Context) Recover() {
	if err := recover(); err != nil {
		if e, ok := err.(*ValidationError); ok == true {
			ctx.Fail(e)
			return
		}

		stack := make([]byte, 64<<10)
		stack = stack[:runtime.Stack(stack, false)]
		log.WithFields(log.Fields{"path": ctx.Request.URL.Path}).Errorln(string(stack))
		if !ctx.Written() {
			ctx.ResponseWriter.Header().Del("Content-Type")

			if ctx.handlersStack.PanicHandler != nil {
				ctx.Data["panic"] = err
				ctx.handlersStack.PanicHandler(ctx)
			} else {
				ctx.Fail((&ServerError{}).New(http.StatusText(http.StatusInternalServerError)))
			}
		}
	}
}

// ctxPool
var ctxPool = sync.Pool{
	New: func() interface{} {
		return &Context{
			Data:          make(map[string]interface{}),
			index:         -1, // Begin with -1 because Next will increment the index before calling the first handler.
			handlersStack: *defaultHandlersStack,
		}
	},
}

func getContext(w http.ResponseWriter, r *http.Request) *Context {
	ctx := ctxPool.Get().(*Context)
	ctx.Request = r
	ctx.ResponseWriter = contextWriter{w, ctx}
	ctx.Data = make(map[string]interface{})
	ctx.handlersStack = *defaultHandlersStack
	return ctx
}

func putContext(ctx *Context) {
	if ctx.Request.Body != nil {
		ctx.Request.Body.Close()
	}
	ctx.Data = nil
	ctx.Params = nil
	ctx.ResponseWriter = nil
	ctx.Request = nil
	ctx.index = -1
	ctx.written = false
	ctx.BodyJSON = nil
	ctxPool.Put(ctx)
}

// contextWriter represents a binder that catches a downstream response writing and sets the context's written flag on the first write.
type contextWriter struct {
	http.ResponseWriter
	context *Context
}

// Write sets the context's written flag before writing the response.
func (w contextWriter) Write(p []byte) (int, error) {
	w.context.written = true
	return w.ResponseWriter.Write(p)
}

// WriteHeader sets the context's written flag before writing the response header.
func (w contextWriter) WriteHeader(code int) {
	w.context.written = true
	w.ResponseWriter.WriteHeader(code)
}
