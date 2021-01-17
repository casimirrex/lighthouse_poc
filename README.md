### Usage:
-       >set LANTERN=DEV
-       >go build main.go
-       >main.exe

#### Sample Request:

- http://localhost:8084/api/v1/metrics/vrops?guestOS=Linux
- http://localhost:8084/api/v1/metrics/vrops?guestOS=Linux&order.Asc=__time

- http://localhost:8084/api/v1/metrics/wikipedia?countryName={"endsWith":"e","startsWith":"Fr","operator":"and"}&column=cityName,countryName,channel&limit=5&order.Asc=__time

- //group by
   http://localhost:8084/api/v1/metrics/wikipedia?countryName=Germany&column=cityName.count,countryName,channel,__time&limit=5

- //having
  http://localhost:8084/api/v1/metrics/wikipedia?column=sum_added.max,cityName&limit=5&sum_added=115

- http://localhost:8084/api/v1/metrics/wikipedia?countryName=Germany&column=sum_added.max,cityName&sum_added={"lt":50,"gt":40,"operator":"and"}

- http://localhost:8084/api/v1/metrics/wikipedia?startTime=5Y-ago

- http://localhost:8084/api/v1/metrics/wikipedia?startTime=5Y-ago&limit=375

- http://localhost:8084/api/v1/metrics/vrops?startTime=1m-ago