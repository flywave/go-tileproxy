package draw

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
)

const (
	Radian = 1
	Degree = (math.Pi / 180) * Radian
)

func DegreesToRadians(d float64) float64 {
	return float64(d * Degree)
}

func RadiansToDegrees(a float64) float64 {
	return float64(a / Degree)
}

func ParseLatLon(s string) (float64, float64, error) {
	lat, lng, err := ParseD(s)
	if err == nil {
		return lat, lng, nil
	}

	lat, lng, err = ParseHD(s)
	if err == nil {
		return lat, lng, nil
	}

	lat, lng, err = ParseHDM(s)
	if err == nil {
		return lat, lng, nil
	}

	lat, lng, err = ParseHDMS(s)
	if err == nil {
		return lat, lng, nil
	}

	return 0, 0, fmt.Errorf("cannot parse coordinates: %s", s)
}

func ParseD(s string) (float64, float64, error) {
	re := regexp.MustCompile(`^\s*([+-]?[\d\.]+)\s*(,|;|:|\s)\s*([+-]?[\d\.]+)\s*$`)

	matches := re.FindStringSubmatch(s)
	if matches == nil {
		return 0, 0, fmt.Errorf("cannot parse 'D' string: %s", s)
	}

	lat, err := strconv.ParseFloat(matches[1], 64)
	if err != nil || lat < -90 || lat > 90 {
		return 0, 0, fmt.Errorf("cannot parse 'D' string: %s", s)
	}

	lng, err := strconv.ParseFloat(matches[3], 64)
	if err != nil || lng < -180 || lng > 180 {
		return 0, 0, fmt.Errorf("cannot parse 'D' string: %s", s)
	}

	return lat, lng, nil
}

func ParseHD(s string) (float64, float64, error) {
	re := regexp.MustCompile(`^\s*([NnSs])\s*([\d\.]+)\s+([EeWw])\s*([\d\.]+)\s*$`)

	matches := re.FindStringSubmatch(s)
	if matches == nil {
		return 0, 0, fmt.Errorf("cannot parse 'HD' string: %s", s)
	}

	lat, err := strconv.ParseFloat(matches[2], 64)
	if err != nil || lat > 90 {
		return 0, 0, fmt.Errorf("cannot parse 'HD' string: %s", s)
	}
	if matches[1] == "S" || matches[1] == "s" {
		lat = -lat
	}

	lng, err := strconv.ParseFloat(matches[4], 64)
	if err != nil || lng > 180 {
		return 0, 0, fmt.Errorf("cannot parse 'HD' string: %s", s)
	}
	if matches[3] == "W" || matches[3] == "w" {
		lng = -lng
	}

	return lat, lng, nil
}

func ParseHDM(s string) (float64, float64, error) {
	re := regexp.MustCompile(`^\s*([NnSs])\s*([\d]+)\s+([\d.]+)\s+([EeWw])\s*([\d]+)\s+([\d.]+)\s*$`)

	matches := re.FindStringSubmatch(s)
	if matches == nil {
		return 0, 0, fmt.Errorf("cannot parse 'HDM' string: %s", s)
	}

	latDeg, err := strconv.ParseFloat(matches[2], 64)
	if err != nil || latDeg > 90 {
		return 0, 0, fmt.Errorf("cannot parse 'HDM' string: %s", s)
	}
	latMin, err := strconv.ParseFloat(matches[3], 64)
	if err != nil || latMin >= 60 {
		return 0, 0, fmt.Errorf("cannot parse 'HDM' string: %s", s)
	}
	lat := latDeg + latMin/60.0
	if matches[1] == "S" || matches[1] == "s" {
		lat = -lat
	}

	lngDeg, err := strconv.ParseFloat(matches[5], 64)
	if err != nil || lngDeg > 180 {
		return 0, 0, fmt.Errorf("cannot parse 'HDM' string: %s", s)
	}
	lngMin, err := strconv.ParseFloat(matches[6], 64)
	if err != nil || lngMin >= 60 {
		return 0, 0, fmt.Errorf("cannot parse 'HDM' string: %s", s)
	}
	lng := lngDeg + lngMin/60.0
	if matches[4] == "W" || matches[4] == "w" {
		lng = -lng
	}

	return lat, lng, nil
}

func ParseHDMS(s string) (float64, float64, error) {
	re := regexp.MustCompile(`^\s*([NnSs])\s*([\d]+)\s+([\d]+)\s+([\d.]+)\s+([EeWw])\s*([\d]+)\s+([\d]+)\s+([\d.]+)\s*$`)

	matches := re.FindStringSubmatch(s)
	if matches == nil {
		return 0, 0, fmt.Errorf("cannot parse 'HDMS' string: %s", s)
	}

	latDeg, err := strconv.ParseFloat(matches[2], 64)
	if err != nil || latDeg > 90 {
		return 0, 0, fmt.Errorf("cannot parse 'HDMS' string: %s", s)
	}
	latMin, err := strconv.ParseFloat(matches[3], 64)
	if err != nil || latMin >= 60 {
		return 0, 0, fmt.Errorf("cannot parse 'HDMS' string: %s", s)
	}
	latSec, err := strconv.ParseFloat(matches[4], 64)
	if err != nil || latSec >= 60 {
		return 0, 0, fmt.Errorf("cannot parse 'HDMS' string: %s", s)
	}
	lat := latDeg + latMin/60.0 + latSec/3600.0
	if matches[1] == "S" || matches[1] == "s" {
		lat = -lat
	}

	lngDeg, err := strconv.ParseFloat(matches[6], 64)
	if err != nil || lngDeg > 180 {
		return 0, 0, fmt.Errorf("cannot parse 'HDMS' string: %s", s)
	}
	lngMin, err := strconv.ParseFloat(matches[7], 64)
	if err != nil || lngMin >= 60 {
		return 0, 0, fmt.Errorf("cannot parse 'HDMS' string: %s", s)
	}
	lngSec, err := strconv.ParseFloat(matches[8], 64)
	if err != nil || lngSec >= 60 {
		return 0, 0, fmt.Errorf("cannot parse 'HDMS' string: %s", s)
	}
	lng := lngDeg + lngMin/60.0 + lngSec/3600.0
	if matches[5] == "W" || matches[5] == "w" {
		lng = -lng
	}

	return lat, lng, nil
}
