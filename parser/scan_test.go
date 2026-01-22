package parser

import "testing"

func TestScanGEDCOMLines(t *testing.T) {
	tests := []struct {
		name        string
		data        string
		atEOF       bool
		wantAdvance int
		wantToken   string
		wantNil     bool
	}{
		{
			name:        "empty at EOF",
			data:        "",
			atEOF:       true,
			wantAdvance: 0,
			wantNil:     true,
		},
		{
			name:        "lf line ending",
			data:        "0 HEAD\n1 GEDC",
			atEOF:       false,
			wantAdvance: len("0 HEAD\n"),
			wantToken:   "0 HEAD",
		},
		{
			name:        "crlf line ending",
			data:        "0 HEAD\r\n",
			atEOF:       false,
			wantAdvance: len("0 HEAD\r\n"),
			wantToken:   "0 HEAD",
		},
		{
			name:        "cr line ending",
			data:        "0 HEAD\r1 GEDC",
			atEOF:       false,
			wantAdvance: len("0 HEAD\r"),
			wantToken:   "0 HEAD",
		},
		{
			name:        "cr at end needs more data",
			data:        "0 HEAD\r",
			atEOF:       false,
			wantAdvance: 0,
			wantNil:     true,
		},
		{
			name:        "cr at end at EOF",
			data:        "0 HEAD\r",
			atEOF:       true,
			wantAdvance: len("0 HEAD\r"),
			wantToken:   "0 HEAD",
		},
		{
			name:        "no terminator at EOF",
			data:        "0 HEAD",
			atEOF:       true,
			wantAdvance: len("0 HEAD"),
			wantToken:   "0 HEAD",
		},
		{
			name:        "no terminator not at EOF",
			data:        "0 HEAD",
			atEOF:       false,
			wantAdvance: 0,
			wantNil:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			advance, token, err := ScanGEDCOMLines([]byte(tt.data), tt.atEOF)
			if err != nil {
				t.Fatalf("ScanGEDCOMLines() error = %v", err)
			}
			if advance != tt.wantAdvance {
				t.Fatalf("ScanGEDCOMLines() advance = %d, want %d", advance, tt.wantAdvance)
			}
			if tt.wantNil {
				if token != nil {
					t.Fatalf("ScanGEDCOMLines() token = %q, want nil", string(token))
				}
				return
			}
			if got := string(token); got != tt.wantToken {
				t.Fatalf("ScanGEDCOMLines() token = %q, want %q", got, tt.wantToken)
			}
		})
	}
}
