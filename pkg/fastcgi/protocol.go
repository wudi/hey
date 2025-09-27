package fastcgi

import (
	"bufio"
	"bytes"
	"io"
	"net"
)

type Protocol struct {
	conn   net.Conn
	reader *bufio.Reader
	writer *bufio.Writer
}

func NewProtocol(conn net.Conn) *Protocol {
	return &Protocol{
		conn:   conn,
		reader: bufio.NewReader(conn),
		writer: bufio.NewWriter(conn),
	}
}

func (p *Protocol) ReadRecord() (*Record, error) {
	return ReadRecord(p.reader)
}

func (p *Protocol) WriteRecord(rec *Record) error {
	if err := rec.WriteTo(p.writer); err != nil {
		return err
	}
	return p.writer.Flush()
}

func (p *Protocol) ReadParams(requestID uint16) (map[string]string, error) {
	var buf bytes.Buffer

	for {
		rec, err := p.ReadRecord()
		if err != nil {
			return nil, err
		}

		if rec.RequestID != requestID {
			continue
		}

		if rec.Type != FCGI_PARAMS {
			continue
		}

		if rec.ContentLength == 0 {
			break
		}

		buf.Write(rec.Content)
	}

	return ReadNameValuePairs(buf.Bytes())
}

func (p *Protocol) ReadStdin(requestID uint16) ([]byte, error) {
	var buf bytes.Buffer

	for {
		rec, err := p.ReadRecord()
		if err != nil {
			return nil, err
		}

		if rec.RequestID != requestID {
			continue
		}

		if rec.Type != FCGI_STDIN {
			continue
		}

		if rec.ContentLength == 0 {
			break
		}

		buf.Write(rec.Content)
	}

	return buf.Bytes(), nil
}

func (p *Protocol) WriteStdout(requestID uint16, data []byte) error {
	const maxWrite = 65535

	for len(data) > 0 {
		chunk := data
		if len(chunk) > maxWrite {
			chunk = chunk[:maxWrite]
		}
		data = data[len(chunk):]

		rec := NewRecord(FCGI_STDOUT, requestID, chunk)
		if err := p.WriteRecord(rec); err != nil {
			return err
		}
	}

	rec := NewRecord(FCGI_STDOUT, requestID, nil)
	return p.WriteRecord(rec)
}

func (p *Protocol) WriteStderr(requestID uint16, data []byte) error {
	const maxWrite = 65535

	for len(data) > 0 {
		chunk := data
		if len(chunk) > maxWrite {
			chunk = chunk[:maxWrite]
		}
		data = data[len(chunk):]

		rec := NewRecord(FCGI_STDERR, requestID, chunk)
		if err := p.WriteRecord(rec); err != nil {
			return err
		}
	}

	rec := NewRecord(FCGI_STDERR, requestID, nil)
	return p.WriteRecord(rec)
}

func (p *Protocol) EndRequest(requestID uint16, appStatus uint32, protocolStatus uint8) error {
	body := &EndRequestBody{
		AppStatus:      appStatus,
		ProtocolStatus: protocolStatus,
	}

	rec := NewRecord(FCGI_END_REQUEST, requestID, body.Marshal())
	return p.WriteRecord(rec)
}

func (p *Protocol) Close() error {
	return p.conn.Close()
}

type Request struct {
	ID     uint16
	Role   uint16
	Flags  uint8
	Params map[string]string
	Stdin  []byte
}

func (p *Protocol) ReadRequest() (*Request, error) {
	rec, err := p.ReadRecord()
	if err != nil {
		return nil, err
	}

	if rec.Type != FCGI_BEGIN_REQUEST {
		return nil, io.ErrUnexpectedEOF
	}

	beginReq := UnmarshalBeginRequest(rec.Content)
	if beginReq == nil {
		return nil, io.ErrUnexpectedEOF
	}

	req := &Request{
		ID:    rec.RequestID,
		Role:  beginReq.Role,
		Flags: beginReq.Flags,
	}

	params, err := p.ReadParams(req.ID)
	if err != nil {
		return nil, err
	}
	req.Params = params

	stdin, err := p.ReadStdin(req.ID)
	if err != nil {
		return nil, err
	}
	req.Stdin = stdin

	return req, nil
}

func (p *Protocol) SendResponse(requestID uint16, stdout, stderr []byte, exitCode int) error {
	if len(stdout) > 0 {
		if err := p.WriteStdout(requestID, stdout); err != nil {
			return err
		}
	}

	if len(stderr) > 0 {
		if err := p.WriteStderr(requestID, stderr); err != nil {
			return err
		}
	}

	return p.EndRequest(requestID, uint32(exitCode), FCGI_REQUEST_COMPLETE)
}