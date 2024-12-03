package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

type DateClose struct {
	Date  string
	Close float64
}

type Balance struct {
	Date              string
	primaryFundShares float64
	primaryFunds      float64
	hedgeFundShares   float64
	hedgeFunds        float64
	totalFunds        float64
}

var quaters bool = true

func Backtests() {

	rawData1, dateNum1, closeNum1 := getSlices("1")
	rawData2, dateNum2, closeNum2 := getSlices("2")

	primaryData := sortDataToStruct(rawData1, dateNum1, closeNum1)
	hedgeData := sortDataToStruct(rawData2, dateNum2, closeNum2)

	primaryData, hedgeData = trimSlicesToSameLength(primaryData, hedgeData)

	// need to backtest on 20 sma and quaters. Do quaters for now
	if quaters {
		quatersBacktest(primaryData, hedgeData)
	}

}

func getSlices(folder string) ([][]string, int, int) {

	files, err := os.ReadDir(*workingDIR + "//" + folder)
	if err != nil {
		panic(err)
	}

	openFile, err := os.Open(filepath.Join(*workingDIR, folder, files[0].Name()))
	if err != nil {
		panic(err)
	}

	defer openFile.Close() // defer excecuted when the function is returned meaning the file is closed when that happens

	reader := csv.NewReader(openFile)
	data, err := reader.ReadAll()
	if err != nil {
		panic(err)
	}

	var dateNum int
	var closeNum int
	for num, str := range data[0] {
		if str == "Date" {
			dateNum = num
		}
		if str == "Close" {
			closeNum = num
		}
	}

	return data, dateNum, closeNum
}

func sortDataToStruct(data [][]string, dateNum, closeNum int) []DateClose {
	returnData := make([]DateClose, len(data)-1)
	var err error
	for n := 1; n < len(data); n++ {
		returnData[n-1].Date = data[n][dateNum]
		returnData[n-1].Close, err = strconv.ParseFloat(data[n][closeNum], 64)
		if err != nil {
			panic(err)
		}
	}
	return returnData
}

func trimSlicesToSameLength(d1, d2 []DateClose) ([]DateClose, []DateClose) {
	if d1[0].Date == d2[0].Date {
		return d1, d2
	}
	date1, err := time.Parse("2006-01-02", d1[0].Date)
	if err != nil {
		panic(err)
	}
	date2, err := time.Parse("2006-01-02", d2[0].Date)
	if err != nil {
		panic(err)
	}

	if date1.Before(date2) {
		for n := 0; n < len(d1); n++ {
			if d1[n].Date == d2[0].Date {
				d1 = d1[n:]
				return d1, d2
			}
		}
	} else {
		for n := 0; n < len(d2); n++ {
			if d2[n].Date == d1[0].Date {
				d2 = d2[n:]
				return d1, d2
			}
		}
	}
	return d1, d2

}

// clean this function
// start with a total of £100. So 60/40 split will 60 primary, 40 hedge
func quatersBacktest(primaryData, hedgeData []DateClose) {
	// set primary to 0% and hedge to 100%
	primaryPercent := 0
	hedgePercent := 100

	// number of times to rebalance in a year
	numberOfTimeToRebalanceInAYear := 4

	// array to store all 20 results. Each slice will store []Balance
	var completeSetOfBackTests [21][]Balance

	// test 0-100, for both, incrementing in 5%
	counterForCompleteSetOfBacktests := 0
	for primaryPercent <= 100 && hedgePercent >= 0 {

		rebalanceMonth := findRebalanceMonths(numberOfTimeToRebalanceInAYear)
		nextQuater := firstIterationQuater(getMonth(primaryData[0].Date), rebalanceMonth)

		var rebalance bool = false

		// used for the backtest
		currentBalanceSlice := make([]Balance, len(primaryData))

		currentBalance := Balance{
			Date:              primaryData[0].Date,
			primaryFundShares: float64(primaryPercent) / primaryData[0].Close,
			primaryFunds:      float64(primaryPercent),
			hedgeFundShares:   float64(hedgePercent) / hedgeData[0].Close,
			hedgeFunds:        float64(hedgePercent),
			totalFunds:        100,
		}

		// could have issues for stocks on different exhcnages as have different trading days due to holidays. Would need to dynamically create size of array currentBalanceSlcie
		currentBalanceSlice[0] = currentBalance
		for n := 1; n < len(currentBalanceSlice); n++ {
			currentBalance := Balance{
				Date:              primaryData[n].Date,
				primaryFundShares: currentBalanceSlice[n-1].primaryFundShares,
				hedgeFundShares:   currentBalanceSlice[n-1].hedgeFundShares,
			}
			currentBalance.primaryFunds = currentBalance.primaryFundShares * primaryData[n].Close
			currentBalance.hedgeFunds = currentBalance.hedgeFundShares * hedgeData[n].Close
			currentBalance.totalFunds = currentBalance.primaryFunds + currentBalance.hedgeFunds

			// rebalancing logic
			rebalance, nextQuater = isItNewQuater(getMonth(primaryData[n].Date), nextQuater, rebalanceMonth)
			if rebalance {
				currentBalance = rebalanceFunds(currentBalance, primaryPercent, hedgePercent, primaryData[n].Close, hedgeData[n].Close)
			}

			// append results
			currentBalanceSlice[n] = currentBalance
		}

		completeSetOfBackTests[counterForCompleteSetOfBacktests] = currentBalanceSlice
		primaryPercent += 5
		hedgePercent -= 5
		counterForCompleteSetOfBacktests += 1
	}

	printBestResults(completeSetOfBackTests)

}

func getMonth(date string) int {
	parseDate, err := time.Parse("2006-01-02", date)
	if err != nil {
		fmt.Println(err)
	}
	return int(parseDate.Month())
}

func findRebalanceMonths(rebalanceFrequency int) []int {
	var month []int
	add := 12 / rebalanceFrequency
	for n := 0; n < 13; n += add {
		if n != 0 {
			if n == 12 {
				month = append(month, 1)
			} else {
				month = append(month, n+1)
			}
		}
	}
	return month
}

// there could be flaws in this logic
func firstIterationQuater(month int, rebalanceMonths []int) int {
	for n := 0; n < len(rebalanceMonths); n++ {
		if n == 0 {
			if month < rebalanceMonths[n] {
				return rebalanceMonths[n]
			}
		}
		if n == len(rebalanceMonths)-1 {
			return rebalanceMonths[n]

		}
		if rebalanceMonths[n] <= month && month < rebalanceMonths[n+1] {
			return rebalanceMonths[n+1]
		}
	}
	return rebalanceMonths[len(rebalanceMonths)-1]
}

// redo this logic
func isItNewQuater(month, nextQuater int, rebalanceMonths []int) (bool, int) {
	if month == nextQuater {
		for n := 0; n < len(rebalanceMonths); n++ {
			if n == len(rebalanceMonths)-1 {
				return true, rebalanceMonths[0]
			}
			if rebalanceMonths[n] == nextQuater {
				return true, rebalanceMonths[n+1]
			}
		}
	}
	return false, nextQuater
}

// maybe add some selling fees logic
func rebalanceFunds(currentBalance Balance, primaryPercent, hedgePercent int, primaryClose, hedgeClose float64) Balance {

	currentBalance.primaryFunds = currentBalance.totalFunds * (float64(primaryPercent) / 100)
	currentBalance.hedgeFunds = currentBalance.totalFunds * (float64(hedgePercent) / 100)

	currentBalance.primaryFundShares = currentBalance.primaryFunds / primaryClose
	currentBalance.hedgeFundShares = currentBalance.hedgeFunds / hedgeClose
	currentBalance.totalFunds = currentBalance.primaryFunds + currentBalance.hedgeFunds

	return currentBalance
}

func printBestResults(allBacktest [21][]Balance) {
	// highest returner int
	var bestRatio int
	var tFunds float64 = 0
	var currentSlice []Balance
	for n := 0; n < len(allBacktest); n++ {
		currentSlice = allBacktest[n]
		if currentSlice[len(currentSlice)-1].totalFunds > tFunds {
			bestRatio = n
			tFunds = currentSlice[len(currentSlice)-1].totalFunds
		}
	}
	bestSlice := allBacktest[bestRatio]

	fmt.Printf("Primary Shares = %.0f%% | Hedge Shares = %.0f%% | Total Returns = £%.2f", bestSlice[0].primaryFunds, bestSlice[0].hedgeFunds, bestSlice[len(bestSlice)-1].totalFunds)
}
