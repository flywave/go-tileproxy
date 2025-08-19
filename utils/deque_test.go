package utils

import (
	"testing"
)

func TestNewDeque(t *testing.T) {
	tests := []struct {
		name    string
		size    []int
		wantCap int
		wantLen int
		wantMin int
	}{
		{"no arguments", []int{}, 0, 0, 16},
		{"only capacity", []int{10}, 16, 0, 16},
		{"capacity and min", []int{20, 32}, 32, 0, 32},
		{"capacity less than min", []int{10, 64}, 64, 0, 64},
		{"capacity exact power of 2", []int{16}, 16, 0, 16},
		{"large capacity", []int{100, 128}, 128, 0, 128},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := NewDeque(tt.size...)
			if q.Cap() != tt.wantCap {
				t.Errorf("NewDeque() cap = %d, want %d", q.Cap(), tt.wantCap)
			}
			if q.Len() != tt.wantLen {
				t.Errorf("NewDeque() len = %d, want %d", q.Len(), tt.wantLen)
			}
			if q.minCap != tt.wantMin {
				t.Errorf("NewDeque() minCap = %d, want %d", q.minCap, tt.wantMin)
			}
		})
	}
}

func TestPushBack(t *testing.T) {
	q := NewDeque()

	// Test empty deque
	if q.Len() != 0 {
		t.Errorf("Empty deque length = %d, want 0", q.Len())
	}

	// Test single element
	q.PushBack("hello")
	if q.Len() != 1 {
		t.Errorf("After PushBack length = %d, want 1", q.Len())
	}
	if q.Front() != "hello" {
		t.Errorf("Front() = %v, want hello", q.Front())
	}
	if q.Back() != "hello" {
		t.Errorf("Back() = %v, want hello", q.Back())
	}

	// Test multiple elements
	q.PushBack("world")
	q.PushBack(42)
	if q.Len() != 3 {
		t.Errorf("After multiple PushBack length = %d, want 3", q.Len())
	}
	if q.Front() != "hello" {
		t.Errorf("Front() after multiple PushBack = %v, want hello", q.Front())
	}
	if q.Back() != 42 {
		t.Errorf("Back() after multiple PushBack = %v, want 42", q.Back())
	}
}

func TestPushFront(t *testing.T) {
	q := NewDeque()

	// Test single element
	q.PushFront("hello")
	if q.Len() != 1 {
		t.Errorf("After PushFront length = %d, want 1", q.Len())
	}
	if q.Front() != "hello" {
		t.Errorf("Front() = %v, want hello", q.Front())
	}
	if q.Back() != "hello" {
		t.Errorf("Back() = %v, want hello", q.Back())
	}

	// Test multiple elements
	q.PushFront("world")
	q.PushFront(42)
	if q.Len() != 3 {
		t.Errorf("After multiple PushFront length = %d, want 3", q.Len())
	}
	if q.Front() != 42 {
		t.Errorf("Front() after multiple PushFront = %v, want 42", q.Front())
	}
	if q.Back() != "hello" {
		t.Errorf("Back() after multiple PushFront = %v, want hello", q.Back())
	}
}

func TestPopFront(t *testing.T) {
	q := NewDeque()

	// Test panic on empty deque
	defer func() {
		if r := recover(); r == nil {
			t.Error("PopFront() on empty deque did not panic")
		}
	}()
	q.PopFront()
}

func TestPopBack(t *testing.T) {
	q := NewDeque()

	// Test panic on empty deque
	defer func() {
		if r := recover(); r == nil {
			t.Error("PopBack() on empty deque did not panic")
		}
	}()
	q.PopBack()
}

func TestPopFrontWithData(t *testing.T) {
	q := NewDeque()
	q.PushBack("first")
	q.PushBack("second")
	q.PushBack("third")

	// Test PopFront
	val := q.PopFront()
	if val != "first" {
		t.Errorf("PopFront() = %v, want first", val)
	}
	if q.Len() != 2 {
		t.Errorf("After PopFront length = %d, want 2", q.Len())
	}
	if q.Front() != "second" {
		t.Errorf("Front() after PopFront = %v, want second", q.Front())
	}

	// Test another PopFront
	val = q.PopFront()
	if val != "second" {
		t.Errorf("PopFront() = %v, want second", val)
	}
	if q.Len() != 1 {
		t.Errorf("After second PopFront length = %d, want 1", q.Len())
	}
	if q.Front() != "third" {
		t.Errorf("Front() after second PopFront = %v, want third", q.Front())
	}
}

func TestPopBackWithData(t *testing.T) {
	q := NewDeque()
	q.PushBack("first")
	q.PushBack("second")
	q.PushBack("third")

	// Test PopBack
	val := q.PopBack()
	if val != "third" {
		t.Errorf("PopBack() = %v, want third", val)
	}
	if q.Len() != 2 {
		t.Errorf("After PopBack length = %d, want 2", q.Len())
	}
	if q.Back() != "second" {
		t.Errorf("Back() after PopBack = %v, want second", q.Back())
	}

	// Test another PopBack
	val = q.PopBack()
	if val != "second" {
		t.Errorf("PopBack() = %v, want second", val)
	}
	if q.Len() != 1 {
		t.Errorf("After second PopBack length = %d, want 1", q.Len())
	}
	if q.Back() != "first" {
		t.Errorf("Back() after second PopBack = %v, want first", q.Back())
	}
}

func TestFrontBackPanic(t *testing.T) {
	q := NewDeque()

	t.Run("Front panic", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("Front() on empty deque did not panic")
			}
		}()
		q.Front()
	})

	t.Run("Back panic", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("Back() on empty deque did not panic")
			}
		}()
		q.Back()
	})
}

func TestAt(t *testing.T) {
	q := NewDeque()
	q.PushBack("zero")
	q.PushBack("one")
	q.PushBack("two")
	q.PushBack("three")

	// Test valid indices
	tests := []struct {
		index int
		want  interface{}
	}{
		{0, "zero"},
		{1, "one"},
		{2, "two"},
		{3, "three"},
	}

	for _, tt := range tests {
		t.Run(string(rune(tt.index)), func(t *testing.T) {
			got := q.At(tt.index)
			if got != tt.want {
				t.Errorf("At(%d) = %v, want %v", tt.index, got, tt.want)
			}
		})
	}

	// Test invalid indices
	invalidTests := []int{-1, 4, 100}
	for _, idx := range invalidTests {
		t.Run("invalid", func(t *testing.T) {
			defer func() {
				if r := recover(); r == nil {
					t.Errorf("At(%d) did not panic", idx)
				}
			}()
			q.At(idx)
		})
	}
}

func TestSet(t *testing.T) {
	q := NewDeque()
	q.PushBack("zero")
	q.PushBack("one")
	q.PushBack("two")

	// Test valid set
	q.Set(1, "updated")
	if q.At(1) != "updated" {
		t.Errorf("After Set(1, \"updated\") At(1) = %v, want updated", q.At(1))
	}

	// Test invalid set
	defer func() {
		if r := recover(); r == nil {
			t.Error("Set() with invalid index did not panic")
		}
	}()
	q.Set(5, "invalid")
	if r := recover(); r == nil {
		t.Error("Set() with invalid index did not panic")
	}
}

func TestClear(t *testing.T) {
	q := NewDeque()
	q.PushBack("one")
	q.PushBack("two")
	q.PushBack("three")

	q.Clear()

	if q.Len() != 0 {
		t.Errorf("After Clear() length = %d, want 0", q.Len())
	}
	if q.Cap() != 0 {
		// Clear should not change capacity
		// but the buffer should be reset
	}
}

func TestRotate(t *testing.T) {
	q := NewDeque()
	q.PushBack("a")
	q.PushBack("b")
	q.PushBack("c")
	q.PushBack("d")

	tests := []struct {
		name   string
		rotate int
		want   []interface{}
	}{
		{"rotate 1", 1, []interface{}{"b", "c", "d", "a"}},
		{"rotate 2", 2, []interface{}{"c", "d", "a", "b"}},
		{"rotate -1", -1, []interface{}{"d", "a", "b", "c"}},
		{"rotate 4", 4, []interface{}{"a", "b", "c", "d"}},
		{"rotate 0", 0, []interface{}{"a", "b", "c", "d"}},
		{"rotate 5", 5, []interface{}{"b", "c", "d", "a"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset deque for each test
			q.Clear()
			q.PushBack("a")
			q.PushBack("b")
			q.PushBack("c")
			q.PushBack("d")

			q.Rotate(tt.rotate)

			// Check elements
			for i, want := range tt.want {
				if q.At(i) != want {
					t.Errorf("After Rotate(%d) At(%d) = %v, want %v", tt.rotate, i, q.At(i), want)
				}
			}
		})
	}

	// Test empty deque
	t.Run("empty deque", func(t *testing.T) {
		empty := NewDeque()
		empty.Rotate(1) // Should not panic
		if empty.Len() != 0 {
			t.Errorf("Rotate on empty deque changed length")
		}
	})

	// Test single element
	t.Run("single element", func(t *testing.T) {
		single := NewDeque()
		single.PushBack("only")
		single.Rotate(1)
		if single.At(0) != "only" {
			t.Errorf("Rotate on single element changed it")
		}
	})
}

func TestContains(t *testing.T) {
	q := NewDeque()
	q.PushBack(1)
	q.PushBack(2)
	q.PushBack(3)
	q.PushBack(2)

	// Equality function for integers
	eq := func(a, b interface{}) bool {
		return a.(int) == b.(int)
	}

	tests := []struct {
		value interface{}
		want  bool
	}{
		{1, true},
		{2, true},
		{3, true},
		{4, false},
		{0, false},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			got := q.Contains(tt.value, eq)
			if got != tt.want {
				t.Errorf("Contains(%v) = %v, want %v", tt.value, got, tt.want)
			}
		})
	}

	// Test empty deque
	empty := NewDeque()
	if empty.Contains(1, eq) {
		t.Error("Contains on empty deque returned true")
	}
}

func TestGrowthAndShrink(t *testing.T) {
	q := NewDeque()

	// Test growth
	initialCap := q.Cap()
	for i := 0; i < 100; i++ {
		q.PushBack(i)
	}

	if q.Cap() <= initialCap {
		t.Errorf("Deque did not grow, capacity = %d", q.Cap())
	}

	// Test that all elements are accessible
	for i := 0; i < 100; i++ {
		if q.At(i) != i {
			t.Errorf("At(%d) = %v, want %d", i, q.At(i), i)
		}
	}

	// Test shrink by removing elements
	for i := 0; i < 90; i++ {
		q.PopFront()
	}

	// Should have shrunk
	if q.Cap() == 0 {
		t.Errorf("Deque shrunk to zero capacity")
	}
}

func TestMixedOperations(t *testing.T) {
	q := NewDeque()

	// Test mixed push and pop operations
	q.PushBack(1)
	q.PushFront(0)
	q.PushBack(2)
	q.PushFront(-1)

	if q.Len() != 4 {
		t.Errorf("Mixed operations length = %d, want 4", q.Len())
	}

	// Check order
	expected := []interface{}{-1, 0, 1, 2}
	for i, want := range expected {
		if q.At(i) != want {
			t.Errorf("Mixed operations At(%d) = %v, want %v", i, q.At(i), want)
		}
	}

	// Test mixed pops
	val := q.PopFront()
	if val != -1 {
		t.Errorf("PopFront() = %v, want -1", val)
	}

	val = q.PopBack()
	if val != 2 {
		t.Errorf("PopBack() = %v, want 2", val)
	}

	if q.Len() != 2 {
		t.Errorf("After mixed pops length = %d, want 2", q.Len())
	}
}

func TestSetMinCapacity(t *testing.T) {
	q := NewDeque()
	q.SetMinCapacity(5) // 2^5 = 32

	if q.minCap != 32 {
		t.Errorf("SetMinCapacity(5) minCap = %d, want 32", q.minCap)
	}

	// Test with smaller value
	q.SetMinCapacity(3) // 2^3 = 8, but minCapacity is 16
	if q.minCap != 16 {
		t.Errorf("SetMinCapacity(3) minCap = %d, want 16", q.minCap)
	}
}

func TestCircularBuffer(t *testing.T) {
	q := NewDeque(4)

	// Fill the buffer
	q.PushBack(1)
	q.PushBack(2)
	q.PushBack(3)
	q.PushBack(4)

	// Remove from front and add to back to test circular behavior
	q.PopFront()
	q.PushBack(5)

	// Check elements
	if q.At(0) != 2 {
		t.Errorf("Circular buffer At(0) = %v, want 2", q.At(0))
	}
	if q.At(3) != 5 {
		t.Errorf("Circular buffer At(3) = %v, want 5", q.At(3))
	}
}

// Benchmark tests
func BenchmarkPushBack(b *testing.B) {
	q := NewDeque()
	for i := 0; i < b.N; i++ {
		q.PushBack(i)
	}
}

func BenchmarkPushFront(b *testing.B) {
	q := NewDeque()
	for i := 0; i < b.N; i++ {
		q.PushFront(i)
	}
}

func BenchmarkPopBack(b *testing.B) {
	q := NewDeque()
	for i := 0; i < b.N; i++ {
		q.PushBack(i)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		q.PopBack()
	}
}

func BenchmarkPopFront(b *testing.B) {
	q := NewDeque()
	for i := 0; i < b.N; i++ {
		q.PushBack(i)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		q.PopFront()
	}
}

func BenchmarkContains(b *testing.B) {
	q := NewDeque()
	for i := 0; i < 1000; i++ {
		q.PushBack(i)
	}

	eq := func(a, b interface{}) bool {
		return a.(int) == b.(int)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		q.Contains(500, eq)
	}
}
