package core

import (
	"crypto/rand"
	"encoding/binary"
	"sync"
	"time"
)

// ULID generation (ADR-027): a 128-bit identifier — 48-bit big-endian millisecond timestamp
// followed by 80 bits of randomness — encoded as 26 Crockford base32 characters. Generation
// is monotonic per process: within the same millisecond the random component is incremented
// so identifiers remain lexicographically sortable in creation order.

const crockford = "0123456789ABCDEFGHJKMNPQRSTVWXYZ"

var ulidGen = &generator{}

type generator struct {
	mu       sync.Mutex
	lastMS   uint64
	lastRand [10]byte
}

// NewULID returns a new monotonic ULID for the current time.
func NewULID() ULID {
	return ulidGen.next(uint64(time.Now().UnixMilli()))
}

func (g *generator) next(ms uint64) ULID {
	g.mu.Lock()
	defer g.mu.Unlock()

	var entropy [10]byte
	if ms == g.lastMS {
		// Same millisecond: increment the previous randomness (monotonicity).
		entropy = g.lastRand
		for i := 9; i >= 0; i-- {
			entropy[i]++
			if entropy[i] != 0 {
				break
			}
		}
	} else {
		if _, err := rand.Read(entropy[:]); err != nil {
			// crypto/rand failure is fatal for identifier integrity; fall back to a
			// time-derived value rather than emitting a colliding zero entropy.
			binary.BigEndian.PutUint64(entropy[:8], ms*2654435761)
		}
		g.lastMS = ms
	}
	g.lastRand = entropy

	var b [16]byte
	b[0] = byte((ms >> 40) & 0xFF)
	b[1] = byte((ms >> 32) & 0xFF)
	b[2] = byte((ms >> 24) & 0xFF)
	b[3] = byte((ms >> 16) & 0xFF)
	b[4] = byte((ms >> 8) & 0xFF)
	b[5] = byte(ms & 0xFF)
	copy(b[6:], entropy[:])
	return encode(b)
}

// encode renders 16 bytes as 26 Crockford base32 characters (ULID layout).
func encode(b [16]byte) ULID {
	out := make([]byte, 26)
	out[0] = crockford[(b[0]&224)>>5]
	out[1] = crockford[b[0]&31]
	out[2] = crockford[(b[1]&248)>>3]
	out[3] = crockford[((b[1]&7)<<2)|((b[2]&192)>>6)]
	out[4] = crockford[(b[2]&62)>>1]
	out[5] = crockford[((b[2]&1)<<4)|((b[3]&240)>>4)]
	out[6] = crockford[((b[3]&15)<<1)|((b[4]&128)>>7)]
	out[7] = crockford[(b[4]&124)>>2]
	out[8] = crockford[((b[4]&3)<<3)|((b[5]&224)>>5)]
	out[9] = crockford[b[5]&31]
	out[10] = crockford[(b[6]&248)>>3]
	out[11] = crockford[((b[6]&7)<<2)|((b[7]&192)>>6)]
	out[12] = crockford[(b[7]&62)>>1]
	out[13] = crockford[((b[7]&1)<<4)|((b[8]&240)>>4)]
	out[14] = crockford[((b[8]&15)<<1)|((b[9]&128)>>7)]
	out[15] = crockford[(b[9]&124)>>2]
	out[16] = crockford[((b[9]&3)<<3)|((b[10]&224)>>5)]
	out[17] = crockford[b[10]&31]
	out[18] = crockford[(b[11]&248)>>3]
	out[19] = crockford[((b[11]&7)<<2)|((b[12]&192)>>6)]
	out[20] = crockford[(b[12]&62)>>1]
	out[21] = crockford[((b[12]&1)<<4)|((b[13]&240)>>4)]
	out[22] = crockford[((b[13]&15)<<1)|((b[14]&128)>>7)]
	out[23] = crockford[(b[14]&124)>>2]
	out[24] = crockford[((b[14]&3)<<3)|((b[15]&224)>>5)]
	out[25] = crockford[b[15]&31]
	return string(out)
}
