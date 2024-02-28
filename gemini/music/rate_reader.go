package music

import (
	"io"

	"github.com/juju/ratelimit"
)

type rate_reader struct {
	r      io.Reader
	bucket *ratelimit.Bucket
}

func RateReader(r io.Reader, bucket *ratelimit.Bucket) io.Reader {
	return &rate_reader{
		r:      r,
		bucket: bucket,
	}
}

func (r *rate_reader) Read(buf []byte) (int, error) {
	n, err := r.r.Read(buf)
	if n <= 0 {
		return n, err
	}
	r.bucket.Wait(int64(n))
	return n, err
}
