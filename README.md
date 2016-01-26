# Parallel

[![Build Status](https://travis-ci.org/krhubert/parallel.png)](https://travis-ci.org/krhubert/parallel)
[![Coverage Status](https://coveralls.io/repos/github/krhubert/parallel/badge.svg)](https://coveralls.io/github/krhubert/parallel)
[![Go Report Card](http://goreportcard.com/badge/krhubert/parallel)](http://goreportcard.com/report/krhubert/parallel)
[![GitHub license](https://img.shields.io/github/license/mashape/apistatus.svg)](https://opensource.org/licenses/MIT)
[![GoDoc](https://godoc.org/github.com/krhubert/parallel?status.svg)](https://godoc.org/github.com/krhubert/parallel)


Package parallel implements running golang functions in parallel.

```Go
import "github.com/krhubert/parallel"
```

## Usage

```Go
package main

import (
  "fmt"
  "log"
  "math/rand"
  "time"

  "github.com/krhubert/parallel"
)

func fn(n int) int {
  // processing
  time.Sleep(time.Second * time.Duration(rand.Intn(n)))
  // output
  return n * n
}

func main() {
  task, err := parallel.NewTask(5, fn)
  if err != nil {
    log.Fatal(err)
  }

  for n := 1; n <= 10; n++ {
    task.Feed(n)
  }

  res, err := task.Run()
  if err != nil {
    log.Fatal(err)
  }

  for i := range res {
    fmt.Println(res[i])
  }
}
```

License
-------

MIT License
