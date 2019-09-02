package queries

import "time"

type NextJobIDQuery struct {
	lastRanAt time.Time
}

func (q *NextJobIDQuery) Name() string {
	return "Next Job ID"
}

func (q *NextJobIDQuery) SQL() string {
	return `BEGIN TRANSACTION;
	UPDATE jobs j SET next_job_id=(select j2.id from jobs j2 where j2.observed > j.observed and j2.pool_observer_id=j.pool_observer_id order by observed asc limit 1) WHERE next_job_id IS NULL;
	COMMIT;`
}

func (q *NextJobIDQuery) ShouldRunAt(t time.Time) bool {
	return (t.Sub(q.lastRanAt).Minutes() > 5)
}

func (q *NextJobIDQuery) RanAt(t time.Time) {
	q.lastRanAt = t
}
