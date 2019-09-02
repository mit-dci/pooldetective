package queries

import "time"

type JobsFirstSeenQuery struct {
	lastRanAt time.Time
}

func (q *JobsFirstSeenQuery) Name() string {
	return ""
}

func (q *JobsFirstSeenQuery) SQL() string {
	return `BEGIN TRANSACTION;
	CREATE TABLE public.analysis_jobs_first_seen_new (previous_block_id bigint primary key, first_observation timestamp without time zone);
	INSERT INTO public.analysis_jobs_first_seen_new(previous_block_id, first_observation)
	SELECT previous_block_id, MIN(observed) FROM jobs WHERE previous_block_id IS NOT NULL group by previous_block_id;
	
	ALTER TABLE IF EXISTS public.analysis_jobs_first_seen RENAME TO analysis_jobs_first_seen_old;
	ALTER TABLE public.analysis_jobs_first_seen_new RENAME TO analysis_jobs_first_seen;
	DROP TABLE public.analysis_jobs_first_seen_old;
	COMMIT;`
}

func (q *JobsFirstSeenQuery) ShouldRunAt(t time.Time) bool {
	return (t.Sub(q.lastRanAt).Minutes() > 1400 && t.Hour() == 4) // Run around 4 AM daily
}

func (q *JobsFirstSeenQuery) RanAt(t time.Time) {
	q.lastRanAt = t
}
