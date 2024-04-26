package share

import (
	"context"
	"fmt"
	"reflect"
	"sync"
)

// var _ context.Context = &Context{}

// Context is a rpcx customized Context that can contains multiple values.
type Context struct {
	tagsLock *sync.Mutex
	tags     map[interface{}]interface{}
	context.Context
}

func NewContext(ctx context.Context) *Context {
	tagsLock := &sync.Mutex{}
	ctx = context.WithValue(ctx, ContextTagsLock, tagsLock)
	return &Context{
		tagsLock: tagsLock,
		Context:  ctx,
		tags:     map[interface{}]interface{}{isShareContext: true},
	}
}

func (c *Context) Lock() {
	c.tagsLock.Lock()
}

func (c *Context) Unlock() {
	c.tagsLock.Unlock()
}

func (c *Context) Value(key interface{}) interface{} {
	c.tagsLock.Lock()
	defer c.tagsLock.Unlock()
	if c.tags == nil {
		c.tags = make(map[interface{}]interface{})
	}

	if v, ok := c.tags[key]; ok {
		return v
	}
	return c.Context.Value(key)
}

func (c *Context) SetValue(key, val interface{}) {
	c.tagsLock.Lock()
	defer c.tagsLock.Unlock()

	if c.tags == nil {
		c.tags = make(map[interface{}]interface{})
	}
	c.tags[key] = val
}

// DeleteKey delete the kv pair by key.
func (c *Context) DeleteKey(key interface{}) {
	c.tagsLock.Lock()
	defer c.tagsLock.Unlock()

	if c.tags == nil || key == nil {
		return
	}
	delete(c.tags, key)
}

func (c *Context) GetReqMetaDataByKey(key string) string {
	c.tagsLock.Lock()
	defer c.tagsLock.Unlock()
	meta := c.getReqMetaData()
	if meta == nil {
		return ""
	}
	return meta[key]
}
func (c *Context) getReqMetaData() map[string]string {
	var meta map[string]string
	if c.tags == nil {
		c.tags = make(map[interface{}]interface{})
	}

	if v, ok := c.tags[ReqMetaDataKey]; ok {
		meta = v.(map[string]string)
	} else if va, ok2 := c.Context.Value(ReqMetaDataKey).(map[string]string); ok2 {
		meta = va
	}
	return meta
}

func (c *Context) SetReqMetaData(key, val string) {
	c.tagsLock.Lock()
	defer c.tagsLock.Unlock()
	meta := c.getReqMetaData()
	if meta == nil {
		meta = make(map[string]string)
		c.tags[ReqMetaDataKey] = meta
	}
	meta[key] = val
}

func (c *Context) String() string {
	return fmt.Sprintf("%v.WithValue(%v)", c.Context, c.tags)
}

func WithValue(parent context.Context, key, val interface{}) *Context {
	if key == nil {
		panic("nil key")
	}
	if !reflect.TypeOf(key).Comparable() {
		panic("key is not comparable")
	}

	tags := make(map[interface{}]interface{})
	tags[key] = val
	return &Context{Context: parent, tags: tags, tagsLock: &sync.Mutex{}}
}

func WithLocalValue(ctx *Context, key, val interface{}) *Context {
	if key == nil {
		panic("nil key")
	}
	if !reflect.TypeOf(key).Comparable() {
		panic("key is not comparable")
	}

	if ctx.tags == nil {
		ctx.tags = make(map[interface{}]interface{})
	}

	ctx.tags[key] = val
	return ctx
}

// IsShareContext checks whether a context is share.Context.
func IsShareContext(ctx context.Context) bool {
	ok := ctx.Value(isShareContext)
	return ok != nil
}
