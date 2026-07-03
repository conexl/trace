package commands

import "bytes"

type limitBuffer struct {
	buf       bytes.Buffer
	limit     int64
	truncated bool
}

func newLimitBuffer(limit int64) *limitBuffer {
	return &limitBuffer{limit: limit}
}

func (b *limitBuffer) Write(p []byte) (int, error) {
	if b.limit <= 0 || int64(b.buf.Len()) >= b.limit {
		b.truncated = true
		return len(p), nil
	}
	remaining := b.limit - int64(b.buf.Len())
	if int64(len(p)) > remaining {
		_, _ = b.buf.Write(p[:remaining])
		b.truncated = true
		return len(p), nil
	}
	_, _ = b.buf.Write(p)
	return len(p), nil
}

func (b *limitBuffer) String() string {
	return b.buf.String()
}
