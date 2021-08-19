package seed

import (
	"testing"
)

func TestSeedProgressCanSkip(t *testing.T) {
	test_old_progress := [][]interface{}{
		nil,
		{},
		{[2]int{0, 4}},
		{[2]int{0, 4}},
		{[2]int{1, 4}},
		{[2]int{0, 4}},
		{[2]int{0, 4}, [2]int{0, 4}, [2]int{2, 4}},
		{[2]int{0, 4}, [2]int{0, 4}, [2]int{2, 4}},
		{[2]int{0, 4}, [2]int{0, 4}, [2]int{2, 4}},
		{[2]int{0, 4}, [2]int{0, 4}, [2]int{2, 4}},
		{[2]int{0, 4}, [2]int{0, 4}, [2]int{2, 4}},
		{[2]int{0, 4}, [2]int{0, 4}, [2]int{2, 4}},
	}
	test_current_progress := [][]interface{}{
		{[2]int{0, 4}},
		{[2]int{0, 4}},
		nil,
		{[2]int{0, 4}},
		{[2]int{0, 4}},
		{[2]int{0, 4}, [2]int{0, 4}},
		{[2]int{0, 4}, [2]int{0, 4}},
		{[2]int{0, 4}, [2]int{0, 4}, [2]int{1, 4}},
		{[2]int{0, 4}, [2]int{0, 4}, [2]int{2, 4}},
		{[2]int{0, 4}, [2]int{0, 4}, [2]int{3, 4}},
		{[2]int{0, 4}, [2]int{1, 4}},
		{[2]int{0, 4}, [2]int{1, 4}, [2]int{0, 4}},
	}
	test_result := []bool{
		false,
		true,
		false,
		false,
		true,
		false,
		false,
		true,
		false,
		false,
		false,
		false,
	}

	seed := &SeedProgress{}

	for i := range test_result {
		if i == 7 {
			print("")
		}
		if test_result[i] != seed.canSkip(test_old_progress[i], test_current_progress[i]) {
			t.FailNow()
		}
	}
}

func TestCleanupProgressCanSkip(t *testing.T) {
	test_old_dir := []string{
		"",
		"",
		"/01/000/001",
		"/01/000/001",
		"/01/000/001",
		"/01/000/001",
		"/01/000/001",
	}
	test_current_dir := []string{
		"/00",
		"/00/000/000",
		"/00",
		"/01/000/000",
		"/01/000/000/000",
		"/01/000/001",
		"/01/000/001/000",
	}
	test_result := []bool{
		false,
		false,
		true,
		true,
		true,
		false,
		false,
	}

	seed := &DirectoryCleanupProgress{}

	for i := range test_result {
		if test_result[i] != seed.canSkip(test_old_dir[i], test_current_dir[i]) {
			t.FailNow()
		}
	}
}

func assert(a []interface{}, b [][2]int, t *testing.T) {
	if len(a) != len(b) {
		t.FailNow()
	}
	for i := range a {
		aa := a[i].([2]int)
		bb := b[i]

		if aa[0] != bb[0] || aa[1] != bb[1] {
			t.FailNow()
		}
	}
}

func TestSeedProgress(t *testing.T) {
	old := NewSeedProgress(nil)

	old.StepDown(0, 2, func() bool {
		old.StepDown(0, 4, func() bool {
			r1 := old.CurrentProgressIdentifier().([]interface{})
			assert(r1, [][2]int{{0, 2}, {0, 4}}, t)
			return true
		})
		r2 := old.CurrentProgressIdentifier().([]interface{})
		assert(r2, [][2]int{{0, 2}, {0, 4}}, t)
		old.StepDown(1, 4, func() bool {
			r3 := old.CurrentProgressIdentifier().([]interface{})
			assert(r3, [][2]int{{0, 2}, {1, 4}}, t)
			return true
		})
		r4 := old.CurrentProgressIdentifier().([]interface{})
		assert(r4, [][2]int{{0, 2}, {1, 4}}, t)
		return true
	})

	r5 := old.CurrentProgressIdentifier().([]interface{})
	assert(r5, [][2]int{}, t)

	old.StepDown(1, 2, func() bool {
		r6 := old.CurrentProgressIdentifier().([]interface{})
		assert(r6, [][2]int{{1, 2}}, t)
		old.StepDown(0, 4, func() bool {
			old.StepDown(1, 4, func() bool {
				r7 := old.CurrentProgressIdentifier().([]interface{})
				assert(r7, [][2]int{{1, 2}, {0, 4}, {1, 4}}, t)
				return true
			})
			return true
		})
		return true
	})
}

func TestSeedProgressAlreadyProcessed(t *testing.T) {
	new := NewSeedProgress([]interface{}{[2]int{0, 2}})
	new.StepDown(0, 2, func() bool {
		if new.AlreadyProcessed() {
			t.FailNow()
		}
		new.StepDown(0, 2, func() bool {
			if new.AlreadyProcessed() {
				t.FailNow()
			}
			return true
		})
		return true
	})

	new = NewSeedProgress([]interface{}{[2]int{1, 2}})
	new.StepDown(0, 2, func() bool {
		if !new.AlreadyProcessed() {
			t.FailNow()
		}
		new.StepDown(0, 2, func() bool {
			if !new.AlreadyProcessed() {
				t.FailNow()
			}
			return true
		})
		return true
	})

	new = NewSeedProgress([]interface{}{[2]int{0, 2}, [2]int{1, 4}, [2]int{2, 4}})
	new.StepDown(0, 2, func() bool {
		if new.AlreadyProcessed() {
			t.FailNow()
		}
		new.StepDown(0, 4, func() bool {
			if !new.AlreadyProcessed() {
				t.FailNow()
			}
			return true
		})
		new.StepDown(1, 4, func() bool {
			if new.AlreadyProcessed() {
				t.FailNow()
			}
			new.StepDown(1, 4, func() bool {
				if !new.AlreadyProcessed() {
					t.FailNow()
				}
				return true
			})
			new.StepDown(2, 4, func() bool {
				if new.AlreadyProcessed() {
					t.FailNow()
				}
				return true
			})
			new.StepDown(3, 4, func() bool {
				if new.AlreadyProcessed() {
					t.FailNow()
				}
				return true
			})
			return true
		})
		new.StepDown(2, 4, func() bool {
			if new.AlreadyProcessed() {
				t.FailNow()
			}
			return true
		})
		return true
	})

}
