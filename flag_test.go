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
		wantErr   error
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
				publishable: true,
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
				publishable: true,
				cmmt:        true,
				strictref:   false,
				obs:         false,
				std:         true,
			},
		},
		{
			name: "source not specified",
			cmdflags: map[string]string{
				FLAG_DESTINATION: "dst",
			},
			wantErr: ErrFlagSourceNotSet,
		},
		{
			name: "destination not specified",
			cmdflags: map[string]string{
				FLAG_SOURCE: "src",
			},
			wantErr: ErrFlagDestinationNotSet,
		},
		{
			name: fmt.Sprintf("%s flag requires %s flag", FLAG_STRICT_REF, FLAG_CONVERT_LINKS),
			cmdflags: map[string]string{
				FLAG_SOURCE:      "src",
				FLAG_DESTINATION: "dst",
				FLAG_STRICT_REF:  "1",
			},
			wantErr: ErrFlagStrictRefNeedsLink,
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

		err := setFlags(flagset, gotflags)
		if tt.wantErr == nil {
			if err != nil {
				t.Fatalf("[Fatal | %s] unexpected err occurred: %v", tt.name, err)
			}
			if *gotflags != tt.wantflags {
				t.Errorf("[ERROR | %s]\n\t got: %+v,\n\twant: %+v", tt.name, *gotflags, tt.wantflags)
			}
		} else {
			if err == nil {
				t.Fatalf("[Fatal | %s] expected err did not occur: %v", tt.name, tt.wantErr)
			}
		}
	}
}
