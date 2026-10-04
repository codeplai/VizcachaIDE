package dap

import (
	"github.com/google/go-dap"
)

const unsupportedMessage = "unsupported"

// answerReverse replies to a request the adapter sent to the client. Only runInTerminal is
// supported, and only when a handler exists; everything else gets an error response.
func (c *Client) answerReverse(request dap.RequestMessage) {
	header := request.GetRequest()
	terminal, isTerminal := request.(*dap.RunInTerminalRequest)
	if !isTerminal || c.reverse == nil {
		c.reply(&dap.ErrorResponse{Response: c.responseTo(header, false, unsupportedMessage)})
		return
	}
	processID, err := c.reverse.RunInTerminal(terminal.Arguments)
	if err != nil {
		c.reply(&dap.ErrorResponse{Response: c.responseTo(header, false, err.Error())})
		return
	}
	response := &dap.RunInTerminalResponse{Response: c.responseTo(header, true, "")}
	response.Body.ProcessId = processID
	c.reply(response)
}

func (c *Client) responseTo(request *dap.Request, success bool, message string) dap.Response {
	c.mu.Lock()
	c.seq++
	seq := c.seq
	c.mu.Unlock()
	return dap.Response{
		ProtocolMessage: dap.ProtocolMessage{Seq: seq, Type: "response"},
		Command:         request.Command,
		RequestSeq:      request.Seq,
		Success:         success,
		Message:         message,
	}
}

func (c *Client) reply(response dap.Message) {
	c.writeMu.Lock()
	defer c.writeMu.Unlock()
	_ = dap.WriteProtocolMessage(c.conn, response)
}
