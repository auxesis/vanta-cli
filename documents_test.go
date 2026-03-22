package main

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestFilterDocumentsDueBefore(t *testing.T) {
	cutoff := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)

	cases := []struct {
		name        string
		input       Document
		wantInclude bool
	}{
		{
			name: "date after cutoff is excluded",
			input: Document{
				ID:               "a",
				UploadStatus:     "Needs document",
				UploadStatusDate: ptr(time.Date(2026, 3, 15, 0, 0, 0, 0, time.UTC)),
			},
			wantInclude: false,
		},
		{
			name: "date before cutoff is included",
			input: Document{
				ID:               "b",
				UploadStatus:     "Needs document",
				UploadStatusDate: ptr(time.Date(2026, 2, 15, 0, 0, 0, 0, time.UTC)),
			},
			wantInclude: true,
		},
		{
			name: "nil date is excluded",
			input: Document{
				ID:               "c",
				UploadStatus:     "Needs document",
				UploadStatusDate: nil,
			},
			wantInclude: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result := filterDocumentsDueBefore([]Document{tc.input}, cutoff)
			if tc.wantInclude {
				assert.Len(t, result, 1)
			} else {
				assert.Len(t, result, 0)
			}
		})
	}
}

func TestFilterDocumentsDueAfter(t *testing.T) {
	cutoff := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)

	cases := []struct {
		name        string
		input       Document
		wantInclude bool
	}{
		{
			name: "date after cutoff is included",
			input: Document{
				ID:               "a",
				UploadStatus:     "Needs document",
				UploadStatusDate: ptr(time.Date(2026, 3, 15, 0, 0, 0, 0, time.UTC)),
			},
			wantInclude: true,
		},
		{
			name: "date before cutoff is excluded",
			input: Document{
				ID:               "b",
				UploadStatus:     "Needs document",
				UploadStatusDate: ptr(time.Date(2026, 2, 15, 0, 0, 0, 0, time.UTC)),
			},
			wantInclude: false,
		},
		{
			name: "nil date is excluded",
			input: Document{
				ID:               "c",
				UploadStatus:     "Needs document",
				UploadStatusDate: nil,
			},
			wantInclude: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result := filterDocumentsDueAfter([]Document{tc.input}, cutoff)
			if tc.wantInclude {
				assert.Len(t, result, 1)
			} else {
				assert.Len(t, result, 0)
			}
		})
	}
}
