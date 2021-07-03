package crawler

import (
	"sync"
)

type Context struct {
	contextMap map[string]interface{}
	lock       *sync.RWMutex
}

func NewContext() *Context {
	return &Context{
		contextMap: make(map[string]interface{}),
		lock:       &sync.RWMutex{},
	}
}

func (c *Context) UnmarshalBinary(_ []byte) error {
	return nil
}

func (c *Context) MarshalBinary() (_ []byte, _ error) {
	return nil, nil
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
		return v.(string)
	}
	return ""
}

func (c *Context) GetAny(key string) interface{} {
	c.lock.RLock()
	defer c.lock.RUnlock()
	if v, ok := c.contextMap[key]; ok {
		return v
	}
	return nil
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
