package core

import (
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"

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
	if len(cookies) == 0 {
		ctx.Next()
		return
	}

	for _, v := range cookies {
		if v.Name == httpCookie.Name {
			cookie = v
			break
		}
	}
	if cookie == nil {
		ctx.Next()
		return
	}

	sid := cookie.Value
	store, err := provider.Get(sid)
	if err != nil {
		log.WithFields(log.Fields{"sid": sid, "err": err}).Warnln("读取session失败")
		ctx.Fail(err)
		return
	}
	if len(store.Values) > 0 {
		//err := provider.refresh(store)
		err := provider.UpExpire(sid)
		if err != nil {
			log.WithFields(log.Fields{"sid": sid, "err": err}).Warnln("刷新session失败")
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
