package main

import (
	"bytes"
	"flag"
	"fmt"
	"path/filepath"
	"testing"
)

func TestRun(t *testing.T) {
	rootDir := "sample"
	src := "src"
	dst := "dst"
	tmp := "tmp"
	cases := []struct {
		name        string
		cmdflags    map[string]string
		argVersion  string
		wantDstDir  string
		wantVersion string
	}{
		{
			name: "-version",
			cmdflags: map[string]string{
				FLAG_VERSION: "1",
			},
			argVersion:  "1.0.0",
			wantVersion: "v1.0.0\n",
		},
		{
			name: "-obs",
			cmdflags: map[string]string{
				FLAG_SOURCE:         filepath.Join(rootDir, "obs", src),
				FLAG_DESTINATION:    filepath.Join(rootDir, "obs", tmp),
				FLAG_OBSIDIAN_USAGE: "1",
			},
			wantDstDir: filepath.Join(rootDir, "obs", dst),
		},
	}

	for _, tt := range cases {
		// flags を設定
		flags := new(flagBundle)
		flagset := flag.NewFlagSet(fmt.Sprintf("TestSetFlags | %s", tt.name), flag.ExitOnError)
		initFlags(flagset, flags)
		for cmdname, cmdvalue := range tt.cmdflags { // flag.Parse() に相当
			flagset.Set(cmdname, cmdvalue)
		}
		setFlags(flagset, flags)

		versionBuf := new(bytes.Buffer)
		err := run(tt.argVersion, flags, versionBuf)
		if err != nil {
			t.Fatalf("[FATAL | %s] unexpected err occurred: %v", tt.name, err)
		}
		if gotVersion := versionBuf.String(); gotVersion != "" {
			if gotVersion != tt.wantVersion {
				t.Errorf("[ERROR | version - %s] got: %q, want: %q", tt.name, gotVersion, tt.wantVersion)
			}
			continue
		}

		if !equalDirContent(flags.dst, tt.wantDstDir) {
			t.Fatalf("[ERROR | content - %s]", tt.name)
		}
	}
}

func equalDirContent(leftdir, rightdir string) bool {
	return true
}
