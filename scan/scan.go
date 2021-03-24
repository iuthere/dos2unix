package scan

import "bytes"

// ScanLinesKeep scans line by line, retaining the endings:
//
//	\r\n
//	\n
//
// The singular \r does not split the lines.
//
// The \n\r results in a line ending with \n and new line that begins with \r.
func ScanLinesKeep(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}
	if i := bytes.Index(data, []byte{'\r', '\n'}); i >= 0 {
		// We have a full newline-terminated line.
		return i + 2, data[0 : i+2], nil
	}

	// \n\r will yield next line beginning with \r

	//if i := bytes.Index(data, []byte{'\n', '\r'}); i >= 0 {
	//	// We have a full newline-terminated line.
	//	return i + 2, data[0 : i+2], nil
	//}

	// Single \r as a line ending is pretty archaic (Mac Classic before 2001)

	//if i := bytes.IndexByte(data, '\r'); i >= 0 {
	//	// We have a full newline-terminated line.
	//	return i + 1, data[0 : i+1], nil
	//}

	if i := bytes.IndexByte(data, '\n'); i >= 0 {
		// We have a full newline-terminated line.
		return i + 1, data[0 : i+1], nil
	}
	// If we're at EOF, we have a final, non-terminated line. Return it.
	if atEOF {
		return len(data), data, nil
	}
	// Request more data.
	return 0, nil, nil
}
