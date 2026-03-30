package misc

import (
	"bytes"
	"sync"
)

type BufferedPipe struct {
	mu     sync.Mutex
	buffer *bytes.Buffer
	cond   *sync.Cond
	// closed bool
}

// NewBufferedPipe creates a new buffered pipe.
func NewBufferedPipe() *BufferedPipe {
	bp := &BufferedPipe{
		buffer: new(bytes.Buffer),
	}
	bp.cond = sync.NewCond(&bp.mu)
	return bp
}

// Write adds data to the buffer and signals any waiting readers.
func (bp *BufferedPipe) Write(p []byte) (n int, err error) {
	bp.mu.Lock()
	defer bp.mu.Unlock()

	// if bp.closed {
	// 	return 0, io.ErrClosedPipe
	// }

	n, err = bp.buffer.Write(p)
	bp.cond.Signal()
	return n, err
}

// Read reads data from the buffer and waits for more if the buffer is empty.
func (bp *BufferedPipe) Read(p []byte) (n int, err error) {
	bp.mu.Lock()
	defer bp.mu.Unlock()

	for {
		if n = bp.buffer.Len(); n > 0 {
			return bp.buffer.Read(p)
		}

		// if bp.closed {
		// 	return 0, io.EOF
		// }

		// Wait for new data to be written
		bp.cond.Wait()
	}
}

// Close marks the pipe as closed.
func (bp *BufferedPipe) Close() error {
	// bp.mu.Lock()
	// defer bp.mu.Unlock()
	// bp.closed = true
	// bp.cond.Broadcast()
	return nil
}
