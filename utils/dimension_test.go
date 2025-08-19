package utils

import (
	"reflect"
	"testing"
)

func TestDimensionBasic(t *testing.T) {
	tests := []struct {
		name     string
		default_ interface{}
		values   []interface{}
		expected interface{}
	}{
		{"string dimension", "default", []interface{}{"value1", "value2"}, "value1"},
		{"int dimension", 42, []interface{}{100, 200}, 100},
		{"float dimension", 3.14, []interface{}{1.23, 4.56}, 1.23},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &dimension{default_: tt.default_, values: tt.values}

			if d.GetDefault() != tt.default_ {
				t.Errorf("GetDefault() = %v, want %v", d.GetDefault(), tt.default_)
			}

			if !reflect.DeepEqual(d.GetValue(), tt.values) {
				t.Errorf("GetValue() = %v, want %v", d.GetValue(), tt.values)
			}

			if d.GetFirstValue() != tt.expected {
				t.Errorf("GetFirstValue() = %v, want %v", d.GetFirstValue(), tt.expected)
			}
		})
	}
}

func TestDimensionEmptyValues(t *testing.T) {
	d := &dimension{default_: "default"}

	if d.GetFirstValue() != "" {
		t.Errorf("GetFirstValue() with empty values = %v, want empty string", d.GetFirstValue())
	}

	if d.GetValue() != nil {
		t.Errorf("GetValue() with nil values = %v, want nil", d.GetValue())
	}
}

func TestDimensionSetters(t *testing.T) {
	d := &dimension{default_: "initial"}

	// Test SetValue
	newValues := []interface{}{"new1", "new2", "new3"}
	d.SetValue(newValues)

	if !reflect.DeepEqual(d.GetValue(), newValues) {
		t.Errorf("After SetValue(), GetValue() = %v, want %v", d.GetValue(), newValues)
	}

	// Test SetOneValue
	d.SetOneValue("single")
	if len(d.GetValue()) != 1 || d.GetValue()[0] != "single" {
		t.Errorf("After SetOneValue(), GetValue() = %v, want [single]", d.GetValue())
	}

	// Test SetDefault
	d.SetDefault("new_default")
	if d.GetDefault() != "new_default" {
		t.Errorf("After SetDefault(), GetDefault() = %v, want new_default", d.GetDefault())
	}
}

func TestDimensionSetValueWithEmptyDefault(t *testing.T) {
	d := &dimension{default_: ""}
	values := []interface{}{"first", "second"}
	d.SetValue(values)

	if d.GetDefault() != "first" {
		t.Errorf("SetValue with empty default should set default to first value, got %v", d.GetDefault())
	}
}

func TestDimensionEq(t *testing.T) {
	tests := []struct {
		name string
		d1   *dimension
		d2   Dimension
		want bool
	}{
		{"equal dimensions", &dimension{values: []interface{}{"a", "b"}}, &dimension{values: []interface{}{"a", "b"}}, true},
		{"different lengths", &dimension{values: []interface{}{"a"}}, &dimension{values: []interface{}{"a", "b"}}, false},
		{"different values", &dimension{values: []interface{}{"a", "b"}}, &dimension{values: []interface{}{"a", "c"}}, false},
		{"empty dimensions", &dimension{}, &dimension{}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.d1.Eq(tt.d2); got != tt.want {
				t.Errorf("Eq() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestValueToString(t *testing.T) {
	tests := []struct {
		name  string
		input interface{}
		want  string
	}{
		{"string value", "hello", "hello"},
		{"int value", 42, "42"},
		{"float64 value", 3.14, "3.14"},
		{"nil value", nil, ""},
		{"unsupported type", true, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ValueToString(tt.input); got != tt.want {
				t.Errorf("ValueToString(%v) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestNewDimensions(t *testing.T) {
	defaults := map[string]interface{}{
		"width":  256,
		"height": 256,
		"format": "png",
	}

	dims := NewDimensions(defaults)

	if len(dims) != 3 {
		t.Errorf("NewDimensions() length = %d, want 3", len(dims))
	}

	for key, want := range defaults {
		if dim, ok := dims[key]; !ok {
			t.Errorf("NewDimensions() missing key %s", key)
		} else if dim.GetDefault() != want {
			t.Errorf("NewDimensions() default for %s = %v, want %v", key, dim.GetDefault(), want)
		}
	}
}

func TestNewDimensionsFromValues(t *testing.T) {
	defaults := map[string][]interface{}{
		"width":  {256, 512, 1024},
		"height": {256, 512},
		"format": {"png", "jpg"},
	}

	dims := NewDimensionsFromValues(defaults)

	if len(dims) != 3 {
		t.Errorf("NewDimensionsFromValues() length = %d, want 3", len(dims))
	}

	for key, values := range defaults {
		if dim, ok := dims[key]; !ok {
			t.Errorf("NewDimensionsFromValues() missing key %s", key)
		} else if !reflect.DeepEqual(dim.GetValue(), values) {
			t.Errorf("NewDimensionsFromValues() values for %s = %v, want %v", key, dim.GetValue(), values)
		} else if dim.GetDefault() != values[0] {
			t.Errorf("NewDimensionsFromValues() default for %s = %v, want %v", key, dim.GetDefault(), values[0])
		}
	}
}

func TestDimensionsGet(t *testing.T) {
	dims := NewDimensions(map[string]interface{}{
		"width":  256,
		"height": 256,
	})

	// Test existing key
	if values := dims.Get("width", nil); !reflect.DeepEqual(values, []interface{}{256}) {
		t.Errorf("Get('width') = %v, want [256]", values)
	}

	// Test non-existing key
	if values := dims.Get("nonexistent", nil); values != nil {
		t.Errorf("Get('nonexistent') = %v, want nil", values)
	}

	// Test with values set
	dims["width"].SetValue([]interface{}{512, 1024})
	if values := dims.Get("width", nil); !reflect.DeepEqual(values, []interface{}{512, 1024}) {
		t.Errorf("Get('width') after SetValue = %v, want [512, 1024]", values)
	}
}

func TestDimensionsSet(t *testing.T) {
	dims := NewDimensions(map[string]interface{}{
		"width": 256,
	})

	// Test setting existing dimension
	newValues := []interface{}{512, 1024}
	dims.Set("width", newValues, nil)

	if values := dims.Get("width", nil); !reflect.DeepEqual(values, newValues) {
		t.Errorf("Set('width') = %v, want %v", values, newValues)
	}

	// Test setting new dimension
	dims.Set("height", []interface{}{256, 512}, nil)
	if values := dims.Get("height", nil); !reflect.DeepEqual(values, []interface{}{256, 512}) {
		t.Errorf("Set('height') = %v, want [256, 512]", values)
	}

	// Test setting with default
	defaultValue := "jpg"
	defaultInterface := interface{}(defaultValue)
	dims.Set("format", []interface{}{"png", "jpg"}, &defaultInterface)
	if dim, ok := dims["format"]; !ok {
		t.Error("Set('format') failed")
	} else if dim.GetDefault() != "jpg" {
		t.Errorf("Set('format') default = %v, want jpg", dim.GetDefault())
	}
}

func TestDimensionsHasValue(t *testing.T) {
	dims := NewDimensions(map[string]interface{}{
		"width": 256,
	})

	if !dims.HasValueOrDefault("width") {
		t.Error("HasValueOrDefault('width') should be true for existing dimension")
	}

	if dims.HasValueOrDefault("nonexistent") {
		t.Error("HasValueOrDefault('nonexistent') should be false for non-existing dimension")
	}

	// Test with values set
	dims["width"].SetValue([]interface{}{512})
	if !dims.HasValue("width") {
		t.Error("HasValue('width') should be true when values are set")
	}

	// Test with empty values
	dims["width"].SetValue([]interface{}{})
	if dims.HasValue("width") {
		t.Error("HasValue('width') should be false when values are empty")
	}
}

func TestDimensionsGetRawMap(t *testing.T) {
	dims := NewDimensionsFromValues(map[string][]interface{}{
		"width":  {256, 512},
		"height": {256},
	})

	rawMap := dims.GetRawMap()

	expected := map[string][]interface{}{
		"width":  {256},
		"height": {256},
	}

	if len(rawMap) != len(expected) {
		t.Errorf("GetRawMap() length = %d, want %d", len(rawMap), len(expected))
	}

	for key, values := range expected {
		if !reflect.DeepEqual(rawMap[key], values) {
			t.Errorf("GetRawMap()[%s] = %v, want %v", key, rawMap[key], values)
		}
	}
}

func TestDimensionsEq(t *testing.T) {
	dims1 := NewDimensionsFromValues(map[string][]interface{}{
		"width":  {256, 512},
		"height": {256, 512},
		"format": {"png", "jpg"},
	})

	dims2 := NewDimensionsFromValues(map[string][]interface{}{
		"width":  {256, 512},
		"height": {256, 512},
		"format": {"png", "jpg"},
	})

	dims3 := NewDimensionsFromValues(map[string][]interface{}{
		"width":  {256, 512},
		"height": {512, 256},
		"format": {"png", "jpg"},
	})

	dims4 := NewDimensionsFromValues(map[string][]interface{}{
		"width": {256, 512},
	})

	if !dims1.Eq(dims2) {
		t.Error("Eq() should return true for identical dimensions")
	}

	if dims1.Eq(dims3) {
		t.Error("Eq() should return false for different values")
	}

	if dims1.Eq(dims4) {
		t.Error("Eq() should return false for different keys")
	}
}

func TestDimensionEdgeCases(t *testing.T) {
	// Test nil values handling
	d := &dimension{default_: "test"}
	if d.GetFirstValue() != "" {
		t.Errorf("GetFirstValue() with nil values = %v, want empty string", d.GetFirstValue())
	}

	// Test empty string default
	d2 := &dimension{default_: ""}
	if d2.GetFirstValue() != "" {
		t.Errorf("GetFirstValue() with empty string default = %v, want empty string", d2.GetFirstValue())
	}

	// Test single value slice
	d3 := &dimension{values: []interface{}{"only"}}
	if d3.GetFirstValue() != "only" {
		t.Errorf("GetFirstValue() with single value = %v, want only", d3.GetFirstValue())
	}

	// Test empty slice
	d4 := &dimension{values: []interface{}{}}
	if d4.GetFirstValue() != "" {
		t.Errorf("GetFirstValue() with empty slice = %v, want empty string", d4.GetFirstValue())
	}
}

// Benchmark tests
func BenchmarkDimensionsGet(b *testing.B) {
	dims := NewDimensionsFromValues(map[string][]interface{}{
		"width":  {256, 512, 1024},
		"height": {256, 512, 1024},
		"format": {"png", "jpg", "webp"},
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = dims.Get("width", nil)
	}
}

func BenchmarkDimensionsSet(b *testing.B) {
	dims := NewDimensions(map[string]interface{}{
		"width":  256,
		"height": 256,
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		dims.Set("width", []interface{}{512, 1024}, nil)
	}
}

func BenchmarkDimensionsEq(b *testing.B) {
	dims1 := NewDimensionsFromValues(map[string][]interface{}{
		"width":  {256, 512},
		"height": {256, 512},
		"format": {"png", "jpg"},
	})

	dims2 := NewDimensionsFromValues(map[string][]interface{}{
		"width":  {256, 512},
		"height": {256, 512},
		"format": {"png", "jpg"},
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = dims1.Eq(dims2)
	}
}
