package fastcgi

import (
	"encoding/binary"
	"io"
)

const (
	FCGI_VERSION_1 = 1

	FCGI_BEGIN_REQUEST     = 1
	FCGI_ABORT_REQUEST     = 2
	FCGI_END_REQUEST       = 3
	FCGI_PARAMS            = 4
	FCGI_STDIN             = 5
	FCGI_STDOUT            = 6
	FCGI_STDERR            = 7
	FCGI_DATA              = 8
	FCGI_GET_VALUES        = 9
	FCGI_GET_VALUES_RESULT = 10
	FCGI_UNKNOWN_TYPE      = 11

	FCGI_RESPONDER  = 1
	FCGI_AUTHORIZER = 2
	FCGI_FILTER     = 3

	FCGI_REQUEST_COMPLETE = 0
	FCGI_CANT_MPX_CONN    = 1
	FCGI_OVERLOADED       = 2
	FCGI_UNKNOWN_ROLE     = 3

	FCGI_KEEP_CONN = 1

	FCGI_NULL_REQUEST_ID = 0

	FCGI_HEADER_LEN = 8
)

type Record struct {
	Version       uint8
	Type          uint8
	RequestID     uint16
	ContentLength uint16
	PaddingLength uint8
	Reserved      uint8
	Content       []byte
}

func NewRecord(recType uint8, requestID uint16, content []byte) *Record {
	contentLen := len(content)
	paddingLen := -contentLen & 7

	return &Record{
		Version:       FCGI_VERSION_1,
		Type:          recType,
		RequestID:     requestID,
		ContentLength: uint16(contentLen),
		PaddingLength: uint8(paddingLen),
		Reserved:      0,
		Content:       content,
	}
}

func ReadRecord(r io.Reader) (*Record, error) {
	var header [FCGI_HEADER_LEN]byte
	if _, err := io.ReadFull(r, header[:]); err != nil {
		return nil, err
	}

	rec := &Record{
		Version:       header[0],
		Type:          header[1],
		RequestID:     binary.BigEndian.Uint16(header[2:4]),
		ContentLength: binary.BigEndian.Uint16(header[4:6]),
		PaddingLength: header[6],
		Reserved:      header[7],
	}

	if rec.ContentLength > 0 {
		rec.Content = make([]byte, rec.ContentLength)
		if _, err := io.ReadFull(r, rec.Content); err != nil {
			return nil, err
		}
	}

	if rec.PaddingLength > 0 {
		padding := make([]byte, rec.PaddingLength)
		if _, err := io.ReadFull(r, padding); err != nil {
			return nil, err
		}
	}

	return rec, nil
}

func (rec *Record) WriteTo(w io.Writer) error {
	header := [FCGI_HEADER_LEN]byte{
		rec.Version,
		rec.Type,
		byte(rec.RequestID >> 8),
		byte(rec.RequestID),
		byte(rec.ContentLength >> 8),
		byte(rec.ContentLength),
		rec.PaddingLength,
		rec.Reserved,
	}

	if _, err := w.Write(header[:]); err != nil {
		return err
	}

	if rec.ContentLength > 0 {
		if _, err := w.Write(rec.Content); err != nil {
			return err
		}
	}

	if rec.PaddingLength > 0 {
		padding := make([]byte, rec.PaddingLength)
		if _, err := w.Write(padding); err != nil {
			return err
		}
	}

	return nil
}

type BeginRequestBody struct {
	Role     uint16
	Flags    uint8
	Reserved [5]byte
}

func (b *BeginRequestBody) Marshal() []byte {
	data := make([]byte, 8)
	binary.BigEndian.PutUint16(data[0:2], b.Role)
	data[2] = b.Flags
	return data
}

func UnmarshalBeginRequest(data []byte) *BeginRequestBody {
	if len(data) < 8 {
		return nil
	}
	return &BeginRequestBody{
		Role:  binary.BigEndian.Uint16(data[0:2]),
		Flags: data[2],
	}
}

type EndRequestBody struct {
	AppStatus      uint32
	ProtocolStatus uint8
	Reserved       [3]byte
}

func (e *EndRequestBody) Marshal() []byte {
	data := make([]byte, 8)
	binary.BigEndian.PutUint32(data[0:4], e.AppStatus)
	data[4] = e.ProtocolStatus
	return data
}

func UnmarshalEndRequest(data []byte) *EndRequestBody {
	if len(data) < 8 {
		return nil
	}
	return &EndRequestBody{
		AppStatus:      binary.BigEndian.Uint32(data[0:4]),
		ProtocolStatus: data[4],
	}
}