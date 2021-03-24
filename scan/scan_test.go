package scan

import (
	"bufio"
	"io"
	"strings"
	"testing"
)

// Test that the line splitter handles a final line without a newline.
func testWithNewline(text string, lines []string, t *testing.T) {
	buf := strings.NewReader(text)
	s := bufio.NewScanner(&slowReader{7, buf})
	s.Split(ScanLinesKeep)
	var b strings.Builder
	var lineNum int
	for lineNum = 0; s.Scan(); lineNum++ {
		//if lineNum < len(lines) { // we want to panic
		line := lines[lineNum]
		if s.Text() != line {
			t.Errorf("%d: bad line: %d %d\n%.100q\n%.100q\n", lineNum, len(s.Bytes()), len(line), s.Bytes(), line)
		}
		//}
		b.WriteString(s.Text())
	}
	if lineNum != len(lines) {
		t.Errorf("wrong number of lines: got: %d, want: %d", lineNum, len(lines))
	}
	if b.String() != text {
		t.Errorf("wrong reproduced result: got: %v, want: %v", b.String(), text)
	}
	err := s.Err()
	if err != nil {
		t.Fatal(err)
	}
}

// slowReader is a reader that returns only a few bytes at a time, to test the incremental
// reads in Scanner.Scan.
type slowReader struct {
	max int
	buf io.Reader
}

func (sr *slowReader) Read(p []byte) (n int, err error) {
	if len(p) > sr.max {
		p = p[0:sr.max]
	}
	return sr.buf.Read(p)
}

// Test that the line splitter handles a final line without a newline.
func TestScanLineNoNewline(t *testing.T) {
	const text = "abcdefghijklmn\nopqrstuvwxyz"
	lines := []string{
		"abcdefghijklmn\n",
		"opqrstuvwxyz",
	}
	testWithNewline(text, lines, t)
}

// Test that the line splitter handles a final line with a newline.
func TestScanLineWithNewline(t *testing.T) {
	const text = "abcdefghijklmn\nopqrstuvwxyz\n"
	lines := []string{
		"abcdefghijklmn\n",
		"opqrstuvwxyz\n",
	}
	testWithNewline(text, lines, t)
}

// Test that the line splitter handles a final line with a carriage return but no newline.
func TestScanLineReturnButNoNewline(t *testing.T) {
	const text = "abcdefghijklmn\nopqrstuvwxyz\r"
	lines := []string{
		"abcdefghijklmn\n",
		"opqrstuvwxyz\r",
	}
	testWithNewline(text, lines, t)
}

// Test that the line splitter handles a final empty line.
func TestScanLineEmptyFinalLines(t *testing.T) {
	const text = "abcdefghijklmn\nopqrstuvwxyz\n\n"
	lines := []string{
		"abcdefghijklmn\n",
		"opqrstuvwxyz\n",
		"\n",
	}
	testWithNewline(text, lines, t)
}

// Test that the line splitter handles a final empty line with a carriage return but no newline.
func TestScanLineEmptyFinalLineWithLFCR(t *testing.T) {
	const text = "abcdefghijklmn\nopqrstuvwxyz\n\rcontinue\n"
	lines := []string{
		"abcdefghijklmn\n",
		"opqrstuvwxyz\n",
		"\rcontinue\n",
	}
	testWithNewline(text, lines, t)
}

// Test that the line splitter handles a final carriage return and newline.
func TestScanLineEmptyFinalLineWithCRLF(t *testing.T) {
	const text = "abcdefghijklmn\nopqrstuvwxyz\r\n"
	lines := []string{
		"abcdefghijklmn\n",
		"opqrstuvwxyz\r\n",
	}
	testWithNewline(text, lines, t)
}

// Test that the line splitter handles a final carriage return and newline.
func TestScanLineCraziness(t *testing.T) {
	const text = "abcdefghijklmn\n\nopqrstuvwxyz\n\n\nABC"
	lines := []string{
		"abcdefghijklmn\n",
		"\n",
		"opqrstuvwxyz\n",
		"\n",
		"\n",
		"ABC",
	}
	testWithNewline(text, lines, t)
}

// Test empty input
func TestScanEmpty(t *testing.T) {
	const text = ""
	var lines []string
	testWithNewline(text, lines, t)
}

// Test empty input
func TestScanOnlyEmptyLines(t *testing.T) {
	const text = "\n"
	var lines = []string{
		"\n",
	}
	testWithNewline(text, lines, t)
}
