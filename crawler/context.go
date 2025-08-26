package crawler

import (
	"sync"
)

// contextPool 复用Context对象，减少GC压力
var contextPool = sync.Pool{
	New: func() interface{} {
		return &Context{
			contextMap: make(map[string]interface{}, 8), // 预分配容量
			lock:       &sync.RWMutex{},
		}
	},
}

type Context struct {
	contextMap map[string]interface{}
	lock       *sync.RWMutex
}

func NewContext() *Context {
	c := contextPool.Get().(*Context)
	// 清空map，但保留容量
	for k := range c.contextMap {
		delete(c.contextMap, k)
	}
	return c
}

// Release 释放Context回池中
func (c *Context) Release() {
	if c != nil {
		contextPool.Put(c)
	}
}

func (c *Context) Put(key string, value interface{}) {
	c.lock.Lock()
	c.contextMap[key] = value
	c.lock.Unlock()
}

func (c *Context) Get(key string) string {
	c.lock.RLock()
	defer c.lock.RUnlock()
	if v, ok := c.contextMap[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func (c *Context) GetAny(key string) interface{} {
	c.lock.RLock()
	defer c.lock.RUnlock()
	return c.contextMap[key]
}

// GetWithExists 避免二次查找
func (c *Context) GetWithExists(key string) (interface{}, bool) {
	c.lock.RLock()
	defer c.lock.RUnlock()
	v, exists := c.contextMap[key]
	return v, exists
}

// BatchPut 批量设置，减少锁操作
func (c *Context) BatchPut(data map[string]interface{}) {
	if len(data) == 0 {
		return
	}
	c.lock.Lock()
	for k, v := range data {
		c.contextMap[k] = v
	}
	c.lock.Unlock()
}

func (c *Context) ForEach(fn func(k string, v interface{}) interface{}) []interface{} {
	c.lock.RLock()
	defer c.lock.RUnlock()

	ret := make([]interface{}, 0, len(c.contextMap))
	for k, v := range c.contextMap {
		ret = append(ret, fn(k, v))
	}

	return ret
}
