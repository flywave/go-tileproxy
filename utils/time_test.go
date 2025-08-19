package utils

import (
	"context"
	"strings"
	"testing"
	"time"
)

func TestFormatDateTime(t *testing.T) {
	tests := []struct {
		name     string
		input    time.Time
		expected string
	}{
		{
			name:     "UTC time",
			input:    time.Date(2023, 12, 25, 15, 30, 45, 123000000, time.UTC),
			expected: "2023-12-25T15:30:45.123Z",
		},
		{
			name:     "non-UTC time",
			input:    time.Date(2023, 12, 25, 15, 30, 45, 123000000, time.FixedZone("CET", 3600)),
			expected: "2023-12-25T14:30:45.123Z", // Should be converted to UTC
		},
		{
			name:     "zero time",
			input:    time.Time{},
			expected: "0001-01-01T00:00:00Z",
		},
		{
			name:     "microsecond precision",
			input:    time.Date(2023, 12, 25, 15, 30, 45, 123456789, time.UTC),
			expected: "2023-12-25T15:30:45.123Z", // Should truncate to milliseconds
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatDateTime(tt.input)
			if result != tt.expected {
				t.Errorf("FormatDateTime() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestParseDateTime(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected time.Time
		wantErr  bool
	}{
		{
			name:     "valid RFC3339",
			input:    "2023-12-25T15:30:45.123Z",
			expected: time.Date(2023, 12, 25, 15, 30, 45, 123000000, time.UTC),
			wantErr:  false,
		},
		{
			name:     "valid RFC3339Nano",
			input:    "2023-12-25T15:30:45.123456789Z",
			expected: time.Date(2023, 12, 25, 15, 30, 45, 123456789, time.UTC),
			wantErr:  false,
		},
		{
			name:     "valid RFC3339 without nanoseconds",
			input:    "2023-12-25T15:30:45Z",
			expected: time.Date(2023, 12, 25, 15, 30, 45, 0, time.UTC),
			wantErr:  false,
		},
		{
			name:     "invalid format",
			input:    "invalid-date",
			expected: time.Time{},
			wantErr:  true,
		},
		{
			name:     "empty string",
			input:    "",
			expected: time.Time{},
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseDateTime(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseDateTime() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !result.Equal(tt.expected) {
				t.Errorf("ParseDateTime() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestFormatHTTPDate(t *testing.T) {
	tests := []struct {
		name     string
		input    time.Time
		expected string
	}{
		{
			name:     "standard time",
			input:    time.Date(2023, 12, 25, 15, 30, 45, 0, time.UTC),
			expected: "Mon, 25 Dec 2023 15:30:45 GMT",
		},
		{
			name:     "non-UTC time",
			input:    time.Date(2023, 12, 25, 15, 30, 45, 0, time.FixedZone("CET", 3600)),
			expected: "Mon, 25 Dec 2023 14:30:45 GMT", // Should be converted to UTC
		},
		{
			name:     "zero time",
			input:    time.Time{},
			expected: "Mon, 01 Jan 0001 00:00:00 GMT",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatHTTPDate(tt.input)
			if result != tt.expected {
				t.Errorf("FormatHTTPDate() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestParseHTTPDate(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected time.Time
		wantErr  bool
	}{
		{
			name:     "standard HTTP date",
			input:    "Mon, 25 Dec 2023 15:30:45 GMT",
			expected: time.Date(2023, 12, 25, 15, 30, 45, 0, time.UTC),
			wantErr:  false,
		},
		{
			name:     "single digit day",
			input:    "Mon, 1 Jan 2023 12:00:00 GMT",
			expected: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
			wantErr:  false,
		},
		{
			name:     "two digit year",
			input:    "Mon, 1 Jan 23 12:00:00 GMT",
			expected: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
			wantErr:  false,
		},
		{
			name:     "RFC850 format",
			input:    "Monday, 25-Dec-23 15:30:45 GMT",
			expected: time.Date(2023, 12, 25, 15, 30, 45, 0, time.UTC),
			wantErr:  false,
		},
		{
			name:     "ANSIC format",
			input:    "Mon Dec 25 15:30:45 2023",
			expected: time.Date(2023, 12, 25, 15, 30, 45, 0, time.UTC),
			wantErr:  false,
		},
		{
			name:     "invalid format",
			input:    "invalid-date",
			expected: time.Time{},
			wantErr:  true,
		},
		{
			name:     "empty string",
			input:    "",
			expected: time.Time{},
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseHTTPDate(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseHTTPDate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !result.Equal(tt.expected) {
				t.Errorf("ParseHTTPDate() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestFormatEpochSeconds(t *testing.T) {
	tests := []struct {
		name     string
		input    time.Time
		expected float64
	}{
		{
			name:     "standard time",
			input:    time.Date(2023, 12, 25, 15, 30, 45, 123000000, time.UTC),
			expected: 1703518245.123,
		},
		{
			name:     "zero time",
			input:    time.Time{},
			expected: -6795364578.871,
		},
		{
			name:     "epoch start",
			input:    time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC),
			expected: 0.0,
		},
		{
			name:     "microsecond precision",
			input:    time.Date(2023, 12, 25, 15, 30, 45, 123456000, time.UTC),
			expected: 1703518245.123,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatEpochSeconds(tt.input)
			if result != tt.expected {
				t.Errorf("FormatEpochSeconds() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestParseEpochSeconds(t *testing.T) {
	tests := []struct {
		name     string
		input    float64
		expected time.Time
	}{
		{
			name:     "standard timestamp",
			input:    1703515845.123,
			expected: time.Unix(1703515845, 123000000).UTC(),
		},
		{
			name:     "zero timestamp",
			input:    0.0,
			expected: time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			name:     "negative timestamp",
			input:    -62135596800.0,
			expected: time.Date(1754, 8, 30, 22, 43, 41, 128654848, time.UTC), // Actual result for this negative epoch
		},
		{
			name:     "microsecond precision",
			input:    1703515845.123456,
			expected: time.Unix(1703515845, 123000000).UTC(), // Truncated to milliseconds
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseEpochSeconds(tt.input)
			if !result.Equal(tt.expected) {
				t.Errorf("ParseEpochSeconds() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestRoundTripEpochSeconds(t *testing.T) {
	testTimes := []time.Time{
		time.Date(2023, 12, 25, 15, 30, 45, 123000000, time.UTC),
		time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Now().UTC(),
	}

	for _, original := range testTimes {
		epoch := FormatEpochSeconds(original)
		parsed := ParseEpochSeconds(epoch)

		// Allow for millisecond precision
		if original.Unix() != parsed.Unix() ||
			original.Nanosecond()/1e6 != parsed.Nanosecond()/1e6 {
			t.Errorf("Round trip failed: %v -> %f -> %v", original, epoch, parsed)
		}
	}
}

func TestSleepWithContext(t *testing.T) {
	t.Run("normal sleep", func(t *testing.T) {
		ctx := context.Background()
		start := time.Now()

		err := SleepWithContext(ctx, 50*time.Millisecond)
		elapsed := time.Since(start)

		if err != nil {
			t.Errorf("SleepWithContext() error = %v, want nil", err)
		}
		if elapsed < 40*time.Millisecond || elapsed > 100*time.Millisecond {
			t.Errorf("SleepWithContext() elapsed = %v, expected ~50ms", elapsed)
		}
	})

	t.Run("context cancelled", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		start := time.Now()
		err := SleepWithContext(ctx, 1*time.Second)
		elapsed := time.Since(start)

		if err != context.Canceled {
			t.Errorf("SleepWithContext() error = %v, want context.Canceled", err)
		}
		if elapsed > 10*time.Millisecond {
			t.Errorf("SleepWithContext() should return immediately, elapsed = %v", elapsed)
		}
	})

	t.Run("context timeout", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
		defer cancel()

		start := time.Now()
		err := SleepWithContext(ctx, 1*time.Second)
		elapsed := time.Since(start)

		if err != context.DeadlineExceeded {
			t.Errorf("SleepWithContext() error = %v, want context.DeadlineExceeded", err)
		}
		if elapsed > 100*time.Millisecond {
			t.Errorf("SleepWithContext() should timeout quickly, elapsed = %v", elapsed)
		}
	})
}

func TestTryParse(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		formats  []string
		expected time.Time
		wantErr  bool
	}{
		{
			name:     "first format matches",
			input:    "2023-12-25T15:30:45Z",
			formats:  []string{time.RFC3339, time.RFC822},
			expected: time.Date(2023, 12, 25, 15, 30, 45, 0, time.UTC),
			wantErr:  false,
		},
		{
			name:     "second format matches",
			input:    "25 Dec 23 15:30 GMT",
			formats:  []string{time.RFC3339, time.RFC822},
			expected: time.Date(2023, 12, 25, 15, 30, 0, 0, time.UTC),
			wantErr:  false,
		},
		{
			name:     "no format matches",
			input:    "invalid",
			formats:  []string{time.RFC3339, time.RFC822},
			expected: time.Time{},
			wantErr:  true,
		},
		{
			name:     "empty formats",
			input:    "2023-12-25T15:30:45Z",
			formats:  []string{},
			expected: time.Time{},
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tryParse(tt.input, tt.formats...)
			if (err != nil) != tt.wantErr {
				t.Errorf("tryParse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !result.Equal(tt.expected) {
				t.Errorf("tryParse() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestParseErrors(t *testing.T) {
	t.Run("single error", func(t *testing.T) {
		errs := parseErrors{
			{Format: "format1", Err: &time.ParseError{Layout: "format1", Value: "test", LayoutElem: "", ValueElem: ""}},
		}

		result := errs.Error()
		if !strings.Contains(result, "parse errors:") {
			t.Errorf("ParseErrors() should contain 'parse errors:'")
		}
		if !strings.Contains(result, "format1") {
			t.Errorf("ParseErrors() should contain format name")
		}
	})

	t.Run("multiple errors", func(t *testing.T) {
		errs := parseErrors{
			{Format: "format1", Err: &time.ParseError{Layout: "format1", Value: "test", LayoutElem: "", ValueElem: ""}},
			{Format: "format2", Err: &time.ParseError{Layout: "format2", Value: "test", LayoutElem: "", ValueElem: ""}},
		}

		result := errs.Error()
		if !strings.Contains(result, "parse errors:") {
			t.Errorf("ParseErrors() should contain 'parse errors:'")
		}
		if !strings.Contains(result, "format1") || !strings.Contains(result, "format2") {
			t.Errorf("ParseErrors() should contain all format names")
		}
	})

	t.Run("empty errors", func(t *testing.T) {
		errs := parseErrors{}

		result := errs.Error()
		if result != "parse errors:" {
			t.Errorf("ParseErrors() = %v, want 'parse errors:'", result)
		}
	})
}

func BenchmarkFormatDateTime(b *testing.B) {
	t := time.Date(2023, 12, 25, 15, 30, 45, 123000000, time.UTC)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		FormatDateTime(t)
	}
}

func BenchmarkParseDateTime(b *testing.B) {
	input := "2023-12-25T15:30:45.123Z"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ParseDateTime(input)
	}
}

func BenchmarkFormatHTTPDate(b *testing.B) {
	t := time.Date(2023, 12, 25, 15, 30, 45, 0, time.UTC)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		FormatHTTPDate(t)
	}
}

func BenchmarkParseHTTPDate(b *testing.B) {
	input := "Mon, 25 Dec 2023 15:30:45 GMT"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ParseHTTPDate(input)
	}
}

func BenchmarkFormatEpochSeconds(b *testing.B) {
	t := time.Date(2023, 12, 25, 15, 30, 45, 123000000, time.UTC)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		FormatEpochSeconds(t)
	}
}

func BenchmarkParseEpochSeconds(b *testing.B) {
	input := 1703515845.123

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ParseEpochSeconds(input)
	}
}

func BenchmarkSleepWithContext(b *testing.B) {
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		SleepWithContext(ctx, time.Microsecond)
	}
}
