package convert

import (
	"path/filepath"
	"testing"
)

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
		{identifier: "211022 #section", wantFileId: "", wantFragments: nil, wantErr: newErrTransform(ERR_KIND_INVALID_INTERNAL_LINK_CONTENT, "")},
		{identifier: "211022\t#section #subsection", wantFileId: "", wantFragments: nil, wantErr: newErrTransform(ERR_KIND_INVALID_INTERNAL_LINK_CONTENT, "")},
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
			if ee, ok := gotErr.(ErrTransform); !(ok || ee.Kind() == ERR_KIND_INVALID_INTERNAL_LINK_CONTENT) {
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

func TestPathMatchScore(t *testing.T) {
	cases := []struct {
		path     string
		filename string
		want     int
	}{
		{path: "test.md", filename: "test.md", want: 0},
		{path: "a/test.md", filename: "test.md", want: 1},
		{path: "a/test.md", filename: "a/test.md", want: 0},
		{path: "test.md", filename: "a/test.md", want: -1},
	}

	for _, tt := range cases {
		if got := pathMatchScore(tt.path, tt.filename); got != tt.want {
			t.Errorf("[ERROR] got: %v, want: %v with %v -> %v", got, tt.want, tt.path, tt.filename)
		}
	}
}

func TestFindPath(t *testing.T) {
	testFindPathRootDir := filepath.Join("..", "testdata", "findpath")
	cases := []struct {
		name   string
		root   string // テストで設定する vault のプロジェクトディレクトリ
		fileId string
		want   string
	}{
		{name: "in cur dir", root: "simple", fileId: "test", want: "test.md"},
		{name: "in subdir", root: "subdir", fileId: "test", want: "a/test.md"},
		{name: "in subdir specified", root: "specified_subdir", fileId: "a/test", want: "a/test.md"},
		{name: "in cur dir and subdir", root: "cur_subdir", fileId: "test", want: "test.md"},
		{name: "in cur dir and specified subdir", root: "cur_specified_subdir", fileId: "a/test", want: "a/test.md"},
		{name: "in multiple subdirs", root: "subdir_x2", fileId: "test", want: "a/test.md"},
		{name: "not found", root: "simple", fileId: "not_found", want: ""},
		{name: "with ext", root: "simple", fileId: "test.md", want: "test.md"},
		{name: "image", root: "image", fileId: "test.png", want: "test.png"},
		{name: "書記素クラスタ", root: "dakuten", fileId: "だくてん", want: "だくてん.md"},
	}

	for _, tt := range cases {
		got, err := findPath(filepath.Join(testFindPathRootDir, tt.root), tt.fileId)
		if err != nil {
			t.Errorf("[FAIL | %v] %v", tt.name, err)
			continue
		}
		if got != tt.want {
			t.Errorf("[ERROR | %v] got: %v, want: %v", tt.name, got, tt.want)
		}
	}
}

func TestBuildLinkText(t *testing.T) {
	cases := []struct {
		displayName string
		fileId      string
		fragments   []string
		want        string
	}{
		{displayName: "test", fileId: "211023", fragments: []string{"", ""}, want: "test"},
		{displayName: "", fileId: "211023", fragments: nil, want: "211023"},
		{displayName: "", fileId: "211023", fragments: []string{"section", "subsection"}, want: "211023 > section > subsection"},
		{displayName: "", fileId: "", fragments: []string{"section", "subsection"}, want: "section > subsection"},
	}

	for _, tt := range cases {
		if got := buildLinkText(tt.displayName, tt.fileId, tt.fragments); got != tt.want {
			t.Errorf("[ERROR] got: %q, want %q", got, tt.want)
		}
	}
}
