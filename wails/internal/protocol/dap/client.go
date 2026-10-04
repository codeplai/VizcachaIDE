package dap

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"sync"

	"github.com/google/go-dap"
)

// errConnectionClosed is returned to requests that were waiting when the adapter hung up.
var errConnectionClosed = errors.New("debug connection closed")

// Client speaks DAP over one connection. Responses are matched to requests by
// sequence number; events go to the handler, in order, from the reading goroutine,
// so the handler must not block.
type Client struct {
	conn    io.ReadWriteCloser
	reader  *bufio.Reader
	onEvent func(dap.EventMessage)
	onClose func()
	reverse ReverseHandler

	writeMu sync.Mutex
	mu      sync.Mutex
	seq     int
	pending map[int]chan dap.ResponseMessage
	closed  bool
}

// NewClient starts reading from conn. onClose runs once when the connection ends. reverse
// answers the requests the adapter sends (nil answers all of them "unsupported").
func NewClient(conn io.ReadWriteCloser, onEvent func(dap.EventMessage), onClose func(), reverse ReverseHandler) *Client {
	client := &Client{
		conn:    conn,
		reader:  bufio.NewReader(conn),
		onEvent: onEvent,
		onClose: onClose,
		reverse: reverse,
		pending: map[int]chan dap.ResponseMessage{},
	}
	go client.readLoop()
	return client
}

// Call sends a request and waits for its response. An unsuccessful response is
// returned as an error built from its message.
func (c *Client) Call(ctx context.Context, request dap.RequestMessage) (dap.ResponseMessage, error) {
	answer := make(chan dap.ResponseMessage, 1)
	seq, err := c.send(request, answer)
	if err != nil {
		return nil, err
	}
	select {
	case response, ok := <-answer:
		if !ok {
			return nil, errConnectionClosed
		}
		return response, responseError(response)
	case <-ctx.Done():
		c.forget(seq)
		return nil, fmt.Errorf("waiting for %s: %w", request.GetRequest().Command, ctx.Err())
	}
}

func (c *Client) send(request dap.RequestMessage, answer chan dap.ResponseMessage) (int, error) {
	c.mu.Lock()
	if c.closed {
		c.mu.Unlock()
		return 0, errConnectionClosed
	}
	c.seq++
	seq := c.seq
	request.GetRequest().Seq = seq
	c.pending[seq] = answer
	c.mu.Unlock()

	c.writeMu.Lock()
	err := dap.WriteProtocolMessage(c.conn, request)
	c.writeMu.Unlock()
	if err != nil {
		c.forget(seq)
		return 0, fmt.Errorf("sending %s: %w", request.GetRequest().Command, err)
	}
	return seq, nil
}

func (c *Client) forget(seq int) {
	c.mu.Lock()
	delete(c.pending, seq)
	c.mu.Unlock()
}

// Close hangs up. Pending calls fail with errConnectionClosed.
func (c *Client) Close() {
	_ = c.conn.Close()
}

func (c *Client) readLoop() {
	for {
		message, err := dap.ReadProtocolMessage(c.reader)
		if err != nil {
			c.shutdown()
			return
		}
		switch typed := message.(type) {
		case dap.ResponseMessage:
			c.deliver(typed)
		case dap.EventMessage:
			c.onEvent(typed)
		case dap.RequestMessage:
			go c.answerReverse(typed) // may take long: never block the reading goroutine
		}
	}
}

func (c *Client) deliver(response dap.ResponseMessage) {
	seq := response.GetResponse().RequestSeq
	c.mu.Lock()
	answer, found := c.pending[seq]
	delete(c.pending, seq)
	c.mu.Unlock()
	if found {
		answer <- response
	}
}

func (c *Client) shutdown() {
	c.mu.Lock()
	c.closed = true
	for seq, answer := range c.pending {
		close(answer)
		delete(c.pending, seq)
	}
	c.mu.Unlock()
	_ = c.conn.Close()
	c.onClose()
}

// responseError turns a failed response into an error with the adapter's own message.
func responseError(response dap.ResponseMessage) error {
	base := response.GetResponse()
	if base.Success {
		return nil
	}
	if failure, ok := response.(*dap.ErrorResponse); ok && failure.Body.Error != nil {
		return fmt.Errorf("%s failed: %s", base.Command, failure.Body.Error.Format)
	}
	return fmt.Errorf("%s failed: %s", base.Command, base.Message)
}
