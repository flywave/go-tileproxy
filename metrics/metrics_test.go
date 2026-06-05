package metrics

import (
	"testing"
)

func TestGetTileKey(t *testing.T) {
	m := NewMetricsCollector().(*ServiceMetrics)

	key := m.getTileKey(1, 2, 3)
	expected := "1:2:3"
	if key != expected {
		t.Errorf("Expected key %q, got %q", expected, key)
	}
}

func TestGetTileKey_DifferentValues(t *testing.T) {
	m := NewMetricsCollector().(*ServiceMetrics)

	key1 := m.getTileKey(0, 0, 0)
	key2 := m.getTileKey(1, 0, 0)
	key3 := m.getTileKey(0, 1, 0)
	key4 := m.getTileKey(0, 0, 1)

	if key1 == key2 || key1 == key3 || key1 == key4 {
		t.Error("Different parameters should produce different keys")
	}
}

func TestRecordTileHit(t *testing.T) {
	collector := NewMetricsCollector()
	collector.RecordTileHit(1, 2, 3)
	collector.RecordTileHit(1, 2, 3)

	stats := collector.GetStats()
	tileStats, ok := stats["tile_stats"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected tile_stats map")
	}

	key := "1:2:3"
	entry, ok := tileStats[key].(map[string]interface{})
	if !ok {
		t.Fatalf("Expected entry for key %q", key)
	}

	hits, ok := entry["hits"].(int64)
	if !ok || hits != 2 {
		t.Errorf("Expected 2 hits, got %v", hits)
	}
}

func TestRecordTileMiss(t *testing.T) {
	collector := NewMetricsCollector()
	collector.RecordTileMiss(0, 0, 0)

	stats := collector.GetStats()
	key := "0:0:0"
	entry, ok := stats["tile_stats"].(map[string]interface{})[key].(map[string]interface{})
	if !ok {
		t.Fatalf("Expected entry for key %q", key)
	}

	misses, ok := entry["misses"].(int64)
	if !ok || misses != 1 {
		t.Errorf("Expected 1 miss, got %v", misses)
	}
}

func TestRecordTileError(t *testing.T) {
	collector := NewMetricsCollector()
	collector.RecordTileError(1, 2, 3, nil)

	stats := collector.GetStats()
	key := "1:2:3"
	entry, ok := stats["tile_stats"].(map[string]interface{})[key].(map[string]interface{})
	if !ok {
		t.Fatalf("Expected entry for key %q", key)
	}

	errs, ok := entry["errors"].(int64)
	if !ok || errs != 1 {
		t.Errorf("Expected 1 error, got %v", errs)
	}
}

func TestStatsReset(t *testing.T) {
	collector := NewMetricsCollector()
	collector.RecordTileHit(1, 2, 3)
	collector.Reset()

	stats := collector.GetStats()
	tileStats := stats["tile_stats"].(map[string]interface{})
	if len(tileStats) != 0 {
		t.Error("Expected empty tile stats after reset")
	}
	if stats["request_count"].(int64) != 0 {
		t.Error("Expected zero request count after reset")
	}
}
