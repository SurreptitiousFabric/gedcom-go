package parser

import (
	"errors"
	"strings"
	"testing"
	"testing/iotest"
)

func TestSetMaxNestingDepthReset(t *testing.T) {
	p := NewParser()
	p.SetMaxNestingDepth(10)
	if p.maxDepth != 10 {
		t.Fatalf("maxDepth = %d, want 10", p.maxDepth)
	}

	p.SetMaxNestingDepth(0)
	if p.maxDepth != MaxNestingDepth {
		t.Fatalf("maxDepth = %d, want %d", p.maxDepth, MaxNestingDepth)
	}
}

func TestParseUsesDefaultDepthWhenMaxDepthZero(t *testing.T) {
	p := NewParser()
	p.maxDepth = 0

	_, err := p.ParseLine("101 DEEP")
	if err == nil {
		t.Fatal("ParseLine() expected nesting depth error")
	}
	var levelErr *InvalidLevelError
	if !errors.As(err, &levelErr) {
		t.Fatalf("expected InvalidLevelError, got %T", err)
	}
}

func TestFieldStartIndex(t *testing.T) {
	if got := fieldStartIndex("0 HEAD", -1); got != -1 {
		t.Fatalf("fieldStartIndex() = %d, want -1 for negative index", got)
	}

	if got := fieldStartIndex("0 HEAD", 2); got != -1 {
		t.Fatalf("fieldStartIndex() = %d, want -1 for missing field", got)
	}

	line := "0 @I1@ INDI value"
	pos := fieldStartIndex(line, 3)
	if pos == -1 {
		t.Fatal("fieldStartIndex() = -1, want value position")
	}
	if got := line[pos:]; got != "value" {
		t.Fatalf("fieldStartIndex() value = %q, want %q", got, "value")
	}
}

func TestValidateTagEmpty(t *testing.T) {
	err := validateTag("")
	if err == nil {
		t.Fatal("validateTag() expected error for empty tag")
	}
	var tagErr *InvalidTagError
	if !errors.As(err, &tagErr) {
		t.Fatalf("expected InvalidTagError, got %T", err)
	}
}

func TestValidateXRef(t *testing.T) {
	tests := []string{
		"@@",
		"@I1@@",
	}

	for _, xref := range tests {
		err := validateXRef(xref)
		if err == nil {
			t.Fatalf("validateXRef(%q) expected error", xref)
		}
		var xrefErr *InvalidXRefError
		if !errors.As(err, &xrefErr) {
			t.Fatalf("expected InvalidXRefError, got %T", err)
		}
	}
}

func TestParseWithRecovery(t *testing.T) {
	input := "0 HEAD\nINVALID\n0 TRLR\n"

	p := NewParser()
	lines, errs := p.ParseWithRecovery(strings.NewReader(input))

	if len(lines) != 2 {
		t.Fatalf("ParseWithRecovery() lines = %d, want 2", len(lines))
	}
	if lines[0].Tag != "HEAD" || lines[1].Tag != "TRLR" {
		t.Fatalf("ParseWithRecovery() tags = %s/%s, want HEAD/TRLR", lines[0].Tag, lines[1].Tag)
	}
	if len(errs) != 1 {
		t.Fatalf("ParseWithRecovery() errs = %d, want 1", len(errs))
	}
	parseErr, ok := errs[0].(*ParseError)
	if !ok {
		t.Fatalf("expected *ParseError, got %T", errs[0])
	}
	if parseErr.Context != "prev: 0 HEAD | line: INVALID" {
		t.Fatalf("ParseWithRecovery() context = %q, want %q", parseErr.Context, "prev: 0 HEAD | line: INVALID")
	}
}

func TestParseWithRecoveryReaderError(t *testing.T) {
	testErr := errors.New("read error")
	reader := iotest.ErrReader(testErr)

	p := NewParser()
	lines, errs := p.ParseWithRecovery(reader)

	if len(lines) != 0 {
		t.Fatalf("ParseWithRecovery() lines = %d, want 0", len(lines))
	}
	if len(errs) != 1 {
		t.Fatalf("ParseWithRecovery() errs = %d, want 1", len(errs))
	}
	if !errors.Is(errs[0], testErr) {
		t.Fatalf("ParseWithRecovery() error = %v, want %v", errs[0], testErr)
	}
}

func TestErrorTypeMessages(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want string
	}{
		{
			name: "invalid level",
			err:  &InvalidLevelError{Raw: "X", Reason: "not a number"},
			want: `invalid level "X": not a number`,
		},
		{
			name: "level mismatch",
			err:  &LevelMismatchError{Previous: 1, Current: 3},
			want: "level jump from 1 to 3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.want {
				t.Fatalf("Error() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestParseErrorNoContext(t *testing.T) {
	err := &ParseError{Line: 2, Message: "bad line"}
	if got := err.Error(); got != "line 2: bad line" {
		t.Fatalf("ParseError.Error() = %q, want %q", got, "line 2: bad line")
	}
}

func TestEnrichParseErrorNonParse(t *testing.T) {
	baseErr := errors.New("boom")
	if got := enrichParseError(baseErr, "prev", "line"); got != baseErr {
		t.Fatalf("enrichParseError() = %v, want %v", got, baseErr)
	}
}
