package main

import (
	"flag"
	"fmt"
	"testing"
)

func TestSetFlags(t *testing.T) {
	cases := []struct {
		name      string
		cmdflags  map[string]string
		wantflags flagBundle
	}{
		{
			name: "for obsidian usage",
			cmdflags: map[string]string{
				FLAG_SOURCE:         "src",
				FLAG_DESTINATION:    "dst",
				FLAG_OBSIDIAN_USAGE: "1",
			},
			wantflags: flagBundle{
				src:   "src",
				dst:   "dst",
				cptag: true,
				title: true,
				alias: true,
				obs:   true,
			},
		},
		{
			name: "for standard usage",
			cmdflags: map[string]string{
				FLAG_SOURCE:         "src",
				FLAG_DESTINATION:    "dst",
				FLAG_STANDARD_USAGE: "1",
			},
			wantflags: flagBundle{
				src:         "src",
				dst:         "dst",
				rmtag:       true,
				cptag:       true,
				title:       true,
				alias:       true,
				link:        true,
				strictref:   true,
				cmmt:        true,
				std:         true,
			},
		},
		{
			name: "standard usage overwritten",
			cmdflags: map[string]string{
				FLAG_SOURCE:         "src",
				FLAG_DESTINATION:    "dst",
				FLAG_REMOVE_TAGS:    "0",
				FLAG_STRICT_REF:     "0",
				FLAG_STANDARD_USAGE: "1",
			},
			wantflags: flagBundle{
				src:         "src",
				dst:         "dst",
				rmtag:       false,
				cptag:       true,
				title:       true,
				alias:       true,
				link:        true,
				cmmt:        true,
				strictref:   false,
				obs:         false,
				std:         true,
			},
		},
	}

	for _, tt := range cases {
		flagset := flag.NewFlagSet(fmt.Sprintf("TestSetFlags | %s", tt.name), flag.ExitOnError)
		gotflags := new(flagBundle)
		initFlags(flagset, gotflags)

		// テスト用コマンドライン引数の設定
		for cmdname, cmdvalue := range tt.cmdflags {
			flagset.Set(cmdname, cmdvalue)
		}

		setFlags(flagset, gotflags)
		if *gotflags != tt.wantflags {
			t.Errorf("[ERROR | %s]\n\t got: %+v,\n\twant: %+v", tt.name, *gotflags, tt.wantflags)
		}
	}
}

func TestVerifyFlags(t *testing.T) {
	cases := []struct {
		name    string
		flags   flagBundle
		wantErr mainErr
	}{
		{
			name: "src not set",
			flags: flagBundle{
				src: "",
				dst: "dst",
			},
			wantErr: newMainErr(MAIN_ERR_KIND_SOURCE_NOT_SET),
		},
		{
			name: "dst not set",
			flags: flagBundle{
				src: "src",
				dst: "",
			},
			wantErr: newMainErr(MAIN_ERR_KIND_DESTINATION_NOT_SET),
		},
		{
			name: fmt.Sprintf("%s set but not %s", FLAG_STRICT_REF, FLAG_CONVERT_LINKS),
			flags: flagBundle{
				src:       "src",
				dst:       "dst",
				link:      false,
				strictref: true,
			},
			wantErr: newMainErr(MAIN_ERR_KIND_STRICTREF_NEEDS_LINK),
		},
		{
			name: "src begins with \"-\"",
			flags: flagBundle{
				src: "-src",
				dst: "dst",
			},
			wantErr: newMainErr(MAIN_ERR_KIND_INVALID_SOURCE_FORMAT),
		},
		{
			name: "dst begins with \"-\"",
			flags: flagBundle{
				src: "src",
				dst: "-dst",
			},
			wantErr: newMainErr(MAIN_ERR_KIND_INVALID_DESTINATION_FORMAT),
		},
	}

	for _, tt := range cases {
		err := verifyFlags(&tt.flags)
		if err == nil && tt.wantErr != nil {
			t.Errorf("[ERROR | %s] expected error did not occurr with %+v", tt.name, tt.flags)
		}
		if err != nil && tt.wantErr == nil {
			t.Fatalf("[FATAL | %s] unexpected error occurred: %v", tt.name, err)
		}
		if err != nil && tt.wantErr != nil {
			e, ok := err.(mainErr)
			if !(ok && e.Kind() == tt.wantErr.Kind()) {
				t.Fatalf("[FATAL | %s] unexpected error occurred: %v", tt.name, err)
			}
		}
	}
}
