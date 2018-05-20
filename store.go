package core

import (
	"net/http"
	"time"

	"github.com/HiLittleCat/conn"
	redis "gopkg.in/redis.v5"
)

var (
	// sessExpire session expire
	sessExpire time.Duration
	// redisPool redis pool
	redisPool *conn.RedisPool
	// httpCookie http cookie
	httpCookie http.Cookie

	// provider redis session provider
	provider *redisProvider

	cookieValueKey = "_id"
)

// redisStore session store
type redisStore struct {
	SID    string
	Values map[string]string
}

// Set value
func (rs *redisStore) Set(key, value string) error {
	rs.Values[key] = value
	err := provider.refresh(rs)
	return err
}

// Get value
func (rs *redisStore) Get(key string) string {
	if v, ok := rs.Values[key]; ok {
		return v
	}
	return ""
}

// Delete value in redis session
func (rs *redisStore) Delete(key string) error {
	delete(rs.Values, key)
	err := provider.refresh(rs)
	return err
}

// SessionID get redis session id
func (rs *redisStore) SessionID() string {
	return rs.SID
}

// redisProvider redis session redisProvider
type redisProvider struct {
}

// Set value in redis session
func (rp *redisProvider) Set(key string, values map[string]string) (*redisStore, error) {
	rs := &redisStore{SID: key, Values: values}
	err := provider.refresh(rs)
	return rs, err
}

// refresh refresh store to redis
func (rp *redisProvider) refresh(rs *redisStore) error {
	var err error
	redisPool.Exec(func(c *redis.Client) {
		err = c.HMSet(rs.SID, rs.Values).Err()
		if err != nil {
			return
		}
		err = c.Expire(rs.SID, sessExpire).Err()
	})
	return nil
}

// Get read redis session by sid
func (rp *redisProvider) Get(sid string) (*redisStore, error) {
	var rs = &redisStore{}
	var val map[string]string
	var err error
	redisPool.Exec(func(c *redis.Client) {
		val, err = c.HGetAll(sid).Result()
		rs.Values = val
	})
	return rs, err
}

// Destroy delete redis session by id
func (rp *redisProvider) Destroy(sid string) error {
	var err error
	redisPool.Exec(func(c *redis.Client) {
		err = c.Del(sid).Err()
	})
	return err
}

// UpExpire refresh session expire
func (rp *redisProvider) UpExpire(sid string) error {
	var err error
	redisPool.Exec(func(c *redis.Client) {
		err = c.Expire(sid, sessExpire).Err()
	})
	return err
}
