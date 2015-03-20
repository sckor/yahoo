# Yahoo
A golang based yahoo finance query driver to be used with the quote go lang package

The quote golang package can be found here: https://github.com/sckor/quote

# Usage
Use go get to download the source

```
go get github.com/sckor/yahoo
```

You can then run the example command:
```
go run src/github.com/sckor/yahoo/yahoo-quote/yahoo-quote.go
```

Output:
```
[{MSFT 42.86} {AAPL 128.0851}]
```

By default it will fetch quotes for MSFT and AAPL. If you run the command and provide a list of tickers after the command, it will 
fetch the prices for those tickers. eg:

```
go run src/github.com/sckor/yahoo/yahoo-quote/yahoo-quote.go TSLA GOOG
```

Output:
```
[{TSLA 198.34} {GOOG 560.01}]
```


