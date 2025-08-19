package utils

import (
	"math/rand"
	"reflect"
	"sort"
	"testing"
)

func TestShuffleStrings(t *testing.T) {
	tests := []struct {
		name string
		data []string
	}{
		{"empty slice", []string{}},
		{"single element", []string{"hello"}},
		{"two elements", []string{"hello", "world"}},
		{"multiple elements", []string{"a", "b", "c", "d", "e"}},
		{"duplicate elements", []string{"a", "a", "b", "b", "c"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 设置随机种子以获得可重现的结果
			rng := rand.New(rand.NewSource(42))

			original := make([]string, len(tt.data))
			copy(original, tt.data)

			// 创建副本用于测试
			testData := make([]string, len(tt.data))
			copy(testData, tt.data)

			ShuffleStringsWithRand(testData, rng)

			// 检查长度是否保持不变
			if len(testData) != len(original) {
				t.Errorf("ShuffleStrings() length changed: got %d, want %d", len(testData), len(original))
			}

			// 检查元素是否保持不变（只是顺序变化）
			originalSorted := make([]string, len(original))
			copy(originalSorted, original)
			sort.Strings(originalSorted)

			testDataSorted := make([]string, len(testData))
			copy(testDataSorted, testData)
			sort.Strings(testDataSorted)

			if !reflect.DeepEqual(originalSorted, testDataSorted) {
				t.Errorf("ShuffleStrings() elements changed: got %v, want %v", testData, original)
			}
		})
	}
}

func TestShuffleStringsN(t *testing.T) {
	tests := []struct {
		name        string
		data        []string
		n           int
		wantPanic   bool
		panicString string
	}{
		{"normal case", []string{"a", "b", "c", "d", "e"}, 3, false, ""},
		{"shuffle all", []string{"a", "b", "c"}, 3, false, ""},
		{"n equals length", []string{"a", "b"}, 2, false, ""},
		{"n larger than length", []string{"a", "b"}, 3, true, "ShuffleStringsN n larger than len(slice)"},
		{"n zero", []string{"a", "b"}, 0, true, "ShuffleStringsN n larger than len(slice)"},
		{"n negative", []string{"a", "b"}, -1, true, "ShuffleStringsN n larger than len(slice)"},
		{"empty slice", []string{}, 1, true, "ShuffleStringsN n larger than len(slice)"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantPanic {
				defer func() {
					if r := recover(); r == nil {
						t.Errorf("ShuffleStringsN() did not panic")
					} else if r != tt.panicString {
						t.Errorf("ShuffleStringsN() panic message: got %v, want %v", r, tt.panicString)
					}
				}()
			}

			// 设置随机种子
			rng := rand.New(rand.NewSource(42))

			// 创建副本
			testData := make([]string, len(tt.data))
			copy(testData, tt.data)

			ShuffleStringsNWithRand(testData, tt.n, rng)

			if !tt.wantPanic {
				// 检查长度是否保持不变
				if len(testData) != len(tt.data) {
					t.Errorf("ShuffleStringsN() length changed: got %d, want %d", len(testData), len(tt.data))
				}

				// 检查前n个元素是否被移动（由于随机性，我们只检查是否发生了变化）
				// 这里我们主要检查函数不会panic
			}
		})
	}
}

func TestShuffleInts(t *testing.T) {
	tests := []struct {
		name string
		data []int
	}{
		{"empty slice", []int{}},
		{"single element", []int{42}},
		{"two elements", []int{1, 2}},
		{"multiple elements", []int{1, 2, 3, 4, 5}},
		{"negative numbers", []int{-1, -2, 3, -4}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rng := rand.New(rand.NewSource(42))

			original := make([]int, len(tt.data))
			copy(original, tt.data)

			testData := make([]int, len(tt.data))
			copy(testData, tt.data)

			ShuffleIntsWithRand(testData, rng)

			// 检查长度
			if len(testData) != len(original) {
				t.Errorf("ShuffleInts() length changed: got %d, want %d", len(testData), len(original))
			}

			// 检查元素
			originalSorted := make([]int, len(original))
			copy(originalSorted, original)
			sort.Ints(originalSorted)

			testDataSorted := make([]int, len(testData))
			copy(testDataSorted, testData)
			sort.Ints(testDataSorted)

			if !reflect.DeepEqual(originalSorted, testDataSorted) {
				t.Errorf("ShuffleInts() elements changed: got %v, want %v", testData, original)
			}
		})
	}
}

func TestShuffleUints(t *testing.T) {
	tests := []struct {
		name string
		data []uint
	}{
		{"empty slice", []uint{}},
		{"single element", []uint{42}},
		{"two elements", []uint{1, 2}},
		{"multiple elements", []uint{1, 2, 3, 4, 5}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rng := rand.New(rand.NewSource(42))

			original := make([]uint, len(tt.data))
			copy(original, tt.data)

			testData := make([]uint, len(tt.data))
			copy(testData, tt.data)

			ShuffleUintsWithRand(testData, rng)

			// 检查长度
			if len(testData) != len(original) {
				t.Errorf("ShuffleUints() length changed: got %d, want %d", len(testData), len(original))
			}

			// 检查元素
			originalSorted := make([]uint, len(original))
			copy(originalSorted, original)
			sort.Slice(originalSorted, func(i, j int) bool { return originalSorted[i] < originalSorted[j] })

			testDataSorted := make([]uint, len(testData))
			copy(testDataSorted, testData)
			sort.Slice(testDataSorted, func(i, j int) bool { return testDataSorted[i] < testDataSorted[j] })

			if !reflect.DeepEqual(originalSorted, testDataSorted) {
				t.Errorf("ShuffleUints() elements changed: got %v, want %v", testData, original)
			}
		})
	}
}

func TestShuffleUint16s(t *testing.T) {
	tests := []struct {
		name string
		data []uint16
	}{
		{"empty slice", []uint16{}},
		{"single element", []uint16{42}},
		{"two elements", []uint16{1, 2}},
		{"multiple elements", []uint16{1, 2, 3, 4, 5}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rng := rand.New(rand.NewSource(42))

			original := make([]uint16, len(tt.data))
			copy(original, tt.data)

			testData := make([]uint16, len(tt.data))
			copy(testData, tt.data)

			ShuffleUint16sWithRand(testData, rng)

			// 检查长度
			if len(testData) != len(original) {
				t.Errorf("ShuffleUint16s() length changed: got %d, want %d", len(testData), len(original))
			}

			// 检查元素
			originalSorted := make([]uint16, len(original))
			copy(originalSorted, original)
			sort.Slice(originalSorted, func(i, j int) bool { return originalSorted[i] < originalSorted[j] })

			testDataSorted := make([]uint16, len(testData))
			copy(testDataSorted, testData)
			sort.Slice(testDataSorted, func(i, j int) bool { return testDataSorted[i] < testDataSorted[j] })

			if !reflect.DeepEqual(originalSorted, testDataSorted) {
				t.Errorf("ShuffleUint16s() elements changed: got %v, want %v", testData, original)
			}
		})
	}
}

func TestContainsString(t *testing.T) {
	tests := []struct {
		name   string
		slice  []string
		search string
		want   bool
	}{
		{"empty slice", []string{}, "hello", false},
		{"single element found", []string{"hello"}, "hello", true},
		{"single element not found", []string{"hello"}, "world", false},
		{"multiple elements found", []string{"a", "b", "c", "hello", "d"}, "hello", true},
		{"multiple elements not found", []string{"a", "b", "c"}, "hello", false},
		{"case sensitive", []string{"Hello", "WORLD"}, "hello", false},
		{"duplicate elements", []string{"a", "a", "b", "b", "c"}, "b", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ContainsString(tt.slice, tt.search)
			if got != tt.want {
				t.Errorf("ContainsString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUniqueStrings(t *testing.T) {
	tests := []struct {
		name  string
		input []string
		want  []string
	}{
		{"empty slice", []string{}, []string{}},
		{"no duplicates", []string{"a", "b", "c"}, []string{"a", "b", "c"}},
		{"with duplicates", []string{"a", "b", "a", "c", "b", "d"}, []string{"a", "b", "c", "d"}},
		{"all duplicates", []string{"a", "a", "a"}, []string{"a"}},
		{"single element", []string{"hello"}, []string{"hello"}},
		{"empty strings", []string{"", "a", "", "b", ""}, []string{"", "a", "b"}},
		{"case sensitive", []string{"Hello", "hello", "HELLO"}, []string{"Hello", "hello", "HELLO"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := UniqueStrings(tt.input)

			// 由于顺序可能不确定，我们比较排序后的结果
			wantSorted := make([]string, len(tt.want))
			copy(wantSorted, tt.want)
			sort.Strings(wantSorted)

			gotSorted := make([]string, len(got))
			copy(gotSorted, got)
			sort.Strings(gotSorted)

			if !reflect.DeepEqual(gotSorted, wantSorted) {
				t.Errorf("UniqueStrings() = %v, want %v", got, tt.want)
			}

			// 检查长度
			if len(got) != len(tt.want) {
				t.Errorf("UniqueStrings() length = %d, want %d", len(got), len(tt.want))
			}
		})
	}
}

func TestEqualsStrings(t *testing.T) {
	tests := []struct {
		name string
		a    []string
		b    []string
		want bool
	}{
		{"both empty", []string{}, []string{}, true},
		{"same order", []string{"a", "b", "c"}, []string{"a", "b", "c"}, true},
		{"different order", []string{"a", "b", "c"}, []string{"c", "b", "a"}, true},
		{"different lengths", []string{"a", "b"}, []string{"a", "b", "c"}, false},
		{"different elements", []string{"a", "b", "c"}, []string{"a", "b", "d"}, false},
		{"case sensitive", []string{"Hello"}, []string{"hello"}, false},
		{"duplicates", []string{"a", "a", "b"}, []string{"a", "b", "b"}, true},
		{"subset", []string{"a", "b"}, []string{"a", "b", "c"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := EqualsStrings(tt.a, tt.b)
			if got != tt.want {
				t.Errorf("EqualsStrings() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTrimAfter(t *testing.T) {
	tests := []struct {
		name  string
		input string
		sep   byte
		want  string
	}{
		{"found separator", "hello/world", '/', "hello"},
		{"not found", "helloworld", '/', "helloworld"},
		{"multiple separators", "hello/world/test", '/', "hello"},
		{"separator at end", "hello/", '/', "hello"},
		{"separator at start", "/hello", '/', ""},
		{"empty string", "", '/', ""},
		{"different separator", "hello,world", ',', "hello"},
		{"separator is space", "hello world", ' ', "hello"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := TrimAfter(tt.input, tt.sep)
			if got != tt.want {
				t.Errorf("TrimAfter() = %v, want %v", got, tt.want)
			}
		})
	}
}

// 并发安全性测试
func TestShuffleConcurrency(t *testing.T) {
	// 测试并发调用是否安全
	data := []string{"a", "b", "c", "d", "e"}

	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			testData := make([]string, len(data))
			copy(testData, data)
			ShuffleStrings(testData)
			done <- true
		}()
	}

	// 等待所有goroutine完成
	for i := 0; i < 10; i++ {
		<-done
	}
}

// 性能测试
func BenchmarkShuffleStrings(b *testing.B) {
	data := []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j"}
	for i := 0; i < b.N; i++ {
		testData := make([]string, len(data))
		copy(testData, data)
		ShuffleStrings(testData)
	}
}

func BenchmarkUniqueStrings(b *testing.B) {
	data := []string{"a", "b", "a", "c", "b", "d", "a", "e", "f", "b"}
	for i := 0; i < b.N; i++ {
		UniqueStrings(data)
	}
}

func BenchmarkContainsString(b *testing.B) {
	data := []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j"}
	for i := 0; i < b.N; i++ {
		ContainsString(data, "e")
	}
}

func BenchmarkEqualsStrings(bb *testing.B) {
	a := []string{"a", "b", "c", "d", "e"}
	b := []string{"e", "d", "c", "b", "a"}
	for i := 0; i < bb.N; i++ {
		EqualsStrings(a, b)
	}
}

func BenchmarkTrimAfter(b *testing.B) {
	input := "hello/world/test/example"
	for i := 0; i < b.N; i++ {
		TrimAfter(input, '/')
	}
}
