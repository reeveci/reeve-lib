package streams

import (
	"fmt"

	"github.com/djherbis/stream"
	"github.com/reeveci/reeve-lib/schema"
)

func NewStreamProvider(stream *stream.Stream) *StreamProvider {
	return &StreamProvider{Stream: stream}
}

type StreamProvider struct {
	*stream.Stream
}

func (s *StreamProvider) Available() bool {
	return s != nil && s.Stream != nil
}

func (s *StreamProvider) Reader() (schema.LogReader, error) {
	if !s.Available() {
		return nil, fmt.Errorf("no logs available")
	}

	return s.NextReader()
}

func (s *StreamProvider) Close() error {
	if !s.Available() {
		return nil
	}

	return s.Stream.Close()
}
