package muxconn

import (
	"sync"
	"sync/atomic"
	"time"
)

type StreamStats struct {
	ID            uint64    `json:"id"`
	RX            uint64    `json:"rx"`
	TX            uint64    `json:"tx"`
	EstablishedAt time.Time `json:"established_at"`
}

type trafficStats struct {
	rx atomic.Uint64
	tx atomic.Uint64
}

func (s *trafficStats) load() (rx, tx uint64) {
	return s.rx.Load(), s.tx.Load()
}

func (s *trafficStats) incrRX(n uint64) { s.rx.Add(n) }
func (s *trafficStats) incrTX(n uint64) { s.tx.Add(n) }

type streamStats struct {
	id  uint64        // ID
	mux *trafficStats // 总线数据传输量（共享）
	stm *trafficStats // 当前子流的数据传输量
	est time.Time     // 建立连接的时间
}

func (s *streamStats) incrRX(n int) {
	if n > 0 {
		s.mux.incrRX(uint64(n))
		s.stm.incrRX(uint64(n))
	}
}

func (s *streamStats) incrTX(n int) {
	if n > 0 {
		s.mux.incrTX(uint64(n))
		s.stm.incrTX(uint64(n))
	}
}

func (s *streamStats) stats() *StreamStats {
	rx, tx := s.stm.load()

	return &StreamStats{
		ID:            s.id,
		RX:            rx,
		TX:            tx,
		EstablishedAt: s.est,
	}
}

func newMUXStreamStats() *muxStreamStats {
	return &muxStreamStats{
		mux:     new(trafficStats),
		streams: make(map[Streamer]struct{}, 8),
	}
}

type muxStreamStats struct {
	mux     *trafficStats         // 总线数据传输量（共享）
	mutex   sync.RWMutex          // streams 读写锁
	count   uint64                // 历史累计子流总数
	streams map[Streamer]struct{} // 当前活跃的子流信息
}

func (s *muxStreamStats) putConn(c Streamer) {
	stats := &streamStats{
		mux: s.mux,
		stm: new(trafficStats),
		est: time.Now(),
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.count++
	stats.id = s.count
	c.setTrunkStats(stats)
	s.streams[c] = struct{}{}
}

func (s *muxStreamStats) delConn(c Streamer) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	delete(s.streams, c)
}

func (s *muxStreamStats) actives() []Streamer {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	ret := make([]Streamer, 0, len(s.streams))
	for c := range s.streams {
		ret = append(ret, c)
	}

	return ret
}

func (s *muxStreamStats) numStreams() (int64, int64) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	cumulative, active := s.count, len(s.streams)

	return int64(cumulative), int64(active)
}
