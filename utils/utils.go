package utils

import (
	"fmt"
	"image/color"
	"os"
)

func FileExists(filename string) (bool, error) {
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return false, nil
	} else if err != nil {
		return false, err
	}
	return true, nil
}

func IsDir(path string) (bool, error) {
	stat, err := os.Stat(path)
	if err != nil {
		return false, err
	}
	return stat.IsDir(), nil
}

func IsSymlink(path string) (bool, error) {
	fi, err := os.Lstat(path)
	if err != nil {
		return false, err
	}
	return fi.Mode()&os.ModeSymlink != 0, nil
}

func ColorRGBFromHex(hexStr string) color.Color {
	c := &color.RGBA{}
	if (len(hexStr) != 4 && len(hexStr) != 7) || hexStr[0] != '#' {
		return c
	}

	var r, g, b int
	if len(hexStr) == 4 {
		var tmp1, tmp2, tmp3 int
		n, err := fmt.Sscanf(hexStr, "#%1x%1x%1x", &tmp1, &tmp2, &tmp3)

		if err != nil {
			return c
		}
		if n != 3 {
			return c
		}

		r = tmp1*16 + tmp1
		g = tmp2*16 + tmp2
		b = tmp3*16 + tmp3
	} else {
		n, err := fmt.Sscanf(hexStr, "#%2x%2x%2x", &r, &g, &b)
		if err != nil {
			return c
		}
		if n != 3 {
			return c
		}
	}

	c.R = uint8(r)
	c.G = uint8(g)
	c.B = uint8(b)

	return c
}
