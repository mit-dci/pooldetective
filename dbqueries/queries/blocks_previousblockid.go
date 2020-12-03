package queries

import "time"

type BlocksPreviousBlockIDQuery struct {
	lastRanAt time.Time
}

func (q *BlocksPreviousBlockIDQuery) Name() string {
	return "Previous Block ID"
}

func (q *BlocksPreviousBlockIDQuery) SQL() string {
	return `UPDATE blocks b SET previous_block_id=(SELECT id FROM blocks b2 WHERE b2.block_hash=b.previous_block_hash AND b2.coin_id=b.coin_id AND b2.height is not null LIMIT 1) WHERE b.previous_block_id IS NULL`
}

func (q *BlocksPreviousBlockIDQuery) ShouldRunAt(t time.Time) bool {
	return (t.Sub(q.lastRanAt).Minutes() > 5)
}

func (q *BlocksPreviousBlockIDQuery) RanAt(t time.Time) {
	q.lastRanAt = t
}
