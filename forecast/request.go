package forecast

import (
	"net/http"
	"strconv"

	"github.com/antonholmquist/jason"
)

type Forecast struct {
	Date     string
	Name     string
	TempHigh string
	TempLow  string
}

const (
	forecastEndpointURL = "http://weather.livedoor.com/forecast/webservice/json/v1?city="
)

func Request(code int) []Forecast {
	res, err := http.Get(forecastEndpointURL + strconv.Itoa(code))
	if err != nil {
		panic(err)
	}

	defer res.Body.Close()

	json, err := jason.NewObjectFromReader(res.Body)
	if err != nil {
		panic(err)
	}

	forecasts, err := json.GetObjectArray("forecasts")
	if err != nil {
		panic(err)
	}

	r := []Forecast{}

	for _, v := range forecasts {
		print(v)
		date, _ := v.GetString("dateLabel")
		telop, _ := v.GetString("telop")
		max, _ := v.GetString("temperature", "max", "celsius")
		min, _ := v.GetString("temperature", "min", "celsius")
		r = append(r, Forecast{date, telop, max, min})
	}

	return r
}
