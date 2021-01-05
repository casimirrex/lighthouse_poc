### Usage:
-       >set LANTERN=DEV
-       >go build main.go
-       >main.exe

#### Sample Request:

- http://127.0.0.1:8084/lantern/api/v1/datasource/vrops?guestOS=Linux

- http://localhost:8084/api/v1/metrics/wikipedia?countryName={"endsWith":"e","startsWith":"Fr","operator":"and"}&column=cityName,countryName,channel&limit=5&order.Asc=__time

- //group by
   http://localhost:8084/api/v1/metrics/wikipedia?countryName=Germany&column=cityName.count,countryName,channel,__time&limit=5

- //having
  http://localhost:8084/api/v1/metrics/wikipedia?column=sum_added.max,cityName&limit=5&sum_added=115

- http://localhost:8084/api/v1/metrics/wikipedia?countryName=Germany&column=sum_added.max,cityName&sum_added={"lt":50,"gt":40,"operator":"and"}
