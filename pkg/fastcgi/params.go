package fastcgi

import (
	"bytes"
	"encoding/binary"
	"io"
)

func ReadNameValuePairs(data []byte) (map[string]string, error) {
	params := make(map[string]string)
	r := bytes.NewReader(data)

	for r.Len() > 0 {
		nameLen, err := readLength(r)
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}

		valueLen, err := readLength(r)
		if err != nil {
			return nil, err
		}

		name := make([]byte, nameLen)
		if _, err := io.ReadFull(r, name); err != nil {
			return nil, err
		}

		value := make([]byte, valueLen)
		if _, err := io.ReadFull(r, value); err != nil {
			return nil, err
		}

		params[string(name)] = string(value)
	}

	return params, nil
}

func WriteNameValuePairs(params map[string]string) []byte {
	var buf bytes.Buffer

	for name, value := range params {
		writeLength(&buf, len(name))
		writeLength(&buf, len(value))
		buf.WriteString(name)
		buf.WriteString(value)
	}

	return buf.Bytes()
}

func readLength(r io.Reader) (int, error) {
	var b [1]byte
	if _, err := r.Read(b[:]); err != nil {
		return 0, err
	}

	if b[0]&0x80 == 0 {
		return int(b[0]), nil
	}

	var length uint32
	var bytes [3]byte
	if _, err := io.ReadFull(r, bytes[:]); err != nil {
		return 0, err
	}

	length = uint32(b[0]&0x7f)<<24 | uint32(bytes[0])<<16 | uint32(bytes[1])<<8 | uint32(bytes[2])
	return int(length), nil
}

func writeLength(w io.Writer, length int) {
	if length < 128 {
		w.Write([]byte{byte(length)})
	} else {
		var buf [4]byte
		binary.BigEndian.PutUint32(buf[:], uint32(length)|0x80000000)
		w.Write(buf[:])
	}
}