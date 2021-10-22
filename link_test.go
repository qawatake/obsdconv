package main

import "testing"

func TestSplitDisplayName(t *testing.T) {
	cases := []struct {
		fullname        string
		wantIdentifier  string
		wantDisplayName string
	}{
		{fullname: "211022", wantIdentifier: "211022", wantDisplayName: ""},
		{fullname: "211022 | displayname", wantIdentifier: "211022", wantDisplayName: "displayname"},
	}
	for _, tt := range cases {
		gotIdentifier, gotDisplayName := splitDisplayName(tt.fullname)
		if gotIdentifier != tt.wantIdentifier {
			t.Errorf("[ERROR] got: %q, want: %q with input %q", gotIdentifier, tt.wantIdentifier, tt.fullname)
		}
		if gotDisplayName != tt.wantDisplayName {
			t.Errorf("[ERROR] got: %q, want: %q with input %q", gotDisplayName, tt.wantDisplayName, tt.fullname)
		}
	}
}

func TestSplitFragments(t *testing.T) {
	cases := []struct {
		identifier    string
		wantFileId    string
		wantFragments []string
		wantErr       error
	}{
		{identifier: "211022", wantFileId: "211022", wantFragments: nil, wantErr: nil},
		{identifier: "211022#section", wantFileId: "211022", wantFragments: []string{"section"}, wantErr: nil},
		{identifier: "211022#section#subsection", wantFileId: "211022", wantFragments: []string{"section", "subsection"}, wantErr: nil},
		{identifier: "#section", wantFileId: "", wantFragments: []string{"section"}, wantErr: nil},
		{identifier: "# section", wantFileId: "", wantFragments: []string{"section"}, wantErr: nil},
		{identifier: "# section # subsection", wantFileId: "", wantFragments: []string{"section", "subsection"}, wantErr: nil},
		{identifier: "211022# section", wantFileId: "211022", wantFragments: []string{"section"}, wantErr: nil},
		{identifier: "211022# section # subsection", wantFileId: "211022", wantFragments: []string{"section", "subsection"}, wantErr: nil},
		{identifier: "211022 #section", wantFileId: "", wantFragments: nil, wantErr: newErr(ErrKindInvalidInternalLinkContent)},
		{identifier: "211022\t#section #subsection", wantFileId: "", wantFragments: nil, wantErr: newErr(ErrKindInvalidInternalLinkContent)},
	}

	for _, tt := range cases {
		gotFileId, gotFragments, gotErr := splitFragments(tt.identifier)
		if gotErr == nil && tt.wantErr != nil {
			t.Errorf("[ERROR] must fail with input %v", tt.identifier)
			continue
		} else if gotErr != nil && tt.wantErr == nil {
			t.Errorf("[ERROR] mustn't fail with input %v: %v", tt.identifier, gotErr)
			continue
		} else if gotErr != nil && tt.wantErr != nil {
			if ee, ok := gotErr.(ErrInvalidInternalLinkContent); !ok || !ee.IsErrInvalidInternalLinkContent() {
				t.Errorf("[ERROR] unexpected error occurred: %v", ee)
				continue
			}
		}

		if gotFileId != tt.wantFileId {
			t.Errorf("[ERROR] got: %q, want: %q with %q", gotFileId, tt.wantFileId, tt.identifier)
		}

		eq := true
		if len(gotFragments) != len(tt.wantFragments) {
			eq = false
		} else {
			for id := 0; id < len(gotFragments); id++ {
				if gotFragments[id] != tt.wantFragments[id] {
					eq = false
					break
				}
			}
		}
		if !eq {
			t.Errorf("[ERROR] got: %q, want: %q with %q", gotFragments, tt.wantFragments, tt.identifier)
		}
	}
}
