package main

import (
	"encoding/json"
	"fmt"

	"github.com/youla-dev/schema/lib/protoschema"
)

func main() {
	topic, record, err := protoschema.ExtractTopicRecord((*Currency)(nil), E_Topic, E_Record)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Currency topic %q, record %q\n", topic, record)

	currency := Currency{
		Left:  "USD",
		Right: "EUR",
		Value: 1,
		Time:  "2020-01-01",
	}
	d, _ := json.Marshal(currency)
	fmt.Println(string(d))

	topic, record, err = protoschema.ExtractTopicRecord((*Weather)(nil), E_Topic, E_Record)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Weather topic %q, record %q\n", topic, record)

	weather := Weather{
		City:        "Terminus",
		Temperature: 25,
		Wind:        5,
		Visibility:  10000,
		Weather:     "Sunny",
		Time:        "2020-01-01",
	}
	d, _ = json.Marshal(weather)
	fmt.Println(string(d))
}
