package core

import (
	"math"
	"testing"
	"time"

	"github.com/mapprotocol/atlas/consensus/istanbul"
)

func TestRoundChangeTimeout(t *testing.T) {
	config := *istanbul.DefaultConfig // RequestTimeout 3s, BlockPeriod 5s, backoff 1s, cap 10m
	maxTimeout := 10 * time.Minute

	for _, tt := range []struct {
		round uint64
		want  time.Duration
	}{
		{0, 8 * time.Second},         // RequestTimeout + BlockPeriod
		{1, 5 * time.Second},         // 3s + 2s
		{5, 35 * time.Second},        // 3s + 32s
		{9, 515 * time.Second},       // 3s + 512s, below the cap
		{10, maxTimeout},             // 1027s, capped
		{16, maxTimeout},             // ~18h, capped
		{34, maxTimeout},             // overflowed int64 before
		{63, maxTimeout},             // shift out of range
		{math.MaxUint64, maxTimeout}, // extreme round
	} {
		if have := roundChangeTimeout(&config, tt.round); have != tt.want {
			t.Errorf("round %d: have %v, want %v", tt.round, have, tt.want)
		}
	}

	// MaxRoundChangeTimeout = 0 keeps the unbounded exponential backoff.
	config.MaxRoundChangeTimeout = 0
	for round := uint64(1); round <= 33; round++ {
		want := 3*time.Second + time.Duration(math.Pow(2, float64(round)))*time.Second
		if have := roundChangeTimeout(&config, round); have != want {
			t.Errorf("no cap, round %d: have %v, want %v", round, have, want)
		}
	}
	for _, round := range []uint64{34, 63, 100, math.MaxUint64} {
		if have := roundChangeTimeout(&config, round); have != math.MaxInt64 {
			t.Errorf("no cap, round %d: have %v, want saturated %v", round, have, time.Duration(math.MaxInt64))
		}
	}
}

func TestCheckMaxRoundChangeTimeout(t *testing.T) {
	for _, tt := range []struct {
		v  uint64
		ok bool
	}{
		{0, true},
		{1, false},
		{29999, false},
		{30000, true},
		{600000, true},
		{istanbul.DefaultConfig.MaxRoundChangeTimeout, true},
	} {
		if err := istanbul.CheckMaxRoundChangeTimeout(tt.v); (err == nil) != tt.ok {
			t.Errorf("%d: have err %v, want ok=%v", tt.v, err, tt.ok)
		}
	}
}
