package cube

import "time"

const patternByteChunkShift = 20
const patternByteChunkSize = 1 << patternByteChunkShift

// Separate backing arrays bound heap clearing as well as explicit fill work.
// A single gigabyte allocation can spend hundreds of milliseconds clearing
// reused pages inside make, where the caller cannot check its deadline.
type patternByteChunks [][]uint8

func (d patternByteChunks) at(x uint32) uint8 {
	return d[x>>patternByteChunkShift][x&(patternByteChunkSize-1)]
}

func (d patternByteChunks) byte(x uint32) *uint8 {
	return &d[x>>patternByteChunkShift][x&(patternByteChunkSize-1)]
}

func (d patternByteChunks) size() int {
	if len(d) == 0 {
		return 0
	}
	return (len(d)-1)*patternByteChunkSize + len(d[len(d)-1])
}

func filledPatternByteChunks(size int, value uint8, deadline time.Time) patternByteChunks {
	if size <= 0 || tableDeadlineExceeded(deadline) {
		return nil
	}
	d := make(patternByteChunks, (size+patternByteChunkSize-1)/patternByteChunkSize)
	for i := range d {
		if tableDeadlineExceeded(deadline) {
			return nil
		}
		chunk := make([]uint8, min(patternByteChunkSize, size-i*patternByteChunkSize))
		for offset := 0; offset < len(chunk); {
			if tableDeadlineExceeded(deadline) {
				return nil
			}
			end := min(offset+(64<<10), len(chunk))
			for j := offset; j < end; j++ {
				chunk[j] = value
			}
			offset = end
		}
		d[i] = chunk
	}
	if tableDeadlineExceeded(deadline) {
		return nil
	}
	return d
}
