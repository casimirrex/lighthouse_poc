package handler

import (
	"encoding/json"
	"net/http"

	optimizer "../optimizers"
	"github.com/gorilla/mux"
)

func getData(w http.ResponseWriter, r *http.Request) {

	pathParams := mux.Vars(r)
	queryParams := r.URL.Query()
	dataSource := pathParams["dataSource"]

	data := optimizer.Optimize(dataSource, queryParams)
	json.NewEncoder(w).Encode(data)
}
