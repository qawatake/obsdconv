package main

import "testing"

func TestConsumeTag(t *testing.T) {
	cases := []struct {
		argRaw      []rune
		argPtr      int
		wantAdvance int
		wantTag     string
	}{
		{
			argRaw:      []rune("#todo"),
			argPtr:      0,
			wantAdvance: 5,
			wantTag:     "todo",
		},
		{
			argRaw:      []rune("#_todo"),
			argPtr:      0,
			wantAdvance: 6,
			wantTag:     "_todo",
		},
		{
			argRaw:      []rune("# todo"),
			argPtr:      0,
			wantAdvance: 0,
			wantTag:     "",
		},
		{
			argRaw:      []rune("#\ttodo"),
			argPtr:      0,
			wantAdvance: 0,
			wantTag:     "",
		},
		{
			argRaw:      []rune("\\#todo"),
			argPtr:      1,
			wantAdvance: 0,
			wantTag:     "",
		},
		{
			argRaw:      []rune("##todo"),
			argPtr:      0,
			wantAdvance: 0,
			wantTag:     "",
		},
	}

	for _, tt := range cases {
		gotAdvance, gotTag := consumeTag(tt.argRaw, tt.argPtr)
		if gotAdvance != tt.wantAdvance {
			t.Errorf("[ERROR] got: %v, want: %v", gotAdvance, tt.wantAdvance)
		}
		if gotTag != tt.wantTag {
			t.Errorf("[ERROR] got: %v, want: %v", gotTag, tt.wantTag)
		}
	}
}

func TestConsumeRepeat(t *testing.T) {
	cases := []struct {
		argRaw    []rune
		argSubstr string
		want      int
	}{
		{
			argRaw:    []rune("###x"),
			argSubstr: "#",
			want:      3,
		},
		{
			argRaw:    []rune("----"),
			argSubstr: "-",
			want:      4,
		},
		{
			argRaw:    []rune("$$x$$"),
			argSubstr: "$",
			want:      2,
		},
	}

	for _, tt := range cases {
		if got := consumeRepeat(tt.argRaw, 0, tt.argSubstr); got != tt.want {
			t.Errorf("[ERROR] got: %v, want: %v", got, tt.want)
		}
	}
}
