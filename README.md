Usage:
	>set LANTERN=DEV
	>go run main.go

Sample Request:

URLPattern: "/api/v1/metrics/{dataSource}"

http://127.0.0.1:8081/api/v1/metrics/vrops?guestOS=Linux&order.Asc=__time
http://localhost:8081/api/v1/metrics/wikipedia?countryName={"endsWith":"e","startsWith":"Fr","operator":"and"}&column=cityName,countryName,channel&limit=5&order.Asc=__time

"# lighthouse_poc" 
