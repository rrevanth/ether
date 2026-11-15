package stremio

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/braedonsaunders/ether/internal/core"
)

// Every add-on searches a special at the same moment. Before, only the
// first caller waited for Cinemeta's listing; the rest were told "no
// mapping", skipped the add-on and cached an empty answer.
func TestJinxConcurrentSpecialLookupsShareOneLoad(t *testing.T) {
	var fetches atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fetches.Add(1)
		time.Sleep(150 * time.Millisecond)
		w.Write([]byte(`{"meta":{"videos":[{"season":0,"episode":1,"name":"NG-Shuu 1"},{"season":0,"episode":7,"name":"Let's Chat OVA"}]}}`))
	}))
	defer srv.Close()
	m := NewSpecialMapper(srv.Client(), srv.URL)
	req := core.SearchRequest{Type: core.TypeEpisode, Title: "Kuroko's Basketball", Season: 0, Episode: 1, EpisodeTitle: "NG-Shuu 1", IDs: core.IDs{IMDB: "tt2348803"}}
	var wg sync.WaitGroup
	got := make([]int, 8)
	for i := range got {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			got[i], _ = m.LookupSpecial(context.Background(), req)
		}(i)
	}
	wg.Wait()
	for i, n := range got {
		if n != 1 {
			t.Fatalf("caller %d mapped to %d, want 1 (all: %v)", i, n, got)
		}
	}
	if f := fetches.Load(); f != 1 {
		t.Fatalf("%d Cinemeta fetches, want 1", f)
	}
	// A caller whose own context ends while waiting gives up without a mapping.
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	m2 := NewSpecialMapper(srv.Client(), srv.URL)
	if n, ok := m2.LookupSpecial(ctx, req); ok || n != 0 {
		t.Fatalf("cancelled lookup mapped to %d", n)
	}
	if n, ok := m2.LookupSpecial(context.Background(), req); !ok || n != 1 {
		t.Fatalf("lookup after a cancelled load mapped to %d %v, want 1 (cancelled load must not be remembered)", n, ok)
	}
}
