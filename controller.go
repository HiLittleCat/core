package core

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"unicode/utf8"
)

// IController 控制器接口定义
type IController interface {
	Register()
	Err(int, string) error
}

// Controller 控制器
type Controller struct {
	Validate *Validation
}

// RegisterRouter this controller to the routers
func (c *Controller) RegisterRouter() {
}

// Err return a controller error
func (c *Controller) Err(errno int, message string) error {
	return (&BusinessError{}).New(errno, message)
}

// GetBodyJSON return a json from body
func (c *Controller) GetBodyJSON(ctx *Context) map[string]interface{} {
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
	return reqJSON
}

func (c *Controller) getRValue(ctx *Context, key string) string {
	value := ctx.Request.FormValue(key)
	if value == "" {
		value = ctx.Param(key)
	}
	return value
}

// IntMin  param must be a integer, and range is [n,]
func (c *Controller) IntMin(fieldName string, p interface{}, n int) int {
	if p == nil {
		p = ""
	}
	value, ok := c.toNumber(p)
	if ok == false {
		panic((&ValidationError{}).New(fieldName + "必须是数字"))
	}
	b := c.Validate.Min(value, n)
	if b == false {
		panic((&ValidationError{}).New(fieldName + "最小值为" + strconv.Itoa(n)))
	}
	return value
}

// IntMax  param must be a integer, and range is [,m]
func (c *Controller) IntMax(fieldName string, p interface{}, m int) int {
	if p == nil {
		p = ""
	}
	value, ok := c.toNumber(p)
	if ok == false {
		panic((&ValidationError{}).New(fieldName + "必须是数字."))
	}
	b := c.Validate.Max(value, m)
	if b == false {
		panic((&ValidationError{}).New(fieldName + "最大值为" + strconv.Itoa(m)))
	}
	return value
}

// IntRange  param must be a integer, and range is [n, m]
func (c *Controller) IntRange(fieldName string, p interface{}, n int, m int) int {
	if p == nil {
		p = 0
	}
	value, ok := c.toNumber(p)

	if ok == false {
		panic((&ValidationError{}).New(fieldName + "必须是数字"))
	}
	b := c.Validate.Range(value, n, m)
	if b == false {
		panic((&ValidationError{}).New(fieldName + "值的范围应该从 " + strconv.Itoa(n) + " 到 " + strconv.Itoa(m)))
	}
	return value
}

// IntRangeZoom  param must be a integer, and range is [n, m], tip is zoom.
func (c *Controller) IntRangeZoom(fieldName string, p interface{}, n int, m int, zoom int) int {
	if p == nil {
		p = 0
	}
	value, ok := c.toNumber(p)

	if ok == false {
		panic((&ValidationError{}).New(fieldName + "必须是数字"))
	}
	b := c.Validate.Range(value, n, m)
	if b == false {
		panic((&ValidationError{}).New(fieldName + "值的范围应该从 " + strconv.Itoa(n/zoom) + " 到 " + strconv.Itoa(m/zoom)))
	}
	return value
}

// StrLength param is a string, length must be n
func (c *Controller) StrLength(fieldName string, p interface{}, n int) string {
	if p == nil {
		p = ""
	}
	v, ok := p.(string)
	if ok == false {
		panic((&ValidationError{}).New(fieldName + "长度应该为" + strconv.Itoa(n)))
	}
	b := c.Validate.Length(v, n)
	if b == false {
		panic((&ValidationError{}).New(fieldName + "长度应该为" + strconv.Itoa(n)))
	}
	return v
}

// StrLenRange param is a string, length range is [n,m]
func (c *Controller) StrLenRange(fieldName string, p interface{}, n int, m int) string {
	if p == nil {
		p = ""
	}
	v, ok := p.(string)
	if ok == false {
		panic((&ValidationError{}).New(fieldName + "格式错误"))
	}
	length := utf8.RuneCountInString(v)
	if length > m || length < n {
		panic((&ValidationError{}).New(fieldName + "长度应该从" + strconv.Itoa(n) + "到" + strconv.Itoa(m)))
	}
	return v
}

// StrLenIn param is a string, length is in array
func (c *Controller) StrLenIn(fieldName string, p interface{}, l ...int) string {
	if p == nil {
		p = ""
	}
	v, ok := p.(string)
	if ok == false {
		panic((&ValidationError{}).New(fieldName + "格式错误"))
	}
	length := utf8.RuneCountInString(v)
	b := false
	for i := 0; i < len(l); i++ {
		if l[i] == length {
			b = true
		}
	}
	if b == false {
		panic((&ValidationError{}).New(fieldName + "值的长度应该在" + strings.Replace(strings.Trim(fmt.Sprint(l), "[]"), " ", ",", -1) + "中"))
	}
	return v
}

// StrIn param is a string, the string is in array
func (c *Controller) StrIn(fieldName string, p interface{}, l ...string) string {
	if p == nil {
		p = ""
	}
	v, ok := p.(string)
	if ok == false {
		panic((&ValidationError{}).New(fieldName + "格式错误"))
	}
	b := false
	for i := 0; i < len(l); i++ {
		if l[i] == v {
			b = true
		}
	}
	if b == false {
		panic((&ValidationError{}).New(fieldName + "值应该在" + strings.Replace(strings.Trim(fmt.Sprint(l), "[]"), " ", ",", -1) + "中"))
	}
	return v
}

// GetEmail check is a email
func (c *Controller) GetEmail(fieldName string, p interface{}) string {
	if p == nil {
		p = ""
	}
	v, ok := p.(string)
	if ok == false {
		panic((&ValidationError{}).New(fieldName + "格式错误"))
	}
	b := c.Validate.Email(v)
	if b == false {
		panic((&ValidationError{}).New(fieldName + "格式错误"))
	}
	return v
}

func (c *Controller) toNumber(obj interface{}) (int, bool) {
	var (
		num int
		ok  bool
		err error
	)
	switch reflect.TypeOf(obj).Kind() {
	case reflect.Float64:
		ok = true
		num = int(obj.(float64))
	case reflect.Float32:
		ok = true
		num = int(obj.(float32))
	case reflect.Int64:
		ok = true
		num = int(obj.(int64))
	case reflect.Int:
		ok = true
		num = obj.(int)
	case reflect.String:
		str := obj.(string)
		num, err = strconv.Atoi(str)
		if err != nil {
			ok = false
		} else {
			ok = true
		}
	}
	return num, ok
}
