package queries

import "time"

type PreviousBlockIDQuery struct {
	lastRanAt time.Time
}

func (q *PreviousBlockIDQuery) Name() string {
	return "Previous Block ID"
}

func (q *PreviousBlockIDQuery) SQL() string {
	return `UPDATE jobs j SET previous_block_id=COALESCE((SELECT id FROM blocks WHERE block_hash=j.previous_block_hash AND height is not null LIMIT 1),-1) WHERE previous_block_id IS NULL AND observed < now()-interval '5 minutes'`
}

func (q *PreviousBlockIDQuery) ShouldRunAt(t time.Time) bool {
	return (t.Sub(q.lastRanAt).Minutes() > 5)
}

func (q *PreviousBlockIDQuery) RanAt(t time.Time) {
	q.lastRanAt = t
}
