package worker

import (
	"reflect"
	"testing"
)

func TestChunkTranscript(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		text      string
		maxTokens int
		overlap   int
		want      []string
	}{
		{
			name:      "empty transcript",
			text:      "",
			maxTokens: 3,
			overlap:   1,
			want:      nil,
		},
		{
			name:      "fits in single chunk",
			text:      "one two",
			maxTokens: 3,
			overlap:   1,
			want:      []string{"one two"},
		},
		{
			name:      "multiple chunks with overlap",
			text:      "one two three four five six seven",
			maxTokens: 3,
			overlap:   1,
			want: []string{
				"one two three",
				"three four five",
				"five six seven",
			},
		},
		{
			name:      "overlap larger than chunk size still progresses",
			text:      "a b c d",
			maxTokens: 2,
			overlap:   3,
			want: []string{
				"a b",
				"b c",
				"c d",
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := ChunkTranscript(tc.text, tc.maxTokens, tc.overlap)
			if !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("ChunkTranscript() = %#v, want %#v", got, tc.want)
			}
		})
	}
}
