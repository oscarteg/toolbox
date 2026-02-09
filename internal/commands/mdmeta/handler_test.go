package mdmeta

import (
	"testing"
	"time"
)

func TestParseDate(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    time.Time
		wantErr bool
	}{
		{
			name:    "ISO date",
			input:   "2024-01-15",
			want:    time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
			wantErr: false,
		},
		{
			name:    "ISO datetime with Z",
			input:   "2024-01-15T10:30:00Z",
			want:    time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
			wantErr: false,
		},
		{
			name:    "ISO datetime with timezone",
			input:   "2024-01-15T10:30:00+02:00",
			want:    time.Date(2024, 1, 15, 10, 30, 0, 0, time.FixedZone("", 2*60*60)),
			wantErr: false,
		},
		{
			name:    "datetime with space",
			input:   "2024-01-15 10:30:00",
			want:    time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
			wantErr: false,
		},
		{
			name:    "long month format",
			input:   "January 15, 2024",
			want:    time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
			wantErr: false,
		},
		{
			name:    "short month format",
			input:   "Jan 15, 2024",
			want:    time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
			wantErr: false,
		},
		{
			name:    "invalid date",
			input:   "not-a-date",
			wantErr: true,
		},
		{
			name:    "empty string",
			input:   "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseDate(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseDate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !got.Equal(tt.want) {
				t.Errorf("parseDate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetStringValue(t *testing.T) {
	now := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	tests := []struct {
		name     string
		metadata map[string]any
		key      string
		want     string
		wantOk   bool
	}{
		{
			name:     "string value",
			metadata: map[string]any{"date": "2024-01-15"},
			key:      "date",
			want:     "2024-01-15",
			wantOk:   true,
		},
		{
			name:     "time.Time value",
			metadata: map[string]any{"date": now},
			key:      "date",
			want:     "2024-01-15T10:30:00Z",
			wantOk:   true,
		},
		{
			name:     "int value",
			metadata: map[string]any{"count": 42},
			key:      "count",
			want:     "42",
			wantOk:   true,
		},
		{
			name:     "missing key",
			metadata: map[string]any{"date": "2024-01-15"},
			key:      "updated",
			want:     "",
			wantOk:   false,
		},
		{
			name:     "empty metadata",
			metadata: map[string]any{},
			key:      "date",
			want:     "",
			wantOk:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := getStringValue(tt.metadata, tt.key)
			if ok != tt.wantOk {
				t.Errorf("getStringValue() ok = %v, wantOk %v", ok, tt.wantOk)
				return
			}
			if got != tt.want {
				t.Errorf("getStringValue() = %v, want %v", got, tt.want)
			}
		})
	}
}
