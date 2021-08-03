package request

import (
	"net/http"
	"net/url"
	"strings"
)

type ParamPair []string

type RequestParams http.Header

func (p RequestParams) genDict(params []ParamPair) map[string][]string {
	dict := make(map[string][]string)
	for _, pa := range params {
		if len(pa) > 1 {
			upkey := strings.ToLower(pa[0])
			dict[upkey] = pa[1:]
		}
	}
	return dict
}

func (p RequestParams) init(params []ParamPair) {
	dict := p.genDict(params)
	for index, element := range dict {
		p[index] = element
	}
}

func (p RequestParams) Update(params []ParamPair) {
	dict := p.genDict(params)
	for index, element := range dict {
		if _, ok := p[index]; ok {
			p[index] = append(p[index], element...)
		} else {
			p[index] = element
		}
	}
}

func (p RequestParams) Get(key string) (val []string, ok bool) {
	upkey := strings.ToLower(key)
	val, ok = p[upkey]
	return
}

func (p RequestParams) GetOne(key string, defaults string) string {
	upkey := strings.ToLower(key)
	val, ok := p[upkey]
	if ok {
		return val[0]
	}
	return defaults
}

func (p RequestParams) Set(key string, val []string) {
	upkey := strings.ToLower(key)
	p[upkey] = val
}

func (p RequestParams) QueryString() string {
	kv_pairs := []string{}
	for key, values := range p {
		value := strings.Join(values, ",")
		kv_pairs = append(kv_pairs, key+"="+url.QueryEscape(value))
	}
	return strings.Join(kv_pairs, "&")
}

func (p RequestParams) copy() RequestParams {
	map_copy := make(RequestParams)
	for index, element := range p {
		map_copy[index] = element
	}
	return map_copy
}

func (p RequestParams) WithDefaults(defaults RequestParams) RequestParams {
	new := p.copy()
	for key, value := range defaults {
		if value != nil {
			new[key] = value
		}
	}
	return new
}

type Request interface {
	Validate() error
	ToString() string
	QueryString() string
	CompleteUrl() string
	GetParams() RequestParams
	GetRequestHandler() string
}

type BaseRequest struct {
	Request
	Params      RequestParams
	Delimiter   string
	Url         string
	Http        *http.Request
	validate    bool
	FixedParams map[string]string
}

func (r *BaseRequest) init(param interface{}, url string, validate bool, ht *http.Request) error {
	r.Delimiter = ","
	r.Http = ht
	r.validate = validate

	if param == nil {
		r.Params = make(RequestParams)
	} else {
		if ps, ok := param.(RequestParams); ok {
			r.Params = ps
		} else if pr, ok := param.([]ParamPair); ok {
			r.Params = make(RequestParams)
			r.Params.init(pr)
		} else if pr, ok := param.(http.Header); ok {
			r.Params = RequestParams(pr)
		}
	}
	r.Url = url
	if r.validate {
		return r.Validate()
	}
	return nil
}

func (r *BaseRequest) ToString() string {
	return r.CompleteUrl()
}

func (r *BaseRequest) GetParams() RequestParams {
	return r.Params
}

func (r *BaseRequest) GetRawParams() map[string][]string {
	return map[string][]string(r.Params)
}

func (r *BaseRequest) QueryString() string {
	kv_pairs := []string{}
	for key, value := range r.FixedParams {
		r.Params[key] = []string{value}
	}

	for k, v := range r.Params {
		var val string
		if len(v) > 1 {
			val = strings.Join(v, ",")
		} else {
			val = v[0]
		}
		kv_pairs = append(kv_pairs, k+"="+url.QueryEscape(val))
	}

	return strings.Join(kv_pairs, "&")
}

func (r *BaseRequest) CompleteUrl() string {
	if r.Url == "" {
		return r.QueryString()
	}
	delimiter := "?"
	if strings.Contains(r.Url, "?") {
		delimiter = "&"
	}
	if r.Url[len(r.Url)-1] == '?' {
		delimiter = ""
	}
	return r.Url + delimiter + r.QueryString()
}

func (r *BaseRequest) CopyWithRequestParams(req *BaseRequest) *BaseRequest {
	new_params := req.Params.WithDefaults(r.Params)
	return &BaseRequest{Params: new_params, Url: r.Url}
}

func SplitMimeType(mime_type string) [3]string {
	options := ""
	mime_class := ""
	if strings.Contains(mime_type, "/") {
		strs := strings.Split(mime_type, "/")
		mime_class, mime_type = strings.TrimSpace(strs[0]), strs[1]
	}
	if strings.Contains(mime_type, ";") {
		strs := strings.Split(mime_type, ";")
		mime_type, options = strings.TrimSpace(strs[0]), strings.TrimSpace(strs[1])
	}
	return [3]string{mime_class, mime_type, options}
}
