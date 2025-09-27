package fastcgi

import (
	"bytes"
	"testing"
)

func TestRecordMarshaling(t *testing.T) {
	content := []byte("Hello, FastCGI!")
	rec := NewRecord(FCGI_STDOUT, 1, content)

	if rec.Version != FCGI_VERSION_1 {
		t.Errorf("Expected version %d, got %d", FCGI_VERSION_1, rec.Version)
	}

	if rec.Type != FCGI_STDOUT {
		t.Errorf("Expected type %d, got %d", FCGI_STDOUT, rec.Type)
	}

	if rec.RequestID != 1 {
		t.Errorf("Expected request ID 1, got %d", rec.RequestID)
	}

	if rec.ContentLength != uint16(len(content)) {
		t.Errorf("Expected content length %d, got %d", len(content), rec.ContentLength)
	}

	var buf bytes.Buffer
	if err := rec.WriteTo(&buf); err != nil {
		t.Fatalf("WriteTo failed: %v", err)
	}

	readRec, err := ReadRecord(&buf)
	if err != nil {
		t.Fatalf("ReadRecord failed: %v", err)
	}

	if !bytes.Equal(readRec.Content, content) {
		t.Errorf("Expected content %s, got %s", content, readRec.Content)
	}
}

func TestBeginRequestBody(t *testing.T) {
	body := &BeginRequestBody{
		Role:  FCGI_RESPONDER,
		Flags: FCGI_KEEP_CONN,
	}

	data := body.Marshal()
	if len(data) != 8 {
		t.Errorf("Expected 8 bytes, got %d", len(data))
	}

	unmarshaled := UnmarshalBeginRequest(data)
	if unmarshaled.Role != FCGI_RESPONDER {
		t.Errorf("Expected role %d, got %d", FCGI_RESPONDER, unmarshaled.Role)
	}

	if unmarshaled.Flags != FCGI_KEEP_CONN {
		t.Errorf("Expected flags %d, got %d", FCGI_KEEP_CONN, unmarshaled.Flags)
	}
}

func TestEndRequestBody(t *testing.T) {
	body := &EndRequestBody{
		AppStatus:      0,
		ProtocolStatus: FCGI_REQUEST_COMPLETE,
	}

	data := body.Marshal()
	if len(data) != 8 {
		t.Errorf("Expected 8 bytes, got %d", len(data))
	}

	unmarshaled := UnmarshalEndRequest(data)
	if unmarshaled.AppStatus != 0 {
		t.Errorf("Expected app status 0, got %d", unmarshaled.AppStatus)
	}

	if unmarshaled.ProtocolStatus != FCGI_REQUEST_COMPLETE {
		t.Errorf("Expected protocol status %d, got %d", FCGI_REQUEST_COMPLETE, unmarshaled.ProtocolStatus)
	}
}

func TestNameValuePairs(t *testing.T) {
	params := map[string]string{
		"REQUEST_METHOD":  "GET",
		"SCRIPT_FILENAME": "/var/www/index.php",
		"QUERY_STRING":    "foo=bar&baz=qux",
	}

	data := WriteNameValuePairs(params)

	parsed, err := ReadNameValuePairs(data)
	if err != nil {
		t.Fatalf("ReadNameValuePairs failed: %v", err)
	}

	if len(parsed) != len(params) {
		t.Errorf("Expected %d params, got %d", len(params), len(parsed))
	}

	for k, v := range params {
		if parsed[k] != v {
			t.Errorf("Param %s: expected %s, got %s", k, v, parsed[k])
		}
	}
}

func TestNameValuePairsLongLength(t *testing.T) {
	longValue := string(make([]byte, 200))
	params := map[string]string{
		"LONG_PARAM": longValue,
	}

	data := WriteNameValuePairs(params)

	parsed, err := ReadNameValuePairs(data)
	if err != nil {
		t.Fatalf("ReadNameValuePairs failed: %v", err)
	}

	if parsed["LONG_PARAM"] != longValue {
		t.Errorf("Long value not preserved")
	}
}

func TestRecordPadding(t *testing.T) {
	testCases := []int{0, 1, 7, 8, 9, 15, 16, 100}

	for _, contentLen := range testCases {
		content := make([]byte, contentLen)
		rec := NewRecord(FCGI_STDOUT, 1, content)

		expectedPadding := (-contentLen) & 7
		if int(rec.PaddingLength) != expectedPadding {
			t.Errorf("Content length %d: expected padding %d, got %d",
				contentLen, expectedPadding, rec.PaddingLength)
		}

		var buf bytes.Buffer
		if err := rec.WriteTo(&buf); err != nil {
			t.Fatalf("WriteTo failed: %v", err)
		}

		expectedTotal := FCGI_HEADER_LEN + contentLen + expectedPadding
		if buf.Len() != expectedTotal {
			t.Errorf("Content length %d: expected total %d, got %d",
				contentLen, expectedTotal, buf.Len())
		}
	}
}