Usage:
	>set LANTERN=DEV
	>go run main.go

Sample Request:

http://127.0.0.1:8081/lantern/api/v1/datasource/vrops?guestOS=Linux

http://localhost:8081/api/v1/metrics/wikipedia?countryName={"endsWith":"e","startsWith":"Fr","operator":"and"}&column=cityName,countryName,channel&limit=5&order.Asc=__time

//group by
http://localhost:8081/api/v1/metrics/wikipedia?countryName=Germany&column=cityName.count,countryName,channel,__time&limit=5

//having
http://localhost:8081/api/v1/metrics/wikipedia?column=sum_added.max,cityName&limit=5&sum_added=115

http://localhost:8081/api/v1/metrics/wikipedia?countryName=Germany&column=sum_added.max,cityName&sum_added={"lt":50,"gt":40,"operator":"and"}
