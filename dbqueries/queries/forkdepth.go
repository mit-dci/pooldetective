package queries

import "time"

type ForkDepthQuery struct {
	lastRanAt time.Time
}

func (q *ForkDepthQuery) Name() string {
	return "Fork Depth"
}

func (q *ForkDepthQuery) SQL() string {
	return `BEGIN TRANSACTION;

	CREATE TABLE public.analysis_fork_depth_new
	(
		coin_id integer NOT NULL,
		fork_block_hash bytea NOT NULL,
		fork_depth integer NOT NULL,
		CONSTRAINT analysis_fork_depth_new_pkey PRIMARY KEY (coin_id, fork_block_hash),
		CONSTRAINT analysis_fork_depth_new_fkey_coin FOREIGN KEY (coin_id)
			REFERENCES public.coins (id) MATCH SIMPLE
			ON UPDATE NO ACTION
			ON DELETE NO ACTION
	);
	
	INSERT INTO public.analysis_fork_depth_new(coin_id, fork_block_hash, fork_depth)
	SELECT 
	  afb.coin_id,
	  afb.fork_block_hash,
	  (SELECT COUNT(*)/2 FROM public.analysis_competing_blocks WHERE coin_id=afb.coin_id AND height >= afb.fork_height AND height < afb.next_fork_height) as fork_depth
	FROM
	  public.analysis_fork_blocks afb;
	
	ALTER TABLE public.analysis_fork_depth
	  RENAME TO analysis_fork_depth_old;
	  
	ALTER TABLE public.analysis_fork_depth_new
	  RENAME TO analysis_fork_depth;
	  
	DROP TABLE public.analysis_fork_depth_old;
	
	ALTER TABLE public.analysis_fork_depth
	  RENAME CONSTRAINT analysis_fork_depth_new_pkey TO analysis_fork_depth_pkey;
	ALTER TABLE public.analysis_fork_depth
	  RENAME CONSTRAINT analysis_fork_depth_new_fkey_coin TO analysis_fork_depth_fkey_coin;
	
	COMMIT;`
}

func (q *ForkDepthQuery) ShouldRunAt(t time.Time) bool {
	return (t.Sub(q.lastRanAt).Minutes() > 1440)
}

func (q *ForkDepthQuery) RanAt(t time.Time) {
	q.lastRanAt = t
}
