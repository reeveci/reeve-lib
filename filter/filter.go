package filter

import (
	"bufio"
	"bytes"
	"io"
)

func LineFilter(r io.Reader, w io.Writer, filter func(line string) string) error {
	s := bufio.NewScanner(r)
	splitter := &lineSplitter{}
	s.Split(splitter.Split)

	for s.Scan() {
		filtered := []byte(filter(s.Text() + "\n"))
		if len(filtered) > 0 {
			_, err := w.Write(filtered)
			if err != nil {
				return err
			}
		}
	}
	if err := s.Err(); err != nil {
		return err
	}
	return nil
}

type lineSplitter struct {
	afterCR bool
}

// Mostly bufio.ScanLines code, however this function also splits at single carriage returns.
func (s *lineSplitter) Split(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}
	if s.afterCR {
		s.afterCR = false
		if data[0] == '\n' {
			// We had a carriage return before, so this newline needs to be skipped.
			return 1, nil, nil
		}
	}
	if i := bytes.IndexAny(data, "\r\n"); i >= 0 {
		if data[i] == '\n' {
			// We have a full line terminated by a single newline.
			return i + 1, data[0:i], nil
		}
		// We have a full line terminated by either a single carriage return or carriage return and newline.
		advance = i + 1
		if len(data) == i+1 {
			// We are at the end of the input and do not know yet if the next symbol corresponds to the current carriage return or not.
			s.afterCR = true
		} else if data[i+1] == '\n' {
			advance += 1
		}
		return advance, data[0:i], nil
	}
	// If we're at EOF, we have a final, non-terminated line. Return it.
	if atEOF {
		return len(data), data, nil
	}
	// Request more data.
	return 0, nil, nil
}
