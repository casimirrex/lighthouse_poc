package optimizer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	builder "../builders"
	config "../configuration"
	model "../models"
)

var durMap = map[string]string{
	"'10' MINUTE": "10min",
	"'30' MINUTE": "30min",
	"'6' HOUR":    "6H",
	"'1' DAY":     "1D",
	"'15' DAY":    "15D",
	"'1' MONTH":   "1M",
	"'6' MONTH":   "6M",
	"'1' YEAR":    "1Y",
	"'5' YEAR":    "5Y",
}

//Optimize is ...
func Optimize(dataSource string, filter map[string][]string) (records []map[string]interface{}) {
	/*We assume that startTime and endTime must be available in filter (query parameters)*/
	var sTime, eTime string
	if len(filter["startTime"]) > 0 {
		sTime = filter["startTime"][0]
	}
	if len(filter["endTime"]) > 0 {
		eTime = filter["endTime"][0]
	}
	log.Printf("sTime %s, eTime %s\n", sTime, eTime)
	/*To check whether requested data is from before 24 hours*/
	timeDiff := diffTimeFromNow(sTime, "NOW")
	log.Println("time.Now().UTC():", time.Now().UTC())
	log.Println("diffTimeFromNow.Hours():", timeDiff.Hours())
	if timeDiff.Hours() <= 24 { //bypassed check using negation '!'
		log.Println("time range: <=24h")
		//Requested data from Cache/Reids
		//....
		//....
	} else {
		log.Println("time range: >24h")
		//Requested data from Druid
		timeRange := diffTimestamps(sTime, eTime)
		if timeRange.Minutes() > 5 {
			log.Println("Time rage for requested data is greater than 5 minutes, so need to fetch data in buckets")
		}

		/*Reading config values*/
		configuration := config.Configure()
		serverurl := configuration["serverurl"]
		sMaxRecords := configuration["recordsize"]

		/*Evaluates the resultSet*/
		iMaxRecords, err := strconv.Atoi(sMaxRecords)
		if err != nil {
			fmt.Println("Invalid record_count in config")
			panic(err.Error())
		}
		log.Println("Max record count limit is", iMaxRecords)

		sqlQuery := builder.SQLBuilder(dataSource, filter)
		log.Println("SQLBuilder:", sqlQuery)

		iCountRecords, sampleDuration := getCount(sqlQuery, serverurl)
		log.Println("Total record count by getCount() sampleDuration", iCountRecords, sampleDuration)
		//Assume how many records using iCountRecords, sampleDuration
		// expectedRecords := (timeRange/sampleDuration)*iCountRecords
		// if expectedRecords > iMaxRecords {
		// 		[{ "message" : "We have %s <- expectedRecords, do you want to get all?"" }]
		// } else {
		//		records = executeQuery(sqlQuery, serverurl)
		// }

		if iCountRecords <= iMaxRecords {
			/*Execute query as it is and return result*/
			log.Println("Case1************************************")
			records = executeQuery(sqlQuery, serverurl)
		} else {
			/*Need to build and hit optimized query*/
			log.Println("Case2************************************")
			posLimit := strings.Index(sqlQuery, " limit")
			if posLimit > 0 {
				r, _ := regexp.Compile(" limit ([0-9]+)")
				s := r.FindStringSubmatch(sqlQuery)
				iCountRecords, _ = strconv.Atoi(s[1])
				sqlQuery = sqlQuery[:posLimit]
			}
			records = executeBucketQueries(sqlQuery, serverurl, iMaxRecords, iCountRecords)
		}
	}
	return
}

func executeBucketQueries(sqlQuery, serverURL string, maxRecords, countRecords int) (records []map[string]interface{}) {

	/*Will update as per scenario, for now LIMIT is used*/
	//Re-write query here either from sqlQuery or Optimize's filter
	var m, n int
	m = countRecords % maxRecords
	n = countRecords / maxRecords
	if m > 0 {
		n = n + 1
	}
	log.Println("numBuckets:", n)
	//n = 5 //Restricted to 5 for debug
	limit, offset := 100, 0
	if maxRecords < limit {
		limit = maxRecords
	}
	for i := 0; i < n; i++ {
		/*Adding time range or limit in query or optimizing*/
		newSQL := fmt.Sprintf("%s limit %d offset %d", sqlQuery, limit, offset)
		offset = offset + limit
		remainder := countRecords - offset
		if remainder < limit {
			limit = remainder
		}
		/*Accumulating all records in records*/
		newRecords := executeQuery(newSQL, serverURL)
		for _, item := range newRecords {
			records = append(records, item)
		}

	}
	return
}
func getCount(strSQL, serverURL string) (iCountRecords int, duration string) {
	//if order by in strSQL
	//if select * from
	//if select colnames from
	var newSQL string
	var strLimitOffset string
	posLimitOffset := strings.Index(strSQL, " limit")
	if posLimitOffset > 0 {
		strLimitOffset = strSQL[posLimitOffset:]
		strSQL = strSQL[:posLimitOffset]
	}
	if strings.Contains(strSQL, "order by") {
		//newSQL = "select count(1) from ( " + strSQL + " )"
		index := strings.Index(strSQL, "order by")
		strSQL = strSQL[:index]
		newSQL = strings.Replace(strSQL, "*", "count(1)", 1)
	} else if strings.Contains(strSQL, "group by") {
		newSQL = strSQL
	} else if strings.Contains(strSQL, "*") {
		newSQL = strings.Replace(strSQL, "*", "count(1)", 1)
	} else if !strings.Contains(strSQL, "*") {
		index := strings.Index(strSQL, "from")
		strSQL = strSQL[index:]
		newSQL = "select count(1) " + strSQL
	} else {
		newSQL = strSQL
	}
	timeArr := []string{
		"'10' MINUTE",
		"'30' MINUTE",
		"'6' HOUR",
		"'1' DAY",
		"'15' DAY",
		"'1' MONTH",
		"'6' MONTH",
		"'1' YEAR",
		"'5' YEAR",
	}
	posTime := strings.Index(newSQL, " __time")
	if posTime > 0 {
		newSQL = newSQL[:posTime]
	}
	posLimit := strings.Index(newSQL, " limit")
	if posLimit > 0 {
		newSQL = newSQL[:posLimit]
	}
	posWhere := strings.Index(newSQL, " where")
	if posWhere < 0 {
		newSQL = newSQL + " where"
	}

	for _, el := range timeArr {
		//Replace timeClause by __time >= CURRENT_TIMESTAMP - INTERVAL '10' MINUTE ...
		newSQLt := newSQL + " __time >= CURRENT_TIMESTAMP - INTERVAL " + el + strLimitOffset
		result := executeQuery(newSQLt, serverURL)
		log.Println("getCount:Query executed...")
		sCountRecords := "0"
		if len(result) > 0 {
			sCountRecords = fmt.Sprint(result[0]["EXPR$0"])
			fmt.Println("getCount=s>", sCountRecords, el)
		}
		nCountRecords, err := strconv.Atoi(sCountRecords)
		if err != nil {
			fmt.Println("Invalid recordCount in sCountRecords")
			panic(err.Error())
		}
		if nCountRecords > 0 {
			iCountRecords = nCountRecords
			duration = durMap[el]
			break
		}
	}
	return
}
func executeQuery(sqlQuery, serverURL string) (result []map[string]interface{}) {
	/*Preparing request body*/
	sqlObject := model.SqlPost{
		Query: sqlQuery,
	}
	reqBody, errMarshal := json.Marshal(sqlObject)
	if errMarshal != nil {
		fmt.Print(errMarshal.Error())
	}
	reqBodyBuf := bytes.NewBuffer(reqBody)

	/*Querying database over http*/
	log.Println("http.Post(...)", reqBodyBuf)
	response, errHTTP := http.Post(serverURL, "application/json", reqBodyBuf)
	if errHTTP != nil {
		fmt.Print(errHTTP.Error())
	}
	/*Decoding results*/
	errDecode := json.NewDecoder(response.Body).Decode(&result)
	if errDecode != nil {
		log.Fatal(errDecode)
	}
	log.Println("Record counts from executeQuery()", len(result))
	return
}
func diffTimestamps(startTime, endTime string) time.Duration {
	//diffTimestamps("2019-01-24T00:00:00.000Z", "2019-01-26T00:30:20.000Z")
	start, _ := time.Parse(time.RFC3339, startTime)
	end, _ := time.Parse(time.RFC3339, endTime)
	return end.Sub(start)
}

func diffTimeFromNow(startTime, endTime string) time.Duration {
	//diffTimestamps("2019-01-24T00:00:00.000Z", "2019-01-26T00:30:20.000Z")
	start, _ := time.Parse(time.RFC3339, startTime)
	if (strings.ToUpper(endTime) == "NOW") || (strings.ToUpper(endTime) == "CURRENT_TIMESTAMP") {
		//match all possibilities like => CURRENT_TIMESTAMP - INTERVAL '1' DAY
		// 1y ago, 2m ago, 1w ago, 2w ago, 3d ago, 6h ago, 30min ago
		fmt.Println("")
	}
	end := time.Now().UTC()
	return end.Sub(start)
}
