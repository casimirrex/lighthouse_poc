package optimizer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	builder "../builders"
	config "../configuration"
	model "../models"
)

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

		iCountRecords := getCount(sqlQuery, serverurl)
		log.Println("Total record count by getCount()", iCountRecords)

		if iCountRecords <= iMaxRecords {
			/*Execute query as it is and return result*/
			records = executeQuery(sqlQuery, serverurl)
		} else {
			/*Need to build and hit optimized query*/
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
	for i := 0; i < n; i++ {
		/*Adding time range or limit in query or optimizing*/
		newSQL := sqlQuery
		/*Accumulating all records in records*/
		newRecords := executeQuery(newSQL, serverURL)
		for _, item := range newRecords {
			records = append(records, item)
		}

	}
	return
}
func getCount(strSQL, serverURL string) (iCountRecords int) {
	//if order by in strSQL
	//if select * from
	//if select colnames from
	var newSQL string
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
	result := executeQuery(newSQL, serverURL)
	sCountRecords := fmt.Sprint(result[0]["EXPR$0"])
	iCountRecords, err := strconv.Atoi(sCountRecords)
	if err != nil {
		fmt.Println("Invalid recordCount in sCountRecords")
		panic(err.Error())
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
