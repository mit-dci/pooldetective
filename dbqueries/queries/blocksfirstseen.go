package queries

import "time"

type BlocksFirstSeenQuery struct {
	lastRanAt time.Time
}

func (q *BlocksFirstSeenQuery) Name() string {
	return "Blocks First Seen"
}

func (q *BlocksFirstSeenQuery) SQL() string {
	return `BEGIN TRANSACTION;

	CREATE TABLE public.analysis_blocks_first_seen_new (block_id bigint primary key, first_observation timestamp without time zone);
	INSERT INTO public.analysis_blocks_first_seen_new(block_id, first_observation)
	SELECT block_id, MIN(observed) FROM public.block_observations group by block_id;
	ALTER TABLE IF EXISTS public.analysis_blocks_first_seen RENAME TO analysis_blocks_first_seen_old;
	ALTER TABLE public.analysis_blocks_first_seen_new RENAME TO analysis_blocks_first_seen;
	DROP TABLE public.analysis_blocks_first_seen_old;
	
	COMMIT;`
}

func (q *BlocksFirstSeenQuery) ShouldRunAt(t time.Time) bool {
	return (t.Sub(q.lastRanAt).Minutes() > 1400 && t.Hour() == 3) // Run around 3 AM daily
}

func (q *BlocksFirstSeenQuery) RanAt(t time.Time) {
	q.lastRanAt = t
}
