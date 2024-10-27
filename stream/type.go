package stream

import "net/http"

type Size uint64

const (
	B  Size = 1
	KB      = 1024 * B
	MB      = 1024 * KB
	GB      = 1024 * MB
)

// An interface to stream content from a file
type IStreamer interface {
	// Get the next buffer from the stream. Can be called only once at a time.
	// n: bytes from the stream in the buffer
	// buf: buffer that contains streaed bytes
	// err: error, potentially an io.EOF or other errors
	Next() (n int, buf *Buffer, err error)
	// Return the buffer to the streamer to be included
	// in the pool to minimize allocations
	Return(buf *Buffer)
}

// An interface for HTTP operations
type IHTTPClient interface {
	// Perform a GET request to the given URL with the given headers
	Get(url string, headers map[string]string) (*http.Response, error)
}

// A buffer to store streamed data
type Buffer struct {
	Data []byte
}

type StreamerConfig struct {
	// The size of the buffer to read from the stream
	BufferSize Size
	// The number of parallel downloads
	Parallelism uint32
}
