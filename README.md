# Stressor

[![Go Reference](https://pkg.go.dev/badge/github.com/akramarenkov/stressor.svg)](https://pkg.go.dev/github.com/akramarenkov/stressor)
[![Go Report Card](https://goreportcard.com/badge/github.com/akramarenkov/stressor)](https://goreportcard.com/report/github.com/akramarenkov/stressor)
[![Coverage Status](https://coveralls.io/repos/github/akramarenkov/stressor/badge.svg)](https://coveralls.io/github/akramarenkov/stressor)

## Purpose

Library that provides to imposes a load on the system and the runtime in order
 to provide the main code with as little processor time as possible

This is a very simple implementation that does not adapt to performance and
 the features of the system and runtime

## Usage

Example:

```go
package main

import (
    "time"

    "github.com/akramarenkov/stressor"
)

func main() {
    opts := stressor.Opts{
        Allocators:     1,
        AllocationSize: 2,
        Lockers:        1,
        Scheduled:      1,
        SleepDuration:  10 * time.Nanosecond,
    }

    strain := stressor.New(opts)
    defer strain.Stop()

    // Main code
    time.Sleep(time.Second)
    // Output:
}
```
