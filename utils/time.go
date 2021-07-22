package utils

import (
	"context"
	"fmt"
	"math/big"
	"strings"
	"time"
)

const (
	dateTimeFormatInput                      = "2006-01-02T15:04:05.999999999Z"
	dateTimeFormatOutput                     = "2006-01-02T15:04:05.999Z"
	httpDateFormat                           = "Mon, 02 Jan 2006 15:04:05 GMT"
	httpDateFormatSingleDigitDay             = "Mon, _2 Jan 2006 15:04:05 GMT"
	httpDateFormatSingleDigitDayTwoDigitYear = "Mon, _2 Jan 06 15:04:05 GMT"
)

var millisecondFloat = big.NewFloat(1e3)

func FormatDateTime(value time.Time) string {
	return value.UTC().Format(dateTimeFormatOutput)
}

func ParseDateTime(value string) (time.Time, error) {
	return tryParse(value,
		dateTimeFormatInput,
		time.RFC3339Nano,
		time.RFC3339,
	)
}

func FormatHTTPDate(value time.Time) string {
	return value.UTC().Format(httpDateFormat)
}

func ParseHTTPDate(value string) (time.Time, error) {
	return tryParse(value,
		httpDateFormat,
		httpDateFormatSingleDigitDay,
		httpDateFormatSingleDigitDayTwoDigitYear,
		time.RFC850,
		time.ANSIC,
	)
}

func FormatEpochSeconds(value time.Time) float64 {
	ms := value.UnixNano() / int64(time.Millisecond)
	return float64(ms) / 1e3
}

func ParseEpochSeconds(value float64) time.Time {
	f := big.NewFloat(value)
	f = f.Mul(f, millisecondFloat)
	i, _ := f.Int64()
	return time.Unix(0, i*1e6).UTC()
}

func tryParse(v string, formats ...string) (time.Time, error) {
	var errs parseErrors
	for _, f := range formats {
		t, err := time.Parse(f, v)
		if err != nil {
			errs = append(errs, parseError{
				Format: f,
				Err:    err,
			})
			continue
		}
		return t, nil
	}

	return time.Time{}, fmt.Errorf("unable to parse time string, %w", errs)
}

type parseErrors []parseError

func (es parseErrors) Error() string {
	var s strings.Builder
	for _, e := range es {
		fmt.Fprintf(&s, "\n * %q: %v", e.Format, e.Err)
	}

	return "parse errors:" + s.String()
}

type parseError struct {
	Format string
	Err    error
}

func SleepWithContext(ctx context.Context, dur time.Duration) error {
	t := time.NewTimer(dur)
	defer t.Stop()

	select {
	case <-t.C:
		break
	case <-ctx.Done():
		return ctx.Err()
	}

	return nil
}
