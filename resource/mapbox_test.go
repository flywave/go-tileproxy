package resource

import (
	"io"
	"io/ioutil"
	"os"
	"testing"
)

func TestE(t *testing.T) {

	f, _ := os.Open("./style.json")

	data, _ := io.ReadAll(f)

	style := ExtractStyle(data, "data.styleContent")

	ioutil.WriteFile("./mbstyle.json", style.GetData(), os.ModePerm)
}
