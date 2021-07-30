package utils

type Dimension interface {
	GetDefault() string
	GetValue() []string
	GetFirstValue() string
	SetValue([]string)
	SetOneValue(string)
	SetDefault(string)
	EQ(o Dimension) bool
}

type dimension struct {
	Dimension
	default_ string
	values   []string
}

func (d *dimension) GetDefault() string {
	return d.default_
}

func (d *dimension) GetValue() []string {
	return d.values
}

func (d *dimension) GetFirstValue() string {
	if d.values != nil && len(d.values) > 0 {
		return d.values[0]
	}
	return ""
}

func (d *dimension) SetValue(vals []string) {
	d.values = vals
	if d.default_ == "" {
		d.default_ = vals[0]
	}
}

func (d *dimension) SetOneValue(v string) {
	d.values = []string{v}
}

func (d *dimension) SetDefault(defa string) {
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

type Dimensions map[string]Dimension

func NewDimensions(defaults map[string]string) Dimensions {
	ret := make(Dimensions)

	for k, v := range defaults {
		ret[k] = &dimension{default_: v}
	}

	return ret
}

func NewDimensionsFromValues(defaults map[string][]string) Dimensions {
	ret := make(Dimensions)

	for k, v := range defaults {
		ret[k] = &dimension{default_: v[0], values: v}
	}

	return ret
}

func (d Dimensions) Get(key string, default_ string) []string {
	if v, ok := d[key]; ok {
		if v.GetValue() != nil && len(v.GetValue()) > 0 {
			return v.GetValue()
		}
		return []string{v.GetDefault()}
	}
	return nil
}

func (d Dimensions) Set(key string, val []string, default_ *string) {
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

func (d Dimensions) GetRawMap() map[string][]string {
	ret := make(map[string][]string)
	for k, d := range d {
		if d.GetValue() != nil && len(d.GetValue()) > 0 {
			ret[k] = d.GetValue()
		}
		ret[k] = []string{d.GetDefault()}
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
