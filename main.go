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
	cFundingRatesSTABLE := make(chan map[string][]float64)
	cFundingRatesCOIN := make(chan map[string][]float64)
	cDatesSTABLE := make(chan []int)
	cDatesCOIN := make(chan []int)

	go requestData("https://open-api.coinglass.com/api/pro/v1/futures/funding_rates_chart?symbol=BTC&type=U", cFundingRatesSTABLE, cDatesSTABLE)
	go requestData("https://open-api.coinglass.com/api/pro/v1/futures/funding_rates_chart?symbol=BTC&type=C", cFundingRatesCOIN, cDatesCOIN)

	// program will wait until all these variables receive a channel's data
	fundingRatesSTABLE := <-cFundingRatesSTABLE
	fundingRatesCOIN := <-cFundingRatesCOIN
	// the dates are separated by 8h between
	datesSTABLE := <-cDatesSTABLE
	datesCOIN := <-cDatesCOIN

	// if the first date of each API fetch is different (maybe because of the time between the two fetches), it means that it will me all up, so we recall the function main
	if datesSTABLE[0] != datesCOIN[0] || len(datesSTABLE) != len(datesCOIN) {
		main()
	}

	fundingRates := make(map[string][]float64)

	// we merge the two funding rates together
	for key, value := range fundingRatesSTABLE {
		for i := 0; i < len(value); i++ {
			fundingRates[key] = append(fundingRates[key], (fundingRatesSTABLE[key][i]+fundingRatesCOIN[key][i])/2)
		}
	}

	average := computeAverage(fundingRates, datesSTABLE)
	fmt.Println(average)
}

func requestData(url string, cFundingRates chan map[string][]float64, cDates chan []int) {
	client := &http.Client{}

	req, err := http.NewRequest("GET", url, nil)

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

	handleData(responseData, cFundingRates, cDates)
}

func handleData(data []byte, cFundingRates chan map[string][]float64, cDates chan []int) {
	type Exchanges struct {
		Bitmex  []float64 `json:"Bitmex"`
		Binance []float64 `json:"Binance"`
		Bybit   []float64 `json:"Bybit"`
		Okex    []float64 `json:"Okex"`
		Huobi   []float64 `json:"Huobi"`
		Gate    []float64 `json:"Gate"`
		// FTX     []float64 `json:"FTX"`
		// Bitget  []float64 `json:"Bitget"`
		// DYdX    []float64 `json:"dYdX"`
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

	// return e.g. ["Binance": [0.001, 0.01, 0.0045], "FTx": []]. And for date : [91659, 1961981, 16198, 654654]
	cFundingRates <- fundingRates
	cDates <- responseObject.Data.DateList

	close(cFundingRates)
	close(cDates)
}

func computeAverage(fundingRates map[string][]float64, dates []int) map[int]float64 {
	// dateRange := time.Duration(int64((dates[len(dates)-1] - dates[0]) * 1000000))

	// everytime a value is negative, we capture the range

	average := make(map[int][]float64)

	// instead of having the name of an exchange as a key, we have the date. E.g: [1651680000000: [-0.32, 0.12], 1651708800000: []]
	for _, value := range fundingRates {
		for i := 0; i < len(value); i++ {
			average[dates[i]] = append(average[dates[i]], value[i])
		}
	}

	averageComputed := make(map[int]float64)

	for key, value := range average {
		averageComputed[key] = averageArray(value)
	}

	return averageComputed
}

func averageArray(array []float64) float64 {
	var total float64
	for j := 0; j < len(array); j++ {
		total += array[j]
	}
	return total / float64(len(array))
}
