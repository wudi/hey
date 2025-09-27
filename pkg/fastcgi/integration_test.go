package fastcgi

import (
	"bytes"
	"net"
	"testing"
	"time"
)

func TestIntegrationFastCGIRequest(t *testing.T) {
	t.Skip("Integration test - requires manual testing with real FPM server")

	conn, err := net.DialTimeout("tcp", "127.0.0.1:9000", 5*time.Second)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	proto := NewProtocol(conn)

	beginReq := &BeginRequestBody{
		Role:  FCGI_RESPONDER,
		Flags: FCGI_KEEP_CONN,
	}

	rec := NewRecord(FCGI_BEGIN_REQUEST, 1, beginReq.Marshal())
	if err := proto.WriteRecord(rec); err != nil {
		t.Fatalf("Failed to write BEGIN_REQUEST: %v", err)
	}

	params := map[string]string{
		"SCRIPT_FILENAME": "/var/www/test.php",
		"REQUEST_METHOD":  "GET",
		"QUERY_STRING":    "foo=bar",
		"SERVER_PROTOCOL": "HTTP/1.1",
	}

	paramsData := WriteNameValuePairs(params)
	paramsRec := NewRecord(FCGI_PARAMS, 1, paramsData)
	if err := proto.WriteRecord(paramsRec); err != nil {
		t.Fatalf("Failed to write PARAMS: %v", err)
	}

	emptyParamsRec := NewRecord(FCGI_PARAMS, 1, nil)
	if err := proto.WriteRecord(emptyParamsRec); err != nil {
		t.Fatalf("Failed to write empty PARAMS: %v", err)
	}

	emptyStdinRec := NewRecord(FCGI_STDIN, 1, nil)
	if err := proto.WriteRecord(emptyStdinRec); err != nil {
		t.Fatalf("Failed to write empty STDIN: %v", err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	for {
		rec, err := proto.ReadRecord()
		if err != nil {
			t.Fatalf("Failed to read record: %v", err)
		}

		switch rec.Type {
		case FCGI_STDOUT:
			if rec.ContentLength > 0 {
				stdout.Write(rec.Content)
			} else {
				t.Logf("STDOUT complete")
			}

		case FCGI_STDERR:
			if rec.ContentLength > 0 {
				stderr.Write(rec.Content)
			} else {
				t.Logf("STDERR complete")
			}

		case FCGI_END_REQUEST:
			endReq := UnmarshalEndRequest(rec.Content)
			t.Logf("Request ended - AppStatus: %d, ProtocolStatus: %d", endReq.AppStatus, endReq.ProtocolStatus)
			t.Logf("STDOUT: %s", stdout.String())
			if stderr.Len() > 0 {
				t.Logf("STDERR: %s", stderr.String())
			}
			return
		}
	}
}