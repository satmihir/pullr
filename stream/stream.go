package stream

import (
	"context"
	"fmt"
	"io"
	"sync"
)

// Implements IStreamer to pull a stream in parallel
// The source has to be a file at a URL that respects range header
// on the GET request and supports multiple range queries in parallel
// S3 presigned URL is one such example.
type Streamer struct {
	// Input parameters

	url         string
	bufferSize  uint64
	parallelism uint32
	fileSize    uint64

	// Internal state

	bufPool    *sync.Pool
	slots      []chan *sData
	curSlot    int
	httpClient IHTTPClient

	ctx        context.Context
	cancelFunc context.CancelFunc

	totalRead uint64
}

// Ceate a new streamer with the given parameters
func NewStreamer(url string, conf *StreamerConfig) (IStreamer, error) {
	httpClient := NewHttp()
	fileSize, err := getFileSize(url, httpClient)
	if err != nil {
		return nil, err
	}

	slots := make([]chan *sData, conf.Parallelism)
	for i := range slots {
		slots[i] = make(chan *sData)
	}

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)

	streamer := &Streamer{
		url:         url,
		bufferSize:  uint64(conf.BufferSize),
		parallelism: conf.Parallelism,
		fileSize:    fileSize,
		bufPool: &sync.Pool{
			New: func() interface{} {
				return &Buffer{make([]byte, conf.BufferSize)}
			},
		},
		slots:      slots,
		curSlot:    0,
		httpClient: httpClient,
		ctx:        ctx,
		cancelFunc: cancel,
	}

	for i := uint32(0); i < conf.Parallelism; i++ {
		go streamer.downloadChunks(ctx, i, slots[i])
	}

	return streamer, nil
}

// Output struct of a downloaded chunk
type sData struct {
	data *Buffer
	len  int
	err  error
}

func (s *Streamer) endRange(startRange uint64) uint64 {
	endRange := startRange + uint64(s.bufferSize)
	if endRange > s.fileSize {
		endRange = s.fileSize
	}

	return endRange
}

func (s *Streamer) downloadInto(ctx context.Context, buf *Buffer, startRange, endRange uint64) error {
	headers := map[string]string{
		"Range": fmt.Sprintf("bytes=%d-%d", startRange, endRange-1),
	}

	resp, err := s.httpClient.Get(s.url, headers)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 206 {
		return fmt.Errorf("HTTP status code %d", resp.StatusCode)
	}

	readData := 0
	for readData < int(endRange-startRange) {
		select {
		case <-ctx.Done():
			return fmt.Errorf("context cancelled")
		default:
			n, err := resp.Body.Read(buf.Data[readData:])
			if err != nil {
				return err
			}

			readData += n
		}
	}

	return nil
}

func (s *Streamer) downloadChunks(ctx context.Context, slotNum uint32, slot chan *sData) {
	startRange := uint64(slotNum) * s.bufferSize
	endRange := s.endRange(startRange)

	for {
		select {
		case <-ctx.Done():
			return
		default:
			buf := s.bufPool.Get().(*Buffer)
			err := s.downloadInto(ctx, buf, startRange, endRange)
			slot <- &sData{
				data: buf,
				len:  int(endRange - startRange),
				err:  err,
			}

			startRange = startRange + s.bufferSize*uint64(s.parallelism)
			if startRange >= s.fileSize {
				// No more data to download
				return
			}
			endRange = s.endRange(startRange)
		}
	}
}

func (s *Streamer) Next() (int, *Buffer, error) {
	data := <-s.slots[s.curSlot]
	if data.err != nil && data.err != io.EOF {
		return 0, nil, data.err
	}
	s.totalRead += uint64(data.len)

	if s.totalRead == uint64(s.fileSize) {
		return data.len, data.data, io.EOF
	}

	s.curSlot = (s.curSlot + 1) % int(s.parallelism)
	return data.len, data.data, nil
}

func (s *Streamer) Close() {
	// Cancel the context to stop all downloads
	s.cancelFunc()

	// Close all slots
	for i := range s.slots {
		close(s.slots[i])
	}
}

func (s *Streamer) Return(buffer *Buffer) {
	s.bufPool.Put(buffer)
}

// Get the size of the file using the content-length header against a GET request
func getFileSize(url string, httpClient IHTTPClient) (uint64, error) {
	resp, err := httpClient.Get(url, nil)
	if err != nil {
		return 0, fmt.Errorf("HTTP GET error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return 0, fmt.Errorf("HTTP status code %d", resp.StatusCode)
	}

	return uint64(resp.ContentLength), nil
}
