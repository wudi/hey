package jit

import (
	"sync"
	"time"
)

// HotspotDetector 热点检测器，用于识别需要JIT编译的函数
type HotspotDetector struct {
	// 编译阈值
	threshold int

	// 函数调用计数
	callCounts map[string]*FunctionCallInfo

	// 互斥锁保护数据结构
	mu sync.RWMutex

	// 清理定时器
	cleanupTicker *time.Ticker
	stopCleanup   chan bool
}

// FunctionCallInfo 函数调用信息
type FunctionCallInfo struct {
	// 调用次数
	CallCount int64

	// 首次调用时间
	FirstCallTime time.Time

	// 最后调用时间
	LastCallTime time.Time

	// 调用频率统计
	CallFrequency float64 // 每秒调用次数

	// 是否已被识别为热点
	IsHotspot bool

	// 热点识别时间
	HotspotTime time.Time
}

// NewHotspotDetector 创建新的热点检测器
func NewHotspotDetector(threshold int) *HotspotDetector {
	detector := &HotspotDetector{
		threshold:     threshold,
		callCounts:    make(map[string]*FunctionCallInfo),
		cleanupTicker: time.NewTicker(5 * time.Minute), // 每5分钟清理一次
		stopCleanup:   make(chan bool),
	}

	// 启动清理协程
	go detector.cleanupRoutine()

	return detector
}

// RecordCall 记录函数调用
func (hd *HotspotDetector) RecordCall(functionName string) {
	hd.mu.Lock()
	defer hd.mu.Unlock()

	now := time.Now()
	info, exists := hd.callCounts[functionName]

	if !exists {
		// 首次调用
		info = &FunctionCallInfo{
			CallCount:     1,
			FirstCallTime: now,
			LastCallTime:  now,
			CallFrequency: 0,
			IsHotspot:     false,
		}
		hd.callCounts[functionName] = info
	} else {
		// 更新调用信息
		info.CallCount++

		// 计算调用频率（每秒调用次数）
		duration := now.Sub(info.FirstCallTime)
		if duration > 0 {
			info.CallFrequency = float64(info.CallCount) / duration.Seconds()
		}

		info.LastCallTime = now
	}

	// 检查是否达到热点阈值（对首次调用和后续调用都检查）
	if !info.IsHotspot && info.CallCount >= int64(hd.threshold) {
		info.IsHotspot = true
		info.HotspotTime = now
	}
}

// IsHotspot 检查函数是否为热点
func (hd *HotspotDetector) IsHotspot(functionName string) bool {
	hd.mu.RLock()
	defer hd.mu.RUnlock()

	if info, exists := hd.callCounts[functionName]; exists {
		return info.IsHotspot
	}
	return false
}

// GetHotspots 获取所有热点函数
func (hd *HotspotDetector) GetHotspots() []string {
	hd.mu.RLock()
	defer hd.mu.RUnlock()

	var hotspots []string
	for funcName, info := range hd.callCounts {
		if info.IsHotspot {
			hotspots = append(hotspots, funcName)
		}
	}
	return hotspots
}

// GetFunctionInfo 获取函数调用信息
func (hd *HotspotDetector) GetFunctionInfo(functionName string) (*FunctionCallInfo, bool) {
	hd.mu.RLock()
	defer hd.mu.RUnlock()

	if info, exists := hd.callCounts[functionName]; exists {
		// 返回副本以避免并发修改
		infoCopy := *info
		return &infoCopy, true
	}
	return nil, false
}

// GetAllFunctionInfo 获取所有函数的调用信息
func (hd *HotspotDetector) GetAllFunctionInfo() map[string]FunctionCallInfo {
	hd.mu.RLock()
	defer hd.mu.RUnlock()

	result := make(map[string]FunctionCallInfo)
	for funcName, info := range hd.callCounts {
		result[funcName] = *info // 复制值以避免并发修改
	}
	return result
}

// GetTopHotspots 获取调用频率最高的N个函数
func (hd *HotspotDetector) GetTopHotspots(n int) []HotspotRank {
	hd.mu.RLock()
	defer hd.mu.RUnlock()

	// 创建排名列表
	var ranks []HotspotRank
	for funcName, info := range hd.callCounts {
		ranks = append(ranks, HotspotRank{
			FunctionName:  funcName,
			CallCount:     info.CallCount,
			CallFrequency: info.CallFrequency,
			IsHotspot:     info.IsHotspot,
		})
	}

	// 按调用次数排序（如果调用频率相同，则按调用次数排序）
	for i := 0; i < len(ranks)-1; i++ {
		for j := i + 1; j < len(ranks); j++ {
			// 优先按调用次数排序，这样更可预测
			if ranks[j].CallCount > ranks[i].CallCount {
				ranks[i], ranks[j] = ranks[j], ranks[i]
			} else if ranks[j].CallCount == ranks[i].CallCount && ranks[j].CallFrequency > ranks[i].CallFrequency {
				ranks[i], ranks[j] = ranks[j], ranks[i]
			}
		}
	}

	// 返回前N个
	if len(ranks) > n {
		ranks = ranks[:n]
	}

	return ranks
}

// HotspotRank 热点函数排名
type HotspotRank struct {
	FunctionName  string
	CallCount     int64
	CallFrequency float64
	IsHotspot     bool
}

// SetThreshold 设置新的编译阈值
func (hd *HotspotDetector) SetThreshold(threshold int) {
	hd.mu.Lock()
	defer hd.mu.Unlock()

	hd.threshold = threshold

	// 重新评估所有函数的热点状态
	for _, info := range hd.callCounts {
		if !info.IsHotspot && info.CallCount >= int64(threshold) {
			info.IsHotspot = true
			info.HotspotTime = time.Now()
		}
	}
}

// Reset 重置热点检测器
func (hd *HotspotDetector) Reset() {
	hd.mu.Lock()
	defer hd.mu.Unlock()

	hd.callCounts = make(map[string]*FunctionCallInfo)
}

// cleanupRoutine 清理过期的函数调用信息
func (hd *HotspotDetector) cleanupRoutine() {
	for {
		select {
		case <-hd.cleanupTicker.C:
			hd.cleanup()
		case <-hd.stopCleanup:
			return
		}
	}
}

// cleanup 清理长时间未调用的函数信息
func (hd *HotspotDetector) cleanup() {
	hd.mu.Lock()
	defer hd.mu.Unlock()

	now := time.Now()
	cleanupThreshold := 10 * time.Minute // 10分钟未调用则清理

	for funcName, info := range hd.callCounts {
		// 如果函数不是热点且长时间未调用，则清理
		if !info.IsHotspot && now.Sub(info.LastCallTime) > cleanupThreshold {
			delete(hd.callCounts, funcName)
		}
	}
}

// Stop 停止热点检测器
func (hd *HotspotDetector) Stop() {
	if hd.cleanupTicker != nil {
		hd.cleanupTicker.Stop()
		close(hd.stopCleanup)
	}
}

// GetStats 获取热点检测器统计信息
func (hd *HotspotDetector) GetStats() HotspotStats {
	hd.mu.RLock()
	defer hd.mu.RUnlock()

	stats := HotspotStats{
		TotalFunctions:      len(hd.callCounts),
		HotspotFunctions:    0,
		Threshold:           hd.threshold,
		TotalCalls:          0,
		AverageCallsPerFunc: 0,
	}

	var totalCalls int64
	for _, info := range hd.callCounts {
		if info.IsHotspot {
			stats.HotspotFunctions++
		}
		totalCalls += info.CallCount
	}

	stats.TotalCalls = totalCalls
	if len(hd.callCounts) > 0 {
		stats.AverageCallsPerFunc = float64(totalCalls) / float64(len(hd.callCounts))
	}

	return stats
}

// HotspotStats 热点检测器统计信息
type HotspotStats struct {
	TotalFunctions      int     // 总函数数
	HotspotFunctions    int     // 热点函数数
	Threshold           int     // 热点阈值
	TotalCalls          int64   // 总调用次数
	AverageCallsPerFunc float64 // 平均每函数调用次数
}
