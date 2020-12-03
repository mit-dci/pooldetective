package queries

import "time"

type MaintenanceQuery interface {
	SQL() string
	ShouldRunAt(time.Time) bool
	RanAt(time.Time)
	Name() string
}

func AllQueries() []MaintenanceQuery {
	return []MaintenanceQuery{
		&ForkBlocksQuery{},
		&CompetingBlocksQuery{},
		&ForkDepthQuery{},
		&NextJobIDQuery{},
		&PreviousBlockIDQuery{},
		&JobTimeSpentQuery{},
		&WrongWorkDailyQuery{},
		&BlocksPreviousBlockIDQuery{},
		&EmptyBlockWorkDailyQuery{},
	}
}
