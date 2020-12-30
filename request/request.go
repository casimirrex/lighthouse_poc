package request

type Request struct {
	DataSource string
	SelectColumns []string
	WhereColumns map[string]interface{}
}