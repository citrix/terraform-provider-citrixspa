package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"testing"
)

// routingDomainsBackend fakes /routingDomains with `total` entries, mirroring
// the behaviour measured on a live 4116-entry tenant: `limit` is an exclusive
// end index (entries [offset, limit)), the window is capped at
// routingDomainsPageSize, a missing limit means an end of
// routingDomainsPageSize, and limit < offset is a 500. Every request's raw
// offset/limit is recorded so tests can assert the paging sequence.
func routingDomainsBackend(t *testing.T, total int, requests *[][2]int) func(int, *http.Request) (int, string) {
	t.Helper()
	return routingDomainsBackendWith(t, total, total, routingDomainsPageSize, requests)
}

// routingDomainsBackendWith is routingDomainsBackend with the stored entry
// count, the reported totalNum (negative omits it) and the per-request cap
// set independently, to model a backend that disagrees with the client's
// assumptions.
func routingDomainsBackendWith(t *testing.T, total, reportedTotal, pageCap int, requests *[][2]int) func(int, *http.Request) (int, string) {
	t.Helper()
	return func(_ int, req *http.Request) (int, string) {
		q := req.URL.Query()
		offset, _ := strconv.Atoi(q.Get("offset"))
		end := pageCap
		recordedLimit := -1
		if v := q.Get("limit"); v != "" {
			end, _ = strconv.Atoi(v)
			recordedLimit = end
		}
		*requests = append(*requests, [2]int{offset, recordedLimit})
		if end < offset {
			return http.StatusInternalServerError, `{"detail":"Internal Server Error"}`
		}
		end = min(end, offset+pageCap, total)

		items := []RoutingDomain{}
		for i := offset; i < end; i++ {
			items = append(items, RoutingDomain{FQDN: fmt.Sprintf("rd-%05d.example.com", i), Type: "internal"})
		}
		resp := map[string]any{"items": items}
		if reportedTotal >= 0 {
			resp["totalNum"] = reportedTotal
		}
		body, _ := json.Marshal(resp)
		return http.StatusOK, string(body)
	}
}

func TestGetRoutingDomains_Paging(t *testing.T) {
	tests := []struct {
		name         string
		total        int
		offset       int
		limit        int
		wantCount    int
		wantRequests [][2]int
	}{
		{
			name:         "no limit, more than one page",
			total:        4116,
			limit:        -1,
			wantCount:    4116,
			wantRequests: [][2]int{{0, 4000}, {4000, 8000}},
		},
		{
			name:         "no limit, fits in one page",
			total:        25,
			limit:        -1,
			wantCount:    25,
			wantRequests: [][2]int{{0, 4000}},
		},
		{
			name:         "no limit, exact multiple of page size stops on totalNum",
			total:        8000,
			limit:        -1,
			wantCount:    8000,
			wantRequests: [][2]int{{0, 4000}, {4000, 8000}},
		},
		{
			name:         "no limit, empty tenant",
			total:        0,
			limit:        -1,
			wantCount:    0,
			wantRequests: [][2]int{{0, 4000}},
		},
		{
			name:         "explicit window within one page is passed through",
			total:        4116,
			limit:        10,
			wantCount:    10,
			wantRequests: [][2]int{{0, 10}},
		},
		{
			name:         "explicit window at a non-zero offset is passed through",
			total:        4116,
			offset:       100,
			limit:        110,
			wantCount:    10,
			wantRequests: [][2]int{{100, 110}},
		},
		{
			name:         "explicit window of exactly one page is passed through",
			total:        9000,
			offset:       4000,
			limit:        8000,
			wantCount:    4000,
			wantRequests: [][2]int{{4000, 8000}},
		},
		{
			name:         "explicit window above page size is paged up to the end index",
			total:        9000,
			limit:        5000,
			wantCount:    5000,
			wantRequests: [][2]int{{0, 4000}, {4000, 5000}},
		},
		{
			name:         "no limit, starting offset is honoured",
			total:        4116,
			offset:       100,
			limit:        -1,
			wantCount:    4016,
			wantRequests: [][2]int{{100, 4100}, {4100, 8100}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var requests [][2]int
			client, _ := newTestClientWithResponder(routingDomainsBackend(t, tt.total, &requests))

			result, err := client.GetRoutingDomains(context.Background(), tt.offset, tt.limit)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got := len(result.RoutingDomains); got != tt.wantCount {
				t.Errorf("expected %d routing domains, got %d", tt.wantCount, got)
			}
			if result.Total != tt.total {
				t.Errorf("expected Total %d, got %d", tt.total, result.Total)
			}
			if fmt.Sprint(requests) != fmt.Sprint(tt.wantRequests) {
				t.Errorf("expected requests (offset, limit) %v, got %v", tt.wantRequests, requests)
			}
			seen := make(map[string]bool)
			for _, rd := range result.RoutingDomains {
				if seen[rd.FQDN] {
					t.Fatalf("duplicate routing domain %q in result", rd.FQDN)
				}
				seen[rd.FQDN] = true
			}
		})
	}
}

// TestGetRoutingDomains_LowerBackendCapStillCompletes asserts a short page is
// not taken as the end of data while totalNum says entries remain, so a
// backend whose real per-request cap is below routingDomainsPageSize still
// yields the complete list.
func TestGetRoutingDomains_LowerBackendCapStillCompletes(t *testing.T) {
	var requests [][2]int
	client, _ := newTestClientWithResponder(routingDomainsBackendWith(t, 4116, 4116, 1500, &requests))

	result, err := client.GetRoutingDomains(context.Background(), 0, -1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := len(result.RoutingDomains); got != 4116 {
		t.Errorf("expected 4116 routing domains, got %d", got)
	}
	if want := [][2]int{{0, 4000}, {1500, 5500}, {3000, 7000}}; fmt.Sprint(requests) != fmt.Sprint(want) {
		t.Errorf("expected requests (offset, limit) %v, got %v", want, requests)
	}
}

// TestGetRoutingDomains_OverReportedTotalStopsOnEmptyPage asserts that a
// totalNum larger than what the backend actually serves costs exactly one
// extra, empty request rather than looping.
func TestGetRoutingDomains_OverReportedTotalStopsOnEmptyPage(t *testing.T) {
	var requests [][2]int
	client, _ := newTestClientWithResponder(routingDomainsBackendWith(t, 10, 9999, routingDomainsPageSize, &requests))

	result, err := client.GetRoutingDomains(context.Background(), 0, -1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := len(result.RoutingDomains); got != 10 {
		t.Errorf("expected 10 routing domains, got %d", got)
	}
	if want := [][2]int{{0, 4000}, {10, 4010}}; fmt.Sprint(requests) != fmt.Sprint(want) {
		t.Errorf("expected requests (offset, limit) %v, got %v", want, requests)
	}
}

// TestGetRoutingDomains_ShortPageWithoutTotalStops asserts that when the
// backend omits totalNum, a short page is still treated as the end of data.
func TestGetRoutingDomains_ShortPageWithoutTotalStops(t *testing.T) {
	var requests [][2]int
	client, _ := newTestClientWithResponder(routingDomainsBackendWith(t, 25, -1, routingDomainsPageSize, &requests))

	result, err := client.GetRoutingDomains(context.Background(), 0, -1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := len(result.RoutingDomains); got != 25 {
		t.Errorf("expected 25 routing domains, got %d", got)
	}
	if want := [][2]int{{0, 4000}}; fmt.Sprint(requests) != fmt.Sprint(want) {
		t.Errorf("expected requests (offset, limit) %v, got %v", want, requests)
	}
}

// TestGetRoutingDomains_PageErrorPropagates asserts a failure on a later page
// is returned as an error instead of a silently truncated result.
func TestGetRoutingDomains_PageErrorPropagates(t *testing.T) {
	var requests [][2]int
	backend := routingDomainsBackend(t, 4116, &requests)
	client, _ := newTestClientWithResponder(func(n int, req *http.Request) (int, string) {
		if n == 2 {
			return http.StatusInternalServerError, `{"error":"boom"}`
		}
		return backend(n, req)
	})

	if _, err := client.GetRoutingDomains(context.Background(), 0, -1); err == nil {
		t.Fatal("expected error from failed second page, got nil")
	}
}

// TestGetRoutingDomains_LimitBelowOffsetMatchesAPI asserts limit < offset is
// passed through unchanged, so callers get the API's own error rather than a
// provider reinterpretation.
func TestGetRoutingDomains_LimitBelowOffsetMatchesAPI(t *testing.T) {
	var requests [][2]int
	client, _ := newTestClientWithResponder(routingDomainsBackend(t, 4116, &requests))

	if _, err := client.GetRoutingDomains(context.Background(), 100, 10); err == nil {
		t.Fatal("expected the API error for limit < offset, got nil")
	}
	if want := [][2]int{{100, 10}}; fmt.Sprint(requests) != fmt.Sprint(want) {
		t.Errorf("expected requests (offset, limit) %v, got %v", want, requests)
	}
}

// TestGetRoutingDomains_PageSizeBackendIsTruncated asserts a paged read never
// takes more than each window, should the backend ever read limit as a page
// size rather than an end index.
func TestGetRoutingDomains_PageSizeBackendIsTruncated(t *testing.T) {
	const total = 9000
	client, _ := newTestClientWithResponder(func(_ int, req *http.Request) (int, string) {
		offset, _ := strconv.Atoi(req.URL.Query().Get("offset"))
		limit, _ := strconv.Atoi(req.URL.Query().Get("limit"))
		items := []RoutingDomain{}
		for i := offset; i < min(offset+limit, total); i++ {
			items = append(items, RoutingDomain{FQDN: fmt.Sprintf("rd-%05d.example.com", i)})
		}
		body, _ := json.Marshal(map[string]any{"items": items, "totalNum": total})
		return http.StatusOK, string(body)
	})

	result, err := client.GetRoutingDomains(context.Background(), 0, 5000)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := len(result.RoutingDomains); got != 5000 {
		t.Fatalf("expected 5000 routing domains, got %d", got)
	}
	if last := result.RoutingDomains[len(result.RoutingDomains)-1].FQDN; last != "rd-04999.example.com" {
		t.Errorf("expected last entry rd-04999.example.com, got %s", last)
	}
}

// TestGetRoutingDomains_InsertMidListingDropsRepeat asserts that an entry
// inserted ahead of the read position between pages, which pushes the last
// entry of page one into page two, is not listed twice and does not stall
// paging: the offset must still advance past the repeated entry.
func TestGetRoutingDomains_InsertMidListingDropsRepeat(t *testing.T) {
	names := make([]string, 4116)
	for i := range names {
		names[i] = fmt.Sprintf("rd-%05d.example.com", i)
	}
	var requests [][2]int
	client, _ := newTestClientWithResponder(func(n int, req *http.Request) (int, string) {
		offset, _ := strconv.Atoi(req.URL.Query().Get("offset"))
		limit, _ := strconv.Atoi(req.URL.Query().Get("limit"))
		requests = append(requests, [2]int{offset, limit})
		if n == 2 {
			names = append([]string{"inserted.example.com"}, names...)
		}
		items := []RoutingDomain{}
		for i := offset; i < min(limit, offset+routingDomainsPageSize, len(names)); i++ {
			items = append(items, RoutingDomain{FQDN: names[i]})
		}
		body, _ := json.Marshal(map[string]any{"items": items, "totalNum": len(names)})
		return http.StatusOK, string(body)
	})

	result, err := client.GetRoutingDomains(context.Background(), 0, -1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	seen := make(map[string]bool)
	for _, rd := range result.RoutingDomains {
		if seen[rd.FQDN] {
			t.Fatalf("duplicate routing domain %q in result", rd.FQDN)
		}
		seen[rd.FQDN] = true
	}
	// The inserted entry landed behind the read position, so it is missed;
	// every original entry is listed exactly once.
	if got := len(result.RoutingDomains); got != 4116 {
		t.Errorf("expected 4116 routing domains, got %d", got)
	}
	if want := [][2]int{{0, 4000}, {4000, 8000}}; fmt.Sprint(requests) != fmt.Sprint(want) {
		t.Errorf("expected requests (offset, limit) %v, got %v", want, requests)
	}
}

// TestGetRoutingDomains_PageCapStopsPaging asserts a backend whose totalNum
// keeps promising more entries is cut off at routingDomainsMaxPages requests
// and fails the read, rather than paging indefinitely or returning the
// partial list as if it were complete.
func TestGetRoutingDomains_PageCapStopsPaging(t *testing.T) {
	orig := routingDomainsMaxPages
	routingDomainsMaxPages = 3
	t.Cleanup(func() { routingDomainsMaxPages = orig })

	var requests [][2]int
	client, _ := newTestClientWithResponder(routingDomainsBackend(t, 100000, &requests))

	result, err := client.GetRoutingDomains(context.Background(), 0, -1)
	if err == nil {
		t.Fatalf("expected error at page cap, got %d routing domains", len(result.RoutingDomains))
	}
	if want := "routing domains: stopped after 3 pages at offset 12000 of 100000"; err.Error() != want {
		t.Errorf("expected error %q, got %q", want, err.Error())
	}
	if want := [][2]int{{0, 4000}, {4000, 8000}, {8000, 12000}}; fmt.Sprint(requests) != fmt.Sprint(want) {
		t.Errorf("expected requests (offset, limit) %v, got %v", want, requests)
	}
}
