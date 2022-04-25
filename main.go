package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"reflect"
)

func main() {
	client := &http.Client{}

	req, err := http.NewRequest("GET", "https://open-api.coinglass.com/api/pro/v1/futures/funding_rates_chart?symbol=BTC&type=U", nil)

	if err != nil {
		fmt.Print(err.Error())
		os.Exit(1)
	}

	req.Header.Set("coinglassSecret", "bbef11827d2e4b9ca59f64f279216dee")

	res, err := client.Do(req)

	if err != nil {
		fmt.Print(err.Error())
		os.Exit(1)
	}

	responseData, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	handleData(responseData)
}

func handleData(data []byte) {
	type Exchanges struct {
		Bitmex  []float64 `json:"Bitmex"`
		Binance []float64 `json:"Binance"`
		Bybit   []float64 `json:"Bybit"`
		Okex    []float64 `json:"Okex"`
		Huobi   []float64 `json:"Huobi"`
		Gate    []float64 `json:"Gate"`
		FTX     []float64 `json:"FTX"`
		Bitget  []float64 `json:"Bitget"`
		DYdX    []float64 `json:"dYdX"`
	}

	type Data struct {
		DateList []int     `json:"dateList"`
		DataMap  Exchanges `json:"dataMap"`
	}

	type Response struct {
		Data Data `json:"data"`
	}

	var responseObject Response

	json.Unmarshal(data, &responseObject)

	v := reflect.ValueOf(responseObject.Data.DataMap)

	fundingRates := make(map[string][]float64)

	for i := 0; i < v.NumField(); i++ {
		fundingRates[v.Type().Field(i).Name] = v.Field(i).Interface().([]float64)
	}
}
