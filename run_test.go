package main

import (
	"bytes"
	"testing"
)

func TestRun(t *testing.T) {
	// rootDir := "sample"
	// gotDstDir := filepath.Join(rootDir, "tmp")
	cases := []struct {
		name        string
		flags       flagBundle
		argVersion  string
		wantDstDir  string
		wantVersion string
	}{
		{
			name: "-version",
			flags: flagBundle{
				ver: true,
			},
			argVersion:  "1.0.0",
			wantVersion: "v1.0.0\n",
		},
	}

	for _, tt := range cases {
		versionBuf := new(bytes.Buffer)
		err := run(tt.argVersion, &tt.flags, versionBuf)
		if err != nil {
			t.Fatalf("[FATAL | %s] unexpected err occurred: %v", tt.name, err)
		}
		if gotVersion := versionBuf.String(); gotVersion != "" {
			if gotVersion != tt.wantVersion {
				t.Errorf("[ERROR | version - %s] got: %q, want: %q", tt.name, gotVersion, tt.wantVersion)
			}
			continue
		}

		if !equalDirContent(tt.flags.dst, tt.wantDstDir) {
			t.Fatalf("[ERROR | content - %s]", tt.name)
		}
	}
}

func equalDirContent(leftdir, rightdir string) bool {
	return false
}
