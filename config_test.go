package main

import (
	"flag"
	"fmt"
	"testing"
)

func TestSetConfig(t *testing.T) {
	cases := []struct {
		name       string
		cmdflags   map[string]string
		wantConfig configuration
	}{
		{
			name: "for obsidian usage",
			cmdflags: map[string]string{
				FLAG_SOURCE:         "src",
				FLAG_DESTINATION:    "dst",
				FLAG_OBSIDIAN_USAGE: "1",
			},
			wantConfig: configuration{
				src:   "src",
				dst:   "dst",
				cptag: true,
				title: true,
				alias: true,
				obs:   true,
				tgt:   "src",
			},
		},
		{
			name: "for standard usage",
			cmdflags: map[string]string{
				FLAG_SOURCE:         "src",
				FLAG_DESTINATION:    "dst",
				FLAG_STANDARD_USAGE: "1",
			},
			wantConfig: configuration{
				src:       "src",
				dst:       "dst",
				rmtag:     true,
				cptag:     true,
				title:     true,
				alias:     true,
				link:      true,
				strictref: true,
				cmmt:      true,
				std:       true,
				tgt:       "src",
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
			wantConfig: configuration{
				src:       "src",
				dst:       "dst",
				rmtag:     false,
				cptag:     true,
				title:     true,
				alias:     true,
				link:      true,
				cmmt:      true,
				strictref: false,
				obs:       false,
				std:       true,
				tgt:       "src",
			},
		},
		{
			name: "tgt overwrittedn",
			cmdflags: map[string]string{
				FLAG_SOURCE:      "src",
				FLAG_DESTINATION: "dst",
				FLAG_TARGET:      "tgt",
			},
			wantConfig: configuration{
				src: "src",
				dst: "dst",
				tgt: "tgt",
			},
		},
	}

	for _, tt := range cases {
		flagset := flag.NewFlagSet(fmt.Sprintf("TestSetFlags | %s", tt.name), flag.ExitOnError)
		gotConfig := new(configuration)
		initFlags(flagset, gotConfig)

		// テスト用コマンドライン引数の設定
		for cmdname, cmdvalue := range tt.cmdflags {
			flagset.Set(cmdname, cmdvalue)
		}

		setConfig(flagset, gotConfig)
		if *gotConfig != tt.wantConfig {
			t.Errorf("[ERROR | %s]\n\t got: %+v,\n\twant: %+v", tt.name, *gotConfig, tt.wantConfig)
		}
	}
}

func TestVerifyConfig(t *testing.T) {
	cases := []struct {
		name    string
		config  configuration
		wantErr mainErr
	}{
		{
			name: "src not set",
			config: configuration{
				src: "",
				dst: "dst",
			},
			wantErr: newMainErr(MAIN_ERR_KIND_SOURCE_NOT_SET),
		},
		{
			name: "dst not set",
			config: configuration{
				src: "src",
				dst: "",
			},
			wantErr: newMainErr(MAIN_ERR_KIND_DESTINATION_NOT_SET),
		},
		{
			name: fmt.Sprintf("%s set but not %s", FLAG_STRICT_REF, FLAG_CONVERT_LINKS),
			config: configuration{
				src:       "src",
				dst:       "dst",
				link:      false,
				strictref: true,
			},
			wantErr: newMainErr(MAIN_ERR_KIND_STRICTREF_NEEDS_LINK),
		},
		{
			name: "src begins with \"-\"",
			config: configuration{
				src: "-src",
				dst: "dst",
			},
			wantErr: newMainErr(MAIN_ERR_KIND_INVALID_SOURCE_FORMAT),
		},
		{
			name: "dst begins with \"-\"",
			config: configuration{
				src: "src",
				dst: "-dst",
			},
			wantErr: newMainErr(MAIN_ERR_KIND_INVALID_DESTINATION_FORMAT),
		},
	}

	for _, tt := range cases {
		err := verifyConfig(&tt.config)
		if err == nil && tt.wantErr != nil {
			t.Errorf("[ERROR | %s] expected error did not occurr with %+v", tt.name, tt.config)
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
