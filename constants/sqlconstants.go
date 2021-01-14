package constants

const (
	Select           = "select "
	Where            = " where "
	And              = " and "
	Or               = " or "
	Asteric          = " * "
	From             = " from "
	Like             = " like "
	NotLike          = " not like "
	OrderBy          = " order by "
	Ascending        = " asc "
	Descending       = " desc "
	Limit            = " limit "
	GroupBy          = " group by "
	Having           = " having "
	Greater          = " > "
	Lesser           = " < "
	Between          = " between "
	NotEqual         = " <> "
	CurrentTimestamp = " CURRENT_TIMESTAMP "
	Interval         = " INTERVAL "
	Month            = " MONTH "
	As               = " AS "
)

// TimeIntervals contains all the values allowed for timeCheckConstants (see below)
var TimeIntervals = map[string]string{
	"m": " MINUTE ",
	"h": " HOUR ",
	"s": " SECOND ",
	"M": " MONTH ",
	"D": " DAY ",
	"Y": " YEAR ",
}

//GroupFunctions contains all the API allowed groupBy fns
var GroupFunctions = map[string]string{
	"count": "count",
	"avg":   "average",
	"sum":   "sum",
	"min":   "min",
	"max":   "max",
}

//BucketBy contains all the values allowed for time bucketing
var BucketBy = []string{"second", "minute", "hour", "day", "week", "month", "quarter", "year"}

//TimeCheckConstants contains all allowed timefilters keywords
var TimeCheckConstants = map[string]string{
	"startTime": "startTime",
	"endTime":   "endTime",
	"time.eq":   "time.eq",
	"time.ne":   "time.ne",
}
