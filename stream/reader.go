package stream

import (
	"compress/gzip"
	"fmt"
	"io"

	"github.com/klauspost/compress/zstd"
)

type Reader struct {
	// The streamer to pull the data from
	streamer IStreamer
	// The inner buffer to store the already read data
	buffer []byte
	// The current position in the buffer
	isEOF bool
}

func NewReaderWithDecompression(url string, config *StreamerConfig, compression Compression) (io.Reader, error) {
	reader, err := NewReader(url, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create streamer: %w", err)
	}

	if compression == Gzip {
		return gzip.NewReader(reader)
	} else if compression == Zstd {
		return zstd.NewReader(reader)
	}

	return reader, nil
}

func NewReader(url string, config *StreamerConfig) (io.Reader, error) {
	streamer, err := NewStreamer(url, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create streamer: %w", err)
	}

	return &Reader{
		streamer: streamer,
	}, nil
}

func (r *Reader) Read(p []byte) (n int, err error) {
	// Load the buffer if it's empty or nil
	if r.buffer == nil || len(r.buffer) == 0 {
		// We've already read the whole buffer
		if r.isEOF {
			return 0, io.EOF
		}

		n, buf, err := r.streamer.Next()
		if err != nil {
			if err == io.EOF {
				r.isEOF = true
			} else {
				return 0, err
			}
		}

		r.buffer = make([]byte, n)
		copy(r.buffer, buf.Data[:n])

		// Return the buffer to the streamer
		r.streamer.Return(buf)
	}

	// Copy the data from the buffer to the given slice
	n = copy(p, r.buffer)
	r.buffer = r.buffer[n:]

	return n, nil
}

// Close the reader and stop all downloads
func (r *Reader) Close() {
	r.streamer.Close()
}
