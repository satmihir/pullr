package stream

import "net/http"

type Size uint64
type Compression uint8

const (
	B  Size = 1
	KB      = 1024 * B
	MB      = 1024 * KB
	GB      = 1024 * MB

	NoCompression Compression = 0
	Gzip          Compression = 1
	Zstd          Compression = 2
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
	// Close the streamer and stop all downloads
	Close()
}

// A reader interface to fill the buffer with streamed data
// similar to bufio.Reader but with parallel async pulling
type IReader interface {
	// Read the data from the stream into the buffer
	Read(p []byte) (n int, err error)
	// Close the reader and stop all downloads
	Close()
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
