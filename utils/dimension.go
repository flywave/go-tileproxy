package utils

import "strconv"

type Dimension interface {
	GetDefault() interface{}
	GetValue() []interface{}
	GetFirstValue() interface{}
	SetValue([]interface{})
	SetOneValue(interface{})
	SetDefault(interface{})
	EQ(o Dimension) bool
}

type dimension struct {
	Dimension
	default_ interface{}
	values   []interface{}
}

func (d *dimension) GetDefault() interface{} {
	return d.default_
}

func (d *dimension) GetValue() []interface{} {
	return d.values
}

func (d *dimension) GetFirstValue() interface{} {
	if d.values != nil && len(d.values) > 0 {
		return d.values[0]
	}
	return ""
}

func (d *dimension) SetValue(vals []interface{}) {
	d.values = vals
	if d.default_ == "" {
		d.default_ = vals[0]
	}
}

func (d *dimension) SetOneValue(v interface{}) {
	d.values = []interface{}{v}
}

func (d *dimension) SetDefault(defa interface{}) {
	d.default_ = defa
}

func (d *dimension) EQ(o Dimension) bool {
	if len(d.values) != len(o.GetValue()) {
		return false
	}
	for i := range d.values {
		if d.values[i] != o.GetValue()[i] {
			return false
		}
	}
	return true
}

func ValueToString(v interface{}) string {
	switch val := v.(type) {
	case string:
		return val
	case float64:
		return strconv.FormatFloat(val, 'f', -1, 64)
	case int:
		return strconv.FormatInt(int64(val), 10)
	}
	return ""
}

type Dimensions map[string]Dimension

func NewDimensions(defaults map[string]interface{}) Dimensions {
	ret := make(Dimensions)

	for k, v := range defaults {
		ret[k] = &dimension{default_: v}
	}

	return ret
}

func NewDimensionsFromValues(defaults map[string][]interface{}) Dimensions {
	ret := make(Dimensions)

	for k, v := range defaults {
		ret[k] = &dimension{default_: v[0], values: v}
	}

	return ret
}

func (d Dimensions) Get(key string, default_ interface{}) []interface{} {
	if v, ok := d[key]; ok {
		if v.GetValue() != nil && len(v.GetValue()) > 0 {
			return v.GetValue()
		}
		return []interface{}{v.GetDefault()}
	}
	return nil
}

func (d Dimensions) Set(key string, val []interface{}, default_ *interface{}) {
	if v, ok := d[key]; ok {
		if default_ != nil {
			v.SetDefault(*default_)
		}
		v.SetValue(val)
		d[key] = v
	} else {
		di := &dimension{}
		if default_ != nil {
			di.default_ = *default_
		}
		di.values = val
		d[key] = di
	}
}

func (d Dimensions) HasValue(key string) bool {
	if v, ok := d[key]; ok {
		if v.GetValue() != nil && len(v.GetValue()) > 0 {
			return true
		}
	}
	return false
}

func (d Dimensions) HasValueOrDefault(key string) bool {
	_, ok := d[key]
	return ok
}

func (d Dimensions) GetRawMap() map[string][]interface{} {
	ret := make(map[string][]interface{})
	for k, d := range d {
		if d.GetValue() != nil && len(d.GetValue()) > 0 {
			ret[k] = d.GetValue()
		}
		ret[k] = []interface{}{d.GetDefault()}
	}
	return ret
}

func (d Dimensions) EQ(o Dimensions) bool {
	for k, vs := range d {
		if v, ok := o[k]; ok {
			if !v.EQ(vs) {
				return false
			}
		} else {
			return false
		}
	}
	return true
}
