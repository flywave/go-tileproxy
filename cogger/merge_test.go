package cogger

import (
	"context"
	"fmt"
	"io"
	"math/rand"
	"sort"
	"strings"
	"testing"

	"golang.org/x/sync/errgroup"
)

func makeTestTileKeys(n int) []TileKey {
	keys := make([]TileKey, n)
	for i := 0; i < n; i++ {
		zoom := uint8(rand.Intn(24))
		x := uint32(rand.Intn(1 << zoom))
		y := uint32(rand.Intn(1 << zoom))
		keys[i] = MakeTileKey(zoom, x, y)
	}
	return keys
}

func TestMergeTileCounts(t *testing.T) {
	sortCounts := func(counts []TileCount) {
		sort.Slice(counts, func(i, j int) bool {
			return TileCountLess(counts[i], counts[j])
		})
	}

	want := make([]TileCount, 0, 10000)

	readers := make([]io.Reader, 0, 100)
	for i := 0; i < 100; i++ {
		var buf strings.Builder
		counts := make([]TileCount, 0, 100)
		for _, tileKey := range makeTestTileKeys(rand.Intn(100)) {
			counts = append(counts, TileCount{Key: tileKey, Count: uint64(i)})
		}
		sortCounts(counts)
		for _, c := range counts {
			want = append(want, c)
			fmt.Fprintf(&buf, "%s %d\n", c.Key, c.Count)
		}
		readers = append(readers, strings.NewReader(buf.String()))
	}
	sortCounts(want)

	got, err := readMerged(readers)
	if err != nil {
		t.Fatal(err)
	}

	if len(got) != len(want) {
		t.Fatalf("got %d TileCounts, want %d", len(got), len(want))
	}

	for i := 0; i < len(got); i++ {
		if got[i] != want[i] {
			t.Fatalf("got TileCount[%d]=%v, want %v", i, got[i], want[i])
		}
	}
}

func readMerged(readers []io.Reader) ([]TileCount, error) {
	result := make([]TileCount, 0, 10000)
	ch := make(chan TileCount, 1)
	g, ctx := errgroup.WithContext(context.Background())
	g.Go(func() error {
		return mergeTileCounts(readers, ch, ctx)
	})
	g.Go(func() error {
		for t := range ch {
			result = append(result, t)
		}
		return nil
	})
	if err := g.Wait(); err != nil {
		return nil, err
	}
	return result, nil
}
