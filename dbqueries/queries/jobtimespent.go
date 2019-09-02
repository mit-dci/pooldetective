package queries

import "time"

type JobTimeSpentQuery struct {
	lastRanAt time.Time
}

func (q *JobTimeSpentQuery) Name() string {
	return "Job Time Spent"
}

func (q *JobTimeSpentQuery) SQL() string {
	return `BEGIN TRANSACTION;
	UPDATE jobs j SET time_spent_msec=least(600000::bigint, (extract(epoch from j2.observed-j.observed))*1000)::bigint FROM jobs j2 WHERE j.next_job_id=j2.id AND j.time_spent_msec IS NULL AND j.next_job_id IS NOT NULL;
	COMMIT;`

	/*	UPDATE jobs j SET time_spent_msec=least(600000::bigint, (extract(epoch from ((select j2.observed from jobs j2 where (j2.id=j.next_job_id))-j.observed))*1000)::bigint) WHERE time_spent_msec IS NULL AND next_job_id IS NOT NULL;
	 */
}

func (q *JobTimeSpentQuery) ShouldRunAt(t time.Time) bool {
	return (t.Sub(q.lastRanAt).Minutes() > 5)
}

func (q *JobTimeSpentQuery) RanAt(t time.Time) {
	q.lastRanAt = t
}
