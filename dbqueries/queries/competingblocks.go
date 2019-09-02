package queries

import "time"

type CompetingBlocksQuery struct {
	lastRanAt time.Time
}

func (q *CompetingBlocksQuery) Name() string {
	return "Competing Blocks"
}

func (q *CompetingBlocksQuery) SQL() string {
	return `BEGIN TRANSACTION;

	CREATE TABLE public.analysis_competing_blocks_new
	(
		coin_id integer NOT NULL,
		block_hash bytea NOT NULL,
		previous_block_hash bytea NOT NULL,
		height int NOT NULL,
		CONSTRAINT analysis_competing_blocks_new_pkey PRIMARY KEY (coin_id, block_hash)
	);
	INSERT INTO public.analysis_competing_blocks_new(coin_id, block_hash, previous_block_hash, height)
	SELECT 
	  b.coin_id, b.block_hash, b.previous_block_hash, b.height
	FROM 
	  public.blocks b
	  JOIN (SELECT coin_id, height FROM public.blocks WHERE height IS NOT NULL GROUP BY coin_id, height HAVING count(*) > 1) cb ON cb.coin_id=b.coin_id and cb.height=b.height;
	
	INSERT INTO public.analysis_competing_blocks_new(coin_id, block_hash, previous_block_hash, height)
	SELECT DISTINCT 
	  b.coin_id, b.block_hash, b.previous_block_hash, b.height
	FROM 
	  public.blocks b
	  JOIN (SELECT coin_id, previous_block_hash FROM public.analysis_competing_blocks_new) cb on cb.coin_id=b.coin_id and cb.previous_block_hash=b.block_hash
	  LEFT OUTER JOIN public.analysis_competing_blocks_new cbn ON cbn.block_hash=b.block_hash and cbn.coin_id=b.coin_id
	WHERE cbn.block_hash IS NULL;
	
	INSERT INTO public.analysis_competing_blocks_new(coin_id, block_hash, previous_block_hash, height)
	SELECT DISTINCT
	  b.coin_id, b.block_hash, b.previous_block_hash, b.height
	FROM 
	  public.blocks b
	  JOIN (SELECT coin_id, fork_block_hash FROM public.analysis_fork_blocks) fb on fb.coin_id=b.coin_id AND fb.fork_block_hash=b.block_hash
	  LEFT OUTER JOIN public.analysis_competing_blocks_new cbn ON cbn.block_hash=b.block_hash and cbn.coin_id=b.coin_id
	WHERE cbn.block_hash IS NULL;
	
	ALTER TABLE public.analysis_competing_blocks
	  RENAME TO analysis_competing_blocks_old;
	  
	ALTER TABLE public.analysis_competing_blocks_new
	  RENAME TO analysis_competing_blocks;
	  
	DROP TABLE public.analysis_competing_blocks_old;
	
	ALTER TABLE public.analysis_competing_blocks
	  RENAME CONSTRAINT analysis_competing_blocks_new_pkey TO analysis_competing_blocks_pkey;
	
	CREATE INDEX idx_analysis_competing_blocks1 ON public.analysis_competing_blocks(coin_id, height);
	
	COMMIT;`
}

func (q *CompetingBlocksQuery) ShouldRunAt(t time.Time) bool {
	return (t.Sub(q.lastRanAt).Minutes() > 1440)
}

func (q *CompetingBlocksQuery) RanAt(t time.Time) {
	q.lastRanAt = t
}
