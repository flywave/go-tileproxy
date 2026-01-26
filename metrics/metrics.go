package metrics

import (
	"sync"
	"time"
)

type MetricsCollector interface {
	RecordTileHit(service, layer, level int)
	RecordTileMiss(service, layer, level int)
	RecordTileError(service, layer, level int, err error)
	RecordTaskProgress(taskId string, completed, total int)
	GetStats() map[string]interface{}
	Reset()
}

type TileStats struct {
	Hits   int64
	Misses int64
	Errors int64
}

type ServiceMetrics struct {
	mu            sync.RWMutex
	tileStats     map[string]*TileStats
	taskProgress  map[string]*TaskProgressMetrics
	requestCount  int64
	errorCount    int64
	startTime     time.Time
	lastResetTime time.Time
}

type TaskProgressMetrics struct {
	Completed int
	Total     int
	StartTime time.Time
	UpdatedAt time.Time
}

func NewMetricsCollector() MetricsCollector {
	return &ServiceMetrics{
		tileStats:     make(map[string]*TileStats),
		taskProgress:  make(map[string]*TaskProgressMetrics),
		startTime:     time.Now(),
		lastResetTime: time.Now(),
	}
}

func (m *ServiceMetrics) getTileKey(service, layer, level int) string {
	return "s:l:z"
}

func (m *ServiceMetrics) getTileStats(service, layer, level int) *TileStats {
	key := m.getTileKey(service, layer, level)
	m.mu.RLock()
	stats, ok := m.tileStats[key]
	m.mu.RUnlock()

	if !ok {
		m.mu.Lock()
		stats, ok = m.tileStats[key]
		if !ok {
			stats = &TileStats{}
			m.tileStats[key] = stats
		}
		m.mu.Unlock()
	}

	return stats
}

func (m *ServiceMetrics) RecordTileHit(service, layer, level int) {
	stats := m.getTileStats(service, layer, level)
	m.mu.Lock()
	stats.Hits++
	m.requestCount++
	m.mu.Unlock()
}

func (m *ServiceMetrics) RecordTileMiss(service, layer, level int) {
	stats := m.getTileStats(service, layer, level)
	m.mu.Lock()
	stats.Misses++
	m.requestCount++
	m.mu.Unlock()
}

func (m *ServiceMetrics) RecordTileError(service, layer, level int, err error) {
	stats := m.getTileStats(service, layer, level)
	m.mu.Lock()
	stats.Errors++
	m.errorCount++
	m.mu.Unlock()
}

func (m *ServiceMetrics) RecordTaskProgress(taskId string, completed, total int) {
	m.mu.Lock()
	defer m.mu.Unlock()

	progress, ok := m.taskProgress[taskId]
	if !ok {
		progress = &TaskProgressMetrics{
			StartTime: time.Now(),
		}
		m.taskProgress[taskId] = progress
	}

	progress.Completed = completed
	progress.Total = total
	progress.UpdatedAt = time.Now()
}

func (m *ServiceMetrics) GetStats() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	stats := map[string]interface{}{
		"request_count":   m.requestCount,
		"error_count":     m.errorCount,
		"start_time":      m.startTime,
		"last_reset_time": m.lastResetTime,
		"uptime_seconds":  time.Since(m.startTime).Seconds(),
		"tile_stats":      m.getTileStatsMap(),
		"task_progress":   m.getTaskProgressMap(),
	}

	return stats
}

func (m *ServiceMetrics) getTileStatsMap() map[string]interface{} {
	tileStatsMap := make(map[string]interface{})
	for key, stats := range m.tileStats {
		tileStatsMap[key] = map[string]interface{}{
			"hits":   stats.Hits,
			"misses": stats.Misses,
			"errors": stats.Errors,
		}
	}
	return tileStatsMap
}

func (m *ServiceMetrics) getTaskProgressMap() map[string]interface{} {
	taskProgressMap := make(map[string]interface{})
	for taskId, progress := range m.taskProgress {
		taskProgressMap[taskId] = map[string]interface{}{
			"completed":   progress.Completed,
			"total":       progress.Total,
			"progress":    float64(progress.Completed) / float64(progress.Total),
			"start_time":  progress.StartTime,
			"updated_at":  progress.UpdatedAt,
			"duration_ms": time.Since(progress.StartTime).Milliseconds(),
		}
	}
	return taskProgressMap
}

func (m *ServiceMetrics) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.tileStats = make(map[string]*TileStats)
	m.taskProgress = make(map[string]*TaskProgressMetrics)
	m.requestCount = 0
	m.errorCount = 0
	m.lastResetTime = time.Now()
}

var DefaultCollector MetricsCollector = NewMetricsCollector()

func GetDefaultCollector() MetricsCollector {
	return DefaultCollector
}

func RecordTileHit(service, layer, level int) {
	DefaultCollector.RecordTileHit(service, layer, level)
}

func RecordTileMiss(service, layer, level int) {
	DefaultCollector.RecordTileMiss(service, layer, level)
}

func RecordTileError(service, layer, level int, err error) {
	DefaultCollector.RecordTileError(service, layer, level, err)
}

func RecordTaskProgress(taskId string, completed, total int) {
	DefaultCollector.RecordTaskProgress(taskId, completed, total)
}

func GetStats() map[string]interface{} {
	return DefaultCollector.GetStats()
}

func ResetMetrics() {
	DefaultCollector.Reset()
}
