package optimizer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	model "../models"
)

//Optimize is ...
func Optimize(sMaxRecords string, data []map[string]interface{}) bool {
	/*Evaluates the resultSet*/
	iMaxRecords, err := strconv.Atoi(sMaxRecords)
	if err != nil {
		fmt.Println("Invalid record_count in config")
		panic(err)
	}
	log.Println("Allowed record size is", iMaxRecords)
	iCountRecords := len(data)
	log.Println("Actual record size is", iCountRecords)
	flag := false
	if iCountRecords > iMaxRecords {
		/*Need to hit optimized query*/
		flag = true
	}
	return flag
}

//ReExecuteQuery is ...
func ReExecuteQuery(query string, conf map[string]string) (result []map[string]interface{}) {
	iMaxRecords, err := strconv.Atoi(conf["recordsize"])

	/*Will update as per scenario, for now LIMIT is used*/
	query = fmt.Sprintf("%s LIMIT %d", query, iMaxRecords)
	log.Println("OPTIMIZED:", query)

	sql := model.SqlPost{
		Query: query,
	}

	reqBody, err := json.Marshal(sql)
	reqBody2 := bytes.NewBuffer(reqBody)

	//Reading conf values
	serverurl := conf["serverurl"]

	response, err := http.Post(serverurl, "application/json", reqBody2)

	if err != nil {
		fmt.Print(err.Error())
	}
	//var data []map[string]interface{}
	err2 := json.NewDecoder(response.Body).Decode(&result)
	if err2 != nil {
		log.Fatal(err2)
	}
	log.Println("Final record size after optimiz", len(result))
	return
}
