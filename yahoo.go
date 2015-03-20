// Copyright (c) 2015 Sean Kormilo. All Rights Reserved.

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

// Package yahoo defines the implementation of a stock quote driver
// to be used in package quote
package yahoo

import (
	"encoding/json"
	"github.com/sckor/quote"
	"github.com/sckor/quote/driver"
	"log"
	"net/http"
	"net/url"
	"strconv"
)

// https://query.yahooapis.com/v1/public/yql?q=select *%20from%20yahoo.finance.quotes%20where%20symbol%20in%20(%22YHOO%22%2C%22AAPL%22%2C%22GOOG%22%2C%22MSFT%22)&format=json

// Should look something like this:
// https://query.yahooapis.com/v1/public/yql?q=select * from yahoo.finance.quotes where symbol in ("MSFT", "AAPL")&format=json
// map[q:[select * from yahoo.finance.quotes where symbol in ("YHOO","AAPL","GOOG","MSFT")] format:[json] env:[store://datatables.org/alltableswithkeys]]

const endpoint = "https://query.yahooapis.com/v1/public/yql"

type yahooDriver struct{}

func init() {
	quote.Register("yahoo", &yahooDriver{})
}

// Open implements the driver required open function
func (d *yahooDriver) Open(name string) (driver.Handle, error) {
	// ignores name for now
	return &yahooHandle{h: http.DefaultClient, name: name}, nil
}

type yahooHandle struct {
	h    *http.Client
	name string
}

// Takes in a slice of ticker names and returns an appropriate query string
func yahooQueryString(tickers []string) string {
	listOfTickers := "("

	numTickers := len(tickers)

	for idx, ticker := range tickers {
		listOfTickers += "\"" + ticker + "\""
		if idx < numTickers-1 {
			listOfTickers += ","
		}
	}

	listOfTickers += ")"

	return "select * from yahoo.finance.quotes where symbol in " + listOfTickers
}

type yahooQuoteItem struct {
	Symbol             string
	LastTradePriceOnly string
}

type yahooQueryResult struct {
	Query struct {
		Count   int `json:"count"`
		Results struct {
			Quote json.RawMessage `json:"quote"`
		}
	}
}

// Implements the driver required Retrieve function for actually getting stock quote data
func (h *yahooHandle) Retrieve(tickers []string) (q []driver.StockQuote, err error) {
	baseURL, err := url.Parse(endpoint)
	if err != nil {
		return
	}

	params := url.Values{}
	params.Add("format", "json")
	params.Add("env", "store://datatables.org/alltableswithkeys")
	params.Add("q", yahooQueryString(tickers))

	baseURL.RawQuery = params.Encode()
	res, err := http.Get(baseURL.String())

	if err != nil {
		return
	}
	defer res.Body.Close()

	var m yahooQueryResult

	err = json.NewDecoder(res.Body).Decode(&m)
	if err != nil {
		return
	}

	if m.Query.Count == 0 {
		return
	}

	quotes := []yahooQuoteItem{}

	if m.Query.Count == 1 {
		var aq yahooQuoteItem
		err = json.Unmarshal(m.Query.Results.Quote, &aq)

		if err != nil {
			return
		}

		quotes = append(quotes, aq)
	} else {
		err = json.Unmarshal(m.Query.Results.Quote, &quotes)

		if err != nil {
			return
		}
	}

	for _, quote := range quotes {
		lastTradePrice, err := strconv.ParseFloat(quote.LastTradePriceOnly, 64)

		if err != nil {
			log.Printf("Failed to convert price from string to float. Source: %v err: %v", quote.LastTradePriceOnly, err)
			continue
		}

		sq := driver.StockQuote{Symbol: quote.Symbol, LastTradePrice: lastTradePrice}
		q = append(q, sq)
	}

	return
}
