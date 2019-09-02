package queries

import "time"

type WrongWorkDailyQuery struct {
	lastRanAt time.Time
}

func (q *WrongWorkDailyQuery) Name() string {
	return "Wrong work daily"
}

func (q *WrongWorkDailyQuery) SQL() string {
	return `BEGIN TRANSACTION;
	INSERT INTO analysis_wrong_work_daily(observed_on, pool_id, pool_observer_id, location_id, expected_coin_id, got_coin_id, total_jobs, total_time_msec, wrong_jobs, wrong_time_msec)
	SELECT
		j.observed::date as observed_on,
		p.id as pool_id,
		po.id as pool_observer_id,
		l.id as location_id,
		c.id as expected_coin_id,
		c2.id as got_coin_id,
		tj.totaljobs as total_jobs,
		tj.totaltime as total_time_msec,
		count(*) as wrong_jobs,
		sum(time_spent_msec) as wrong_time_msec
	FROM 
		jobs j
		left join pool_observers po on po.id=j.pool_observer_id
		left join pools p on p.id=po.pool_id
		left join coins c on c.id=po.coin_id
		left join locations l on l.id=po.location_id
		LEFT JOIN blocks b on b.id=j.previous_block_id
		left join coins c2 on c2.id=b.coin_id
		LEFT JOIN (SELECT COUNT(*) as totaljobs, SUM(time_spent_msec) as totaltime, pool_observer_id, observed::date FROM jobs WHERE observed::date > COALESCE((SELECT max(observed_on) FROM analysis_wrong_work_daily), '2019-01-01'::date) AND observed < NOW()::date GROUP BY pool_observer_id, observed::date) tj on tj.pool_observer_id=j.pool_observer_id and tj.observed=j.observed::date
	WHERE l.id=2 AND j.observed::date > COALESCE((SELECT max(observed_on) FROM analysis_wrong_work_daily), '2019-01-01'::date) AND j.observed < NOW()::date AND c2.name IS NOT NULL AND c.id != c2.id
	GROUP BY 
	p.id ,
		po.id ,
		l.id,
		c.id,
		c2.id,
		j.observed::date,
		tj.totaljobs,
		tj.totaltime;
	COMMIT;`
}

func (q *WrongWorkDailyQuery) ShouldRunAt(t time.Time) bool {
	return (t.Sub(q.lastRanAt).Minutes() > 1400 && t.Hour() == 2) // Run around 2 AM daily
}

func (q *WrongWorkDailyQuery) RanAt(t time.Time) {
	q.lastRanAt = t
}
