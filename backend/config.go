package backend

import (
	"strings"
	"unicode"
)

const (
	Enable    = "enable"
	EnableOn  = "on"
	EnableOff = "off"
)

const (
	KvSeparator      = `=`
	KvSpaceSeparator = ` `
	KvDoubleQuote    = `"`
)

const (
	EnvPrefix = "TILEPROXY_"
	EnvProxy  = "TILEPROXY_PROXY"
)

type KV struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type KVS []KV

func (kvs KVS) Empty() bool {
	return len(kvs) == 0
}

func (kvs KVS) Keys() []string {
	var keys = make([]string, len(kvs))
	for i := range kvs {
		keys[i] = kvs[i].Key
	}
	return keys
}

func HasSpace(s string) bool {
	for _, r := range s {
		if unicode.IsSpace(r) {
			return true
		}
	}
	return false
}

func (kvs KVS) String() string {
	var s strings.Builder
	for _, kv := range kvs {
		if kv.Key == Enable && kv.Value == EnableOn {
			continue
		}
		s.WriteString(kv.Key)
		s.WriteString(KvSeparator)
		spc := HasSpace(kv.Value)
		if spc {
			s.WriteString(KvDoubleQuote)
		}
		s.WriteString(kv.Value)
		if spc {
			s.WriteString(KvDoubleQuote)
		}
		s.WriteString(KvSpaceSeparator)
	}
	return s.String()
}

func (kvs *KVS) Set(key, value string) {
	for i, kv := range *kvs {
		if kv.Key == key {
			(*kvs)[i] = KV{
				Key:   key,
				Value: value,
			}
			return
		}
	}
	*kvs = append(*kvs, KV{
		Key:   key,
		Value: value,
	})
}

func (kvs KVS) Get(key string) string {
	v, ok := kvs.Lookup(key)
	if ok {
		return v
	}
	return ""
}

func (kvs *KVS) Delete(key string) {
	for i, kv := range *kvs {
		if kv.Key == key {
			*kvs = append((*kvs)[:i], (*kvs)[i+1:]...)
			return
		}
	}
}

func (kvs KVS) Lookup(key string) (string, bool) {
	for _, kv := range kvs {
		if kv.Key == key {
			return kv.Value, true
		}
	}
	return "", false
}

type Config map[string]map[string]KVS
