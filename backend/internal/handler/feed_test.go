package handler

import (
	"net/http/httptest"
	"testing"
)

func TestParsePage(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		rawQuery string
		want     int
	}{
		{name: "missing page", rawQuery: "", want: 1},
		{name: "valid page", rawQuery: "page=3", want: 3},
		{name: "zero page", rawQuery: "page=0", want: 1},
		{name: "negative page", rawQuery: "page=-2", want: 1},
		{name: "invalid page", rawQuery: "page=abc", want: 1},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequest("GET", "/api/feed?"+tt.rawQuery, nil)
			if got := parsePage(req); got != tt.want {
				t.Fatalf("parsePage() = %d, want %d", got, tt.want)
			}
		})
	}
}
