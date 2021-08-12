package raster

import (
	"os"
	"testing"
)

func TestLerc(t *testing.T) {
	lercio := &LercIO{Mode: BORDER_UNILATERAL}

	f, _ := os.Open("../data/title_11_806_1697.atm")
	tiledata2, _ := lercio.Decode(f)

	if tiledata2 == nil {
		t.FailNow()
	}

	lercio2 := &LercIO{Mode: BORDER_UNILATERAL}

	data, _ := lercio2.Encode(tiledata2)

	f, _ = os.Create("./data.atm")
	f.Write(data)
	f.Close()
}
