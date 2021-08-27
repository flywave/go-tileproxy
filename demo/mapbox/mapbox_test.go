package mapbox

import "testing"

func TestDataset(t *testing.T) {
	dataset := GetDataset()
	if dataset == nil {
		t.FailNow()
	}
}
