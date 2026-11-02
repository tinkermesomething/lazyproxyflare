package cloudflare

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"lazyproxyflare/internal/cmdlog"
)

// newTestClient points a Client at the test server and removes retry sleeps.
func newTestClient(t *testing.T, srv *httptest.Server) *Client {
	t.Helper()
	c := NewClient("test-token")
	c.baseURL = srv.URL
	c.wait = func(time.Duration) {}
	c.log = cmdlog.New(50)
	return c
}

const okListBody = `{"success":true,"errors":[],"result":[{"id":"r1","type":"CNAME","name":"a.example.com","content":"b.example.com"}],"result_info":{"page":1,"total_pages":1}}`
const okRecordBody = `{"success":true,"errors":[],"result":{"id":"r1","type":"CNAME","name":"a.example.com","content":"b.example.com"}}`

func TestListDNSRecordsSendsAuthAndEscapesQuery(t *testing.T) {
	var gotAuth, gotRawQuery string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		gotRawQuery = r.URL.RawQuery
		io.WriteString(w, okListBody)
	}))
	defer srv.Close()

	records, err := newTestClient(t, srv).ListDNSRecords("abcdef0123456789abcdef0123456789", "A&x=1")
	if err != nil {
		t.Fatal(err)
	}
	if len(records) != 1 || records[0].ID != "r1" {
		t.Errorf("unexpected records: %+v", records)
	}
	if gotAuth != "Bearer test-token" {
		t.Errorf("Authorization = %q", gotAuth)
	}
	if !strings.Contains(gotRawQuery, "type=A%26x%3D1") {
		t.Errorf("record type not escaped, raw query = %q", gotRawQuery)
	}
}

func TestZoneAndRecordIDsAreEscaped(t *testing.T) {
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.EscapedPath()
		io.WriteString(w, `{"success":true,"errors":[],"result":{}}`)
	}))
	defer srv.Close()

	if err := newTestClient(t, srv).DeleteDNSRecord("zone/../x", "rec?id"); err != nil {
		t.Fatal(err)
	}
	want := "/zones/zone%2F..%2Fx/dns_records/rec%3Fid"
	if gotPath != want {
		t.Errorf("path = %q, want %q", gotPath, want)
	}
}

func TestRetriesOn429AndReplaysBody(t *testing.T) {
	var calls int32
	var bodies []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		bodies = append(bodies, string(b))
		if atomic.AddInt32(&calls, 1) == 1 {
			w.Header().Set("Retry-After", "1")
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		io.WriteString(w, okRecordBody)
	}))
	defer srv.Close()

	rec, err := newTestClient(t, srv).CreateDNSRecord("z", DNSRecord{Type: "CNAME", Name: "a.example.com", Content: "b.example.com"})
	if err != nil {
		t.Fatal(err)
	}
	if rec.ID != "r1" {
		t.Errorf("unexpected record: %+v", rec)
	}
	if calls != 2 {
		t.Fatalf("calls = %d, want 2", calls)
	}
	if bodies[0] == "" || bodies[0] != bodies[1] {
		t.Errorf("body not replayed intact: %q vs %q", bodies[0], bodies[1])
	}
}

func TestGivesUpAfterMaxAttempts(t *testing.T) {
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer srv.Close()

	_, err := newTestClient(t, srv).ListDNSRecords("z", "")
	if err == nil || !strings.Contains(err.Error(), "429") {
		t.Fatalf("expected 429 error, got %v", err)
	}
	if int(calls) != maxAttempts {
		t.Errorf("calls = %d, want %d", calls, maxAttempts)
	}
}

func TestNon429ErrorIsNotRetried(t *testing.T) {
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		w.WriteHeader(http.StatusForbidden)
		io.WriteString(w, `{"success":false,"errors":[{"code":9109,"message":"Invalid access token"}]}`)
	}))
	defer srv.Close()

	_, err := newTestClient(t, srv).ListDNSRecords("z", "")
	if err == nil || !strings.Contains(err.Error(), "403") {
		t.Fatalf("expected 403 error, got %v", err)
	}
	if calls != 1 {
		t.Errorf("calls = %d, want 1", calls)
	}
}

func TestResponseBodyIsCapped(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		chunk := []byte(strings.Repeat("x", 1<<20))
		for i := 0; i < (maxResponseBytes>>20)+2; i++ {
			w.Write(chunk)
		}
	}))
	defer srv.Close()

	_, err := newTestClient(t, srv).ListDNSRecords("z", "")
	if err == nil || !strings.Contains(err.Error(), "too large") {
		t.Fatalf("expected size error, got %v", err)
	}
}

func TestRequestsAreRecordedWithoutSecrets(t *testing.T) {
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if atomic.AddInt32(&calls, 1) == 1 {
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		io.WriteString(w, okRecordBody)
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	if _, err := c.CreateDNSRecord("zone1", DNSRecord{Type: "CNAME", Name: "secret-name.example.com", Content: "b.example.com"}); err != nil {
		t.Fatal(err)
	}

	entries := c.log.Entries()
	if len(entries) != 2 {
		t.Fatalf("got %d entries, want 2 (one per attempt): %+v", len(entries), entries)
	}
	if entries[0].Status != "429" || entries[1].Status != "200" {
		t.Errorf("statuses = %q, %q; want 429, 200", entries[0].Status, entries[1].Status)
	}
	for _, e := range entries {
		if e.Kind != cmdlog.API || e.Summary != "POST /zones/zone1/dns_records" {
			t.Errorf("unexpected entry: %+v", e)
		}
		all := e.Summary + e.Status + e.Err
		if strings.Contains(all, "test-token") || strings.Contains(all, "secret-name") {
			t.Errorf("entry leaks token or request body: %+v", e)
		}
	}
}

func TestFailedRequestsRecordTheError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		io.WriteString(w, `{"errors":[{"message":"Invalid access token"}]}`)
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	if _, err := c.ListDNSRecords("z", ""); err == nil {
		t.Fatal("expected error")
	}
	e := c.log.Entries()[0]
	if e.Status != "403" || !strings.Contains(e.Err, "Invalid access token") {
		t.Errorf("failure not recorded: %+v", e)
	}

	// A connection that cannot be made is recorded too.
	srv.Close()
	if _, err := c.ListDNSRecords("z", ""); err == nil {
		t.Fatal("expected error")
	}
	entries := c.log.Entries()
	if last := entries[len(entries)-1]; last.Status != "error" || last.Err == "" {
		t.Errorf("transport failure not recorded: %+v", last)
	}
}
