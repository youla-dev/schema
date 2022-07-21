# Example of schema usage

Code in the `main.go` demonstrates example of proto schemes and options usage.

1. Regenerate models

```bash
protoc --proto_path . --go_out=. *.proto
```

2. Run example

```bash
 go run *.go
Currency topic "weather", record "currency"
{"left":"USD","right":"EUR","value":1,"time":"2020-01-01"}
Weather topic "weather", record "weather"
{"city":"Terminus","temperature":25,"wind":5,"visibility":10000,"weather":"Sunny","time":"2020-01-01"}
```