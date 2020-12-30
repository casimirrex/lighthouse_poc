package model

type SqlPost struct {
	Query string `json:"query"`
}

type SqlParams struct {
	Columns        string
	GroupByColumns string
}
