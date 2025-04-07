# simultaneous - limit the number of concurrent operations, Go

[![GoDoc](https://godoc.org/github.com/singlestore-labs/simultaneous?status.svg)](https://pkg.go.dev/github.com/singlestore-labs/simultaneous)
![unit tests](https://github.com/singlestore-labs/simultaneous/actions/workflows/go.yml/badge.svg)
[![report card](https://goreportcard.com/badge/github.com/singlestore-labs/simultaneous)](https://goreportcard.com/report/github.com/singlestore-labs/simultaneous)
[![codecov](https://codecov.io/gh/singlestore-labs/simultaneous/branch/main/graph/badge.svg)](https://codecov.io/gh/singlestore-labs/simultaneous)

Install:

	go get github.com/singlestore-labs/simultaneous

---

Simultaneous limits the number of concurrent operations. It supports passing proof of 
obtaining a limit.

## Simple usage

```go
var limit = simultaneous.New[any](10)

func unlimitedWait() {
	defer limit.Forever()()
	// do stuff
}

func limitedWait() error {
	done, err := limit.Timeout(time.Minute)
	if err != nil { 
		return fmt.Errorf("timeout: %w", err)
	}
	defer done()

	// do stuff
}
```

## Proving that operating within a limit

You can prove that you've got permission


```go
type myLimitType struct{}

var limit = simultaneous.New[myLimitType](10)

func wantsProof(_ Enforced[myLimitType]) {
	// do something
}

func providesProof() {
	done := limit.Forever()
	defer done()
	wantsProof(done)
}
```
