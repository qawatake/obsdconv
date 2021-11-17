package scan

import "testing"

func TestRuneIndex(t *testing.T) {
	cases := []struct{
		s string
		substr string
		wantId int
	}{
		{
			s: "[[abc]]",
			substr: "]]",
			wantId: 5,
		},
		{
			s: "[[abc]]",
			substr: "]]]",
			wantId: -1,
		},
		{
			s: "[[abc]]",
			substr: "[[",
			wantId: 0,
		},
	}

	for _, tt := range cases {
		gotId := indexInRunes([]rune(tt.s), tt.substr)
		if gotId != tt.wantId {
			t.Errorf("[ERROR] got: %d, want: %d with s: %q, substr: %q", gotId, tt.wantId, tt.s, tt.substr)
		}
	}
}
