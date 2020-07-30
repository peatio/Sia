package api

import (
	"net/url"
	"testing"

	"gitlab.com/NebulousLabs/errors"

	"gitlab.com/NebulousLabs/Sia/modules"
)

// TestDefaultPath ensures defaultPath functions correctly.
func TestDefaultPath(t *testing.T) {
	tests := []struct {
		name               string
		queryForm          url.Values
		subfiles           modules.SkyfileSubfiles
		defaultPath        string
		disableDefaultPath bool
		err                error
	}{
		{
			name:               "single file not multipart nil",
			queryForm:          url.Values{},
			subfiles:           nil,
			defaultPath:        "",
			disableDefaultPath: false,
			err:                nil,
		},
		{
			name:               "single file not multipart empty",
			queryForm:          url.Values{modules.SkyfileDisableDefaultPathParamName: []string{"true"}},
			subfiles:           nil,
			defaultPath:        "",
			disableDefaultPath: false,
			err:                ErrInvalidDefaultPath,
		},
		{
			name:               "single file not multipart set",
			queryForm:          url.Values{modules.SkyfileDefaultPathParamName: []string{"about.html"}},
			subfiles:           nil,
			defaultPath:        "",
			disableDefaultPath: false,
			err:                ErrInvalidDefaultPath,
		},

		{
			name:               "single file multipart nil",
			queryForm:          url.Values{},
			subfiles:           modules.SkyfileSubfiles{"about.html": modules.SkyfileSubfileMetadata{}},
			defaultPath:        "/about.html",
			disableDefaultPath: false,
			err:                nil,
		},
		{
			name:               "single file multipart empty",
			queryForm:          url.Values{modules.SkyfileDisableDefaultPathParamName: []string{"true"}},
			subfiles:           modules.SkyfileSubfiles{"about.html": modules.SkyfileSubfileMetadata{}},
			defaultPath:        "",
			disableDefaultPath: true,
			err:                nil,
		},
		{
			name:               "single file multipart set to only",
			queryForm:          url.Values{modules.SkyfileDefaultPathParamName: []string{"about.html"}},
			subfiles:           modules.SkyfileSubfiles{"about.html": modules.SkyfileSubfileMetadata{}},
			defaultPath:        "/about.html",
			disableDefaultPath: false,
			err:                nil,
		},
		{
			name:               "single file multipart set to nonexistent",
			queryForm:          url.Values{modules.SkyfileDefaultPathParamName: []string{"nonexistent.html"}},
			subfiles:           modules.SkyfileSubfiles{"about.html": modules.SkyfileSubfileMetadata{}},
			defaultPath:        "",
			disableDefaultPath: false,
			err:                ErrInvalidDefaultPath,
		},
		{
			name:               "single file multipart set to non-html",
			queryForm:          url.Values{modules.SkyfileDefaultPathParamName: []string{"about.js"}},
			subfiles:           modules.SkyfileSubfiles{"about.js": modules.SkyfileSubfileMetadata{}},
			defaultPath:        "",
			disableDefaultPath: false,
			err:                ErrInvalidDefaultPath,
		},
		{
			name: "single file multipart both set",
			queryForm: url.Values{
				modules.SkyfileDefaultPathParamName:        []string{"about.html"},
				modules.SkyfileDisableDefaultPathParamName: []string{"true"},
			},
			subfiles:           modules.SkyfileSubfiles{"about.html": modules.SkyfileSubfileMetadata{}},
			defaultPath:        "",
			disableDefaultPath: false,
			err:                ErrInvalidDefaultPath,
		},
		{
			name:               "single file multipart set to non-root",
			queryForm:          url.Values{modules.SkyfileDefaultPathParamName: []string{"foo/bar/about.html"}},
			subfiles:           modules.SkyfileSubfiles{"foo/bar/about.html": modules.SkyfileSubfileMetadata{}},
			defaultPath:        "",
			disableDefaultPath: false,
			err:                ErrInvalidDefaultPath,
		},

		{
			name:      "multi file nil has index.html",
			queryForm: url.Values{},
			subfiles: modules.SkyfileSubfiles{
				"about.html": modules.SkyfileSubfileMetadata{},
				"index.html": modules.SkyfileSubfileMetadata{},
			},
			defaultPath:        "/index.html",
			disableDefaultPath: false,
			err:                nil,
		},
		{
			name:      "multi file nil no index.html",
			queryForm: url.Values{},
			subfiles: modules.SkyfileSubfiles{
				"about.html": modules.SkyfileSubfileMetadata{},
				"hello.html": modules.SkyfileSubfileMetadata{},
			},
			defaultPath:        "",
			disableDefaultPath: false,
			err:                nil,
		},
		{
			name:      "multi file set to empty",
			queryForm: url.Values{modules.SkyfileDisableDefaultPathParamName: []string{"true"}},
			subfiles: modules.SkyfileSubfiles{
				"about.html": modules.SkyfileSubfileMetadata{},
				"index.html": modules.SkyfileSubfileMetadata{},
			},
			defaultPath:        "",
			disableDefaultPath: true,
			err:                nil,
		},
		{
			name:      "multi file set to existing",
			queryForm: url.Values{modules.SkyfileDefaultPathParamName: []string{"about.html"}},
			subfiles: modules.SkyfileSubfiles{
				"about.html": modules.SkyfileSubfileMetadata{},
				"index.html": modules.SkyfileSubfileMetadata{},
			},
			defaultPath:        "/about.html",
			disableDefaultPath: false,
			err:                nil,
		},
		{
			name:      "multi file set to nonexistent",
			queryForm: url.Values{modules.SkyfileDefaultPathParamName: []string{"nonexistent.html"}},
			subfiles: modules.SkyfileSubfiles{
				"about.html": modules.SkyfileSubfileMetadata{},
				"index.html": modules.SkyfileSubfileMetadata{},
			},
			defaultPath:        "",
			disableDefaultPath: false,
			err:                ErrInvalidDefaultPath,
		},
		{
			name:      "multi file set to non-html",
			queryForm: url.Values{modules.SkyfileDefaultPathParamName: []string{"about.js"}},
			subfiles: modules.SkyfileSubfiles{
				"about.js":   modules.SkyfileSubfileMetadata{},
				"index.html": modules.SkyfileSubfileMetadata{},
			},
			defaultPath:        "",
			disableDefaultPath: false,
			err:                ErrInvalidDefaultPath,
		},
		{
			name: "multi file both set",
			queryForm: url.Values{
				modules.SkyfileDefaultPathParamName:        []string{"about.html"},
				modules.SkyfileDisableDefaultPathParamName: []string{"true"},
			},
			subfiles: modules.SkyfileSubfiles{
				"about.html": modules.SkyfileSubfileMetadata{},
				"index.html": modules.SkyfileSubfileMetadata{},
			},
			defaultPath:        "",
			disableDefaultPath: false,
			err:                ErrInvalidDefaultPath,
		},
		{
			name:      "multi file set to non-root",
			queryForm: url.Values{modules.SkyfileDefaultPathParamName: []string{"foo/bar/about.html"}},
			subfiles: modules.SkyfileSubfiles{
				"foo/bar/about.html": modules.SkyfileSubfileMetadata{},
				"foo/bar/baz.html":   modules.SkyfileSubfileMetadata{},
			},
			defaultPath:        "",
			disableDefaultPath: false,
			err:                ErrInvalidDefaultPath,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dp, ddp, err := defaultPath(tt.queryForm, tt.subfiles)
			if (err != nil || tt.err != nil) && !errors.Contains(err, tt.err) {
				t.Fatalf("Expected error %v, got %v\n", tt.err, err)
			}
			if dp != tt.defaultPath {
				t.Fatalf("Expected defaultPath '%v', got '%v'\n", tt.defaultPath, dp)
			}
			if ddp != tt.disableDefaultPath {
				t.Fatalf("Expected disableDefaultPath '%v', got '%v'\n", tt.disableDefaultPath, ddp)
			}
		})
	}
}
