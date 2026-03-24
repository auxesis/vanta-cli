package main

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func ptr[T any](v T) *T { return &v }

func TestFilterTestsDueBefore(t *testing.T) {
	cutoff := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)

	cases := []struct {
		name        string
		input       Test
		wantInclude bool
	}{
		{
			name: "DUE_SOON with date after cutoff is excluded",
			input: Test{
				ID:     "a",
				Status: "NEEDS_ATTENTION",
				RemediationStatusInfo: &TestRemediationStatusInfo{
					Status:                 "DUE_SOON",
					SoonestRemediateByDate: ptr(time.Date(2026, 3, 15, 0, 0, 0, 0, time.UTC)),
				},
			},
			wantInclude: false,
		},
		{
			name: "DUE_SOON with date before cutoff is included",
			input: Test{
				ID:     "b",
				Status: "NEEDS_ATTENTION",
				RemediationStatusInfo: &TestRemediationStatusInfo{
					Status:                 "DUE_SOON",
					SoonestRemediateByDate: ptr(time.Date(2026, 2, 15, 0, 0, 0, 0, time.UTC)),
				},
			},
			wantInclude: true,
		},
		{
			name: "NEEDS_WORK with nil date is excluded",
			input: Test{
				ID:     "c",
				Status: "NEEDS_ATTENTION",
				RemediationStatusInfo: &TestRemediationStatusInfo{
					Status:                 "NEEDS_WORK",
					SoonestRemediateByDate: nil,
				},
			},
			wantInclude: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result := filterTestsDueBefore([]Test{tc.input}, cutoff)
			if tc.wantInclude {
				assert.Len(t, result, 1)
			} else {
				assert.Len(t, result, 0)
			}
		})
	}
}

func TestFilterTestsDueAfter(t *testing.T) {
	cutoff := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)

	cases := []struct {
		name        string
		input       Test
		wantInclude bool
	}{
		{
			name: "DUE_SOON with date after cutoff is included",
			input: Test{
				ID:     "a",
				Status: "NEEDS_ATTENTION",
				RemediationStatusInfo: &TestRemediationStatusInfo{
					Status:                 "DUE_SOON",
					SoonestRemediateByDate: ptr(time.Date(2026, 3, 15, 0, 0, 0, 0, time.UTC)),
				},
			},
			wantInclude: true,
		},
		{
			name: "DUE_SOON with date before cutoff is excluded",
			input: Test{
				ID:     "b",
				Status: "NEEDS_ATTENTION",
				RemediationStatusInfo: &TestRemediationStatusInfo{
					Status:                 "DUE_SOON",
					SoonestRemediateByDate: ptr(time.Date(2026, 2, 15, 0, 0, 0, 0, time.UTC)),
				},
			},
			wantInclude: false,
		},
		{
			name: "NEEDS_WORK with nil date is excluded",
			input: Test{
				ID:     "c",
				Status: "NEEDS_ATTENTION",
				RemediationStatusInfo: &TestRemediationStatusInfo{
					Status:                 "NEEDS_WORK",
					SoonestRemediateByDate: nil,
				},
			},
			wantInclude: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result := filterTestsDueAfter([]Test{tc.input}, cutoff)
			if tc.wantInclude {
				assert.Len(t, result, 1)
			} else {
				assert.Len(t, result, 0)
			}
		})
	}
}
