package muxconn

import (
	"context"
	"errors"
	"io"
	"math"
	"os"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

const (
	defaultLimitWait = 10 * time.Second

	// 每个 Muxer 可以产生 N 个 stream，这些 stream 共享同一个限流器，
	// maximumChunk 为了每个 stream 每次读/写申请最大的 token 数，为了
	// 避免多 stream 竞争同一个限流器，特此对 minimumBurst 做基础限制，
	// minimumBurst = maximumChunk * 10，这个 10 倍只是经验值。
	//
	// 为什么 maximumChunk = 32K？
	// 因为 io.Copy 默认 32K，这个值小一点也可以，但是可能会让每次写入
	// 的 syscall 变得频繁，该值也是经验值。
	maximumChunk = 2 << 14
	minimumBurst = maximumChunk * 10
)

func newUnlimit() *rateLimiter {
	return &rateLimiter{
		rlimit: rate.NewLimiter(rate.Inf, minimumBurst),
		wlimit: rate.NewLimiter(rate.Inf, minimumBurst),
	}
}

type rateLimiter struct {
	rlimit *rate.Limiter
	wlimit *rate.Limiter
	mutex  sync.Mutex
}

func (rl *rateLimiter) Limit() rate.Limit { return rl.rlimit.Limit() }

func (rl *rateLimiter) SetLimit(bps rate.Limit) bool {
	if bps < minimumBurst {
		return false
	}

	now := time.Now()
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	rl.rlimit.SetLimitAt(now, bps)
	rl.wlimit.SetLimitAt(now, bps)

	if bps != rate.Inf {
		burst := int(math.Ceil(float64(bps) * 1.2))
		burst = max(burst, minimumBurst)
		rl.rlimit.SetBurstAt(now, burst)
		rl.wlimit.SetBurstAt(now, burst)
	}

	return true
}

func (rl *rateLimiter) warpReadWriter(parent context.Context, rw io.ReadWriter) io.ReadWriter {
	if parent == nil {
		parent = context.Background()
	}

	return &limitReadWriter{
		rlimit: rl.rlimit,
		wlimit: rl.wlimit,
		under:  rw,
		parent: parent,
	}
}

type limitReadWriter struct {
	rlimit *rate.Limiter
	wlimit *rate.Limiter
	under  io.ReadWriter
	parent context.Context
}

func (lrw *limitReadWriter) Read(p []byte) (int, error) {
	if len(p) == 0 || lrw.rlimit.Limit() == rate.Inf {
		return lrw.under.Read(p)
	}

	tokens := min(len(p), maximumChunk)
	if err := lrw.waitN(lrw.rlimit, tokens); err != nil {
		return 0, err
	}

	return lrw.under.Read(p[:tokens])
}

func (lrw *limitReadWriter) Write(p []byte) (int, error) {
	total := len(p)
	if total == 0 || lrw.wlimit.Limit() == rate.Inf {
		return lrw.under.Write(p)
	}

	remain := total
	var written int
	for remain > 0 {
		tokens := min(remain, maximumChunk)
		if err := lrw.waitN(lrw.wlimit, tokens); err != nil {
			return written, err
		}

		n, err := lrw.under.Write(p[written : written+tokens])
		if n > 0 {
			written += n
		}
		if err != nil {
			return written, err
		}

		remain = total - written
	}

	return written, nil
}

func (lrw *limitReadWriter) waitN(limit *rate.Limiter, n int) error {
	ctx, cancel := context.WithTimeout(lrw.parent, defaultLimitWait)
	defer cancel()

	if err := limit.WaitN(ctx, n); err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return os.ErrDeadlineExceeded
		}

		return err
	}

	return nil
}
