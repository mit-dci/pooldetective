package queries

import "time"

type Query struct {
	lastRanAt time.Time
}

func (q *Query) Name() string {
	return ""
}

func (q *Query) SQL() string {
	return ``
}

func (q *Query) ShouldRunAt(t time.Time) bool {
	return (t.Sub(q.lastRanAt).Minutes() > 1440)
}

func (q *Query) RanAt(t time.Time) {
	q.lastRanAt = t
}
