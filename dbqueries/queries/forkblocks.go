package queries

import "time"

type ForkBlocksQuery struct {
	lastRanAt time.Time
}

func (q *ForkBlocksQuery) Name() string {
	return "Fork Blocks"
}

func (q *ForkBlocksQuery) SQL() string {
	return `BEGIN TRANSACTION;

	CREATE TABLE public.analysis_fork_blocks_new
	(
		coin_id integer NOT NULL,
		fork_block_hash bytea NOT NULL,
		fork_height integer NOT NULL,
		next_fork_height integer,
		CONSTRAINT analysis_fork_blocks_new_pkey PRIMARY KEY (coin_id, fork_block_hash),
		CONSTRAINT analysis_fork_blocks_new_fkey_coin FOREIGN KEY (coin_id)
			REFERENCES public.coins (id) MATCH SIMPLE
			ON UPDATE NO ACTION
			ON DELETE NO ACTION
	);
	INSERT INTO public.analysis_fork_blocks_new(coin_id, fork_block_hash, fork_height)
	SELECT coin_id, previous_block_hash, height FROM public.blocks WHERE previous_block_hash IS NOT NULL GROUP BY coin_id, previous_block_hash, height having count(*) > 1;
	
	UPDATE public.analysis_fork_blocks_new afb SET next_fork_height=COALESCE((SELECT fork_height FROM public.analysis_fork_blocks_new afb2 WHERE afb2.coin_id = afb.coin_id AND afb2.fork_height > afb.fork_height ORDER BY fork_height LIMIT 1), 999999999);
	
	ALTER TABLE public.analysis_fork_blocks
	  RENAME TO analysis_fork_blocks_old;
	  
	ALTER TABLE public.analysis_fork_blocks_new
	  RENAME TO analysis_fork_blocks;
	  
	DROP TABLE public.analysis_fork_blocks_old;
	
	ALTER TABLE public.analysis_fork_blocks
	  RENAME CONSTRAINT analysis_fork_blocks_new_pkey TO analysis_fork_blocks_pkey;
	ALTER TABLE public.analysis_fork_blocks
	  RENAME CONSTRAINT analysis_fork_blocks_new_fkey_coin TO analysis_fork_blocks_fkey_coin;
	
	COMMIT;`
}

func (q *ForkBlocksQuery) ShouldRunAt(t time.Time) bool {
	return (t.Sub(q.lastRanAt).Minutes() > 1440)
}

func (q *ForkBlocksQuery) RanAt(t time.Time) {
	q.lastRanAt = t
}
