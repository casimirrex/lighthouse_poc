package builder

import (
	"encoding/json"
	"fmt"
	"log"
	"regexp"
	s "strings"

	sqlConstant "../constants"
)

//SQLBuilder is exported func
func SQLBuilder(dataSource string, filter map[string][]string) string {

	/* calling select clause Builder */
	selectStatement, groupByStatement, havingColumn := selectAndGroupByBuilder(filter)

	if selectStatement != "" {
		selectStatement = sqlConstant.Select + selectStatement
	} else {
		selectStatement = sqlConstant.Select + sqlConstant.Asteric
	}

	if groupByStatement != "" {
		groupByStatement = sqlConstant.GroupBy + groupByStatement
	}

	/* calling where clause Builder */
	whereCondition, havingClause := whereClauseBuilder(filter, havingColumn)
	if whereCondition != "" {
		whereCondition = sqlConstant.Where + whereCondition
	}

	if havingClause != "" {
		havingClause = sqlConstant.Having + havingClause
	}

	var orderByCondition string
	/* calling orderBy Builder */
	if filter["order.Asc"] != nil || filter["order.Desc"] != nil {
		orderByCondition = orderByBuilder(filter)
	}

	var limitCondition string
	/* if limit key in query params exists then only call limitClauseBuilder */
	if filter["limit"] != nil {
		limitCondition = limitClauseBuilder(filter)
	}

	return selectStatement + sqlConstant.From + dataSource + whereCondition + groupByStatement + havingClause + orderByCondition + limitCondition
}

func selectAndGroupByBuilder(filter map[string][]string) (string, string, string) {
	// var sqlParams model.SqlParams
	tempCols := filter["column"] // get all the columns to tempCols
	var havingColumn string
	if len(tempCols) > 0 {
		/* if the any of the column contains . operator */
		if s.Contains(tempCols[0], ".") {
			selectColumns := make([]string, 0, len(filter))    // array to store all select columns
			groupByColumns := make([]string, 0, len(filter)-1) // array to store all groupby columns
			enableGroupBy := false                             // flag to identify group function is there or not
			Cols := s.Split(tempCols[0], ",")                  // split and get all columns
			for _, col := range Cols {
				if s.Contains(col, ".") { // if . operator is found need to build group by clause
					enableGroupBy = true
					colSplit := s.Split(col, ".")  // splitting the column and group function
					groupByFunction := colSplit[1] // gives group function
					if colSplit[0] == "all" {
						colSplit[0] = sqlConstant.Asteric
					}
					actualColumn := groupByFunction + "(" + colSplit[0] + ")" //=count(column)
					havingColumn = actualColumn
					// fmt.Println("havingColumn : ", havingColumn)
					selectColumns = append(selectColumns, actualColumn+sqlConstant.As+"\""+col+"\"")
				} else {
					selectColumns = append(selectColumns, col)   // append select columns
					groupByColumns = append(groupByColumns, col) //append group by columns
				}
			}
			columns := s.Join(selectColumns[:], ",") // converting array to comma separated string
			if enableGroupBy {
				groupByColumns := s.Join(groupByColumns[:], ",") // converting array to comma separated string

				//added for bucketBy  - timebucket
				bucketClause := bucketFilter(filter)
				if bucketClause != "" {
					columns = columns + "," + bucketClause
					if groupByColumns == "" {
						groupByColumns = bucketClause
					}
					groupByColumns = groupByColumns + "," + bucketClause
				}
				return columns, groupByColumns, havingColumn
			}
			return columns, "", havingColumn
		}
		columns := tempCols[0]
		return columns, "", havingColumn
	}
	return "", "", havingColumn
}

func whereClauseBuilder(filter map[string][]string, havingColumn string) (string, string) {

	columns := make([]string, 0, len(filter))

	havingClause := ""

	for column := range filter {
		// append all the column names from the queryparams except with column, limit, order keys
		if column != "column" && column != "limit" && column != "order.Asc" && column != "order.Desc" && column != "or" && column != "and" && column != "startTime" && column != "endTime" && column != "time.eq" && column != "time.ne" && column != "by" {
			columns = append(columns, column)
		}
	}

	whereCondition := ""
	for index, column := range columns {
		values := filter[column] // get values based on column key
		if index > 0 {
			if !s.Contains(havingColumn, column) && whereCondition != "" {
				whereCondition = whereCondition + sqlConstant.And // operator need to be taken from user
			}
		}
		for j, v := range values {
			if j > 0 {
				whereCondition = whereCondition + sqlConstant.Or // operator need to be taken from user
			}
			if s.Contains(v, "{") { // if given value is json
				var dataMap map[string]interface{}
				r := s.NewReader(v)
				decodeError := json.NewDecoder(r).Decode(&dataMap) // assigning the json to dataMap
				if decodeError != nil {
					fmt.Println("decodeError : ", decodeError)
				}
				jsonKeys := make([]string, 0, len(dataMap))
				for jsonKey := range dataMap {
					jsonKeys = append(jsonKeys, jsonKey) //appending all the keys from the json like 'startsWith' etc
				}
				// fmt.Println("len(filter) : ", len(filter))
				likes := make([]string, 0, len(filter))
				havings := make([]string, 0, len(filter))
				var like string
				var having string
				for _, n := range jsonKeys {
					switch n {

					case "notEqual":
						like = column + sqlConstant.NotEqual + "'" + fmt.Sprintf("%v", dataMap[n]) + "'"
						break

					case "startsWith":
						like = column + sqlConstant.Like + "'" + fmt.Sprintf("%v", dataMap[n]) + "%'"
						break

					case "endsWith":
						like = column + sqlConstant.Like + "'%" + fmt.Sprintf("%v", dataMap[n]) + "'"
						break

					case "contains":
						like = column + sqlConstant.Like + "'%" + fmt.Sprintf("%v", dataMap[n]) + "%'"
						break

					case "doesNotStartsWith":
						like = column + sqlConstant.NotLike + "'" + fmt.Sprintf("%v", dataMap[n]) + "%'"
						break

					case "doesNotEndsWith":
						like = column + sqlConstant.NotLike + "'%" + fmt.Sprintf("%v", dataMap[n]) + "'"
						break

					case "doesNotContains":
						like = column + sqlConstant.NotLike + "'%" + fmt.Sprintf("%v", dataMap[n]) + "%'"
						break

					case "gt":
						if s.Contains(havingColumn, column) && dataMap["operator"] != nil && dataMap["operator"] != "between" {
							/* if group function is available and operator is other than between i.e., and/or */
							having = havingColumn + sqlConstant.Greater + fmt.Sprintf("%v", dataMap[n])
						} else if s.Contains(havingColumn, column) && dataMap["operator"] != nil {
							/* if group function is available and operator is between */
							having = fmt.Sprintf("%v", dataMap[n])
						} else if s.Contains(havingColumn, column) {
							/* if group function is available and no operator */
							having = havingColumn + sqlConstant.Greater + fmt.Sprintf("%v", dataMap[n])
						} else {
							/* if no group function then take as where condition */
							like = column + sqlConstant.Greater + fmt.Sprintf("%v", dataMap[n])
						}
						break

					case "lt":
						if s.Contains(havingColumn, column) && dataMap["operator"] != nil && dataMap["operator"] != "between" {
							/* if group function is available and operator is other than between i.e., and/or */
							having = havingColumn + sqlConstant.Lesser + fmt.Sprintf("%v", dataMap[n])
						} else if s.Contains(havingColumn, column) && dataMap["operator"] != nil {
							/* if group function is available and operator is between */
							having = fmt.Sprintf("%v", dataMap[n])
						} else if s.Contains(havingColumn, column) {
							/* if group function is available and no operator */
							having = havingColumn + sqlConstant.Lesser + fmt.Sprintf("%v", dataMap[n])
						} else {
							/* if no group function then take as where condition */
							like = column + sqlConstant.Lesser + fmt.Sprintf("%v", dataMap[n])
						}
						break
					}

					if n != "operator" {
						if s.Contains(havingColumn, column) {
							/* checking column is present in having column */
							havings = append(havings, having)
						} else {
							likes = append(likes, like)
							// fmt.Println("likes >> ", likes)
						}
					}
				}
				var tempLike string
				var tempHaving string
				if dataMap["operator"] != nil { // applying operators 'and/or' between given conditions
					if dataMap["operator"] == "between" {
						tempHaving = havingColumn + " " + fmt.Sprintf("%v", dataMap["operator"]) + " " + havings[0] + " " + sqlConstant.And + " " + havings[1]
					} else {
						if s.Contains(havingColumn, column) {
							tempHaving = "(" + havings[0] + " " + fmt.Sprintf("%v", dataMap["operator"]) + " " + havings[1] + ")"
						} else {
							tempLike = "(" + likes[0] + " " + fmt.Sprintf("%v", dataMap["operator"]) + " " + likes[1] + ")"
						}
					}
				} else {
					// fmt.Println("havings[0] >> ", havings[0])
					if s.Contains(havingColumn, column) {
						tempHaving = havings[0]
					} else {
						tempLike = likes[0]
					}
				}
				whereCondition = whereCondition + tempLike
				havingClause = havingClause + tempHaving
			} else { // if given value is not a json then append as equal to
				if s.Contains(havingColumn, column) {
					havingClause = havingClause + havingColumn + " = " + quoteString(v)
				} else {
					whereCondition = whereCondition + column + " = " + quoteString(v)
				}
			}
		}
	}

	// log.Println("here also")
	// log.Println("filter after others: ", filter)
	// append where clause for __time conditions
	timeClause := timeCheckFilter(filter)
	// log.Println("after time clause: ", timeClause)
	if timeClause != "" {
		if whereCondition == "" {
			whereCondition = timeClause
		} else {
			whereCondition = whereCondition + sqlConstant.And + timeClause
		}
	}
	// log.Println(">>>>>>>>>>>where condition: ", whereCondition)

	return whereCondition, havingClause
}

func orderByBuilder(filter map[string][]string) string {
	orderBy := sqlConstant.OrderBy
	if len(filter["order.Asc"]) > 0 { // if order.Asc is not empty
		orderBy = orderBy + fmt.Sprintf("%v", filter["order.Asc"][0]) + sqlConstant.Ascending
	}
	if len(filter["order.Desc"]) > 0 { // if order.Desc key is not empty
		orderBy = orderBy + fmt.Sprintf("%v", filter["order.Desc"][0]) + sqlConstant.Descending
	}

	return orderBy
}

func limitClauseBuilder(filter map[string][]string) string {

	limit := sqlConstant.Limit + fmt.Sprintf("%v", filter["limit"][0])
	return limit
}

func timeCheckFilter(filter map[string][]string) string {
	// log.Println("inside timeCheckFilter")

	startTime := filter["startTime"]
	endTime := filter["endTime"]
	timeEqual := filter["time.eq"]
	timeNotEqual := filter["time.ne"]

	// log.Println("starttime >> ", startTime)
	// log.Println("starttime >> ", s.Contains(startTime[0], "-ago"))

	timeStampLiteral := `TIMESTAMP `
	timeClause := ""

	// date should be in yyyy-MM-dd format
	// date and time should be in yyyy-MM-dd hh:mm:ss
	// we will add TIMESTAMP literal before the value
	// endTime= means all time less than and equal to endTime
	// startTime= means all time greater and equal to startTime
	// startTime= 1h-ago / 1m-ago / 1s-ago / 1D-ago / 1M-ago / 1Y-ago >> (h-hour/m-minute/s-second/D-Day/M-Month/Y-Year)
	// time= means all time equal to only time
	// time.ne= means all time not equal to time

	if len(startTime) != 0 && len(endTime) != 0 {
		timeClause = ` __time between ` + timeStampLiteral + quoteString(startTime[0]) + " and " + timeStampLiteral + quoteString(endTime[0])
	} else if len(startTime) != 0 {
		if s.Contains(startTime[0], "-ago") {
			/* regular exp to extract numeric data */
			reExpNum := regexp.MustCompile("[0-9]+")
			data := reExpNum.FindAllString(startTime[0], -1)
			/* regular exp to extract time interval i.e., h/m/s/Y/M/D */
			reExpChar := regexp.MustCompile("[a-zA-Z]+")
			char := reExpChar.FindAllString(startTime[0], -1)
			timeClause = ` __time >= ` + sqlConstant.CurrentTimestamp + `-` + sqlConstant.Interval + quoteString(data[0]) + sqlConstant.TimeIntervals[char[0]]
		} else {
			timeClause = ` __time >= ` + timeStampLiteral + quoteString(startTime[0])
		}
	} else if len(endTime) != 0 {
		timeClause = ` __time >= ` + timeStampLiteral + quoteString(endTime[0])
	} else if len(timeEqual) != 0 {
		timeClause = ` __time = ` + timeStampLiteral + quoteString(timeEqual[0])
	} else if len(timeNotEqual) != 0 {
		timeClause = ` __time <> ` + timeStampLiteral + quoteString(timeNotEqual[0])
	} else {
		return ""
	}

	return timeClause
}

func quoteString(s string) string {
	return "'" + s + "'"
}

func bucketFilter(filter map[string][]string) string {
	bucketBy := filter["by"]
	bucketByClause := ""
	// can be DAY, WEEK, MONTH, HOUR, SECOND,MINUTE, QUARTER, YEAR
	if len(bucketBy) != 0 {
		if validateBucketFilter(bucketBy[0]) {
			bucketByClause = `floor( __time to ` + bucketBy[0] + ` )`
		} else {
			log.Fatalln("Invalid parameter for timeBucket")
		}
	}
	return bucketByClause
}

func validateBucketFilter(bucketBy string) bool {
	for _, val := range sqlConstant.BucketBy {
		if bucketBy == val {
			return true
		}
	}
	return false
}
