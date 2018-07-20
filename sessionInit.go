package core

import (
	"net/http"
	"time"

	"github.com/HiLittleCat/conn"
)

// SessionInit 初始化并加载session中间件
func SessionInit(expire time.Duration, pool *conn.RedisPool, cookie http.Cookie) {
	sessExpire = expire
	redisPool = pool
	httpCookie = cookie
	httpCookie.MaxAge = int(sessExpire.Seconds())
	Use(session)
}

// session session处理
func session(ctx *Context) {
	var cookie *http.Cookie
	cookies := ctx.Request.Cookies()
	if len(cookies) > 0 {
		cookie = cookies[0]
	} else {
		ctx.Next()
		return
	}
	sid := cookie.Value
	store, err := provider.Get(sid)
	if err != nil {
		ctx.Fail(err)
		return
	}

	if len(store.Values) > 0 {
		err := provider.refresh(store)
		if err != nil {
			ctx.Fail(err)
			return
		}
		cookie := httpCookie
		cookie.Value = sid
		ctx.Data["session"] = store
		ctx.Data["sid"] = sid
		http.SetCookie(ctx.ResponseWriter, &cookie)
	}

	ctx.Next()
}
