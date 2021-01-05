package handler

import (
	"bytes"
	"encoding/json"
	"fmt"

	"log"
	"net/http"

	"github.com/gorilla/mux"

	builder "../builders"
	config "../configuration"
	model "../models"
	optimizer "../optimizers"
	_ "../request"
)

func getData(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	dataSource := vars["dataSource"]

	query := r.URL.Query()
	// fmt.Println("query : ", query)

	sql1 := builder.SqlBuilder(dataSource, query)
	log.Println("SQL1:", sql1)

	sql := model.SqlPost{
		Query: sql1,
	}

	reqBody, err := json.Marshal(sql)
	reqBody2 := bytes.NewBuffer(reqBody)

	//Reading conf values
	conf := config.Configure()
	serverurl := conf["serverurl"]
	recordsize := conf["recordsize"]

	response, err := http.Post(serverurl, "application/json", reqBody2)

	if err != nil {
		fmt.Print(err.Error())
	}
	var data []map[string]interface{}
	err2 := json.NewDecoder(response.Body).Decode(&data)
	if err2 != nil {
		log.Fatal(err2)
	}

	flag := optimizer.Optimize(recordsize, data)
	if flag == true {
		data = optimizer.ReExecuteQuery(sql1, conf)
	}

	json.NewEncoder(w).Encode(data)
}

//http://127.0.0.1:8081/lantern/api/v1/datasource/vrops?guestOS=Mac
