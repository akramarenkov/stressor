# Stressor

[![Go Reference](https://pkg.go.dev/badge/github.com/akramarenkov/stressor.svg)](https://pkg.go.dev/github.com/akramarenkov/stressor)
[![Go Report Card](https://goreportcard.com/badge/github.com/akramarenkov/stressor)](https://goreportcard.com/report/github.com/akramarenkov/stressor)
[![codecov](https://codecov.io/gh/akramarenkov/stressor/branch/master/graph/badge.svg?token=PqZPad4rov)](https://codecov.io/gh/akramarenkov/stressor)

## Purpose

Library that allows you to imposes a load on the system and the runtime in order to provide the main code with as little processor time as possible

This is a very simple implementation that does not adapt to performance and the features of the system and runtime

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
        AllocFactor:    1,
        AllocSize:      10,
        LockFactor:     1,
        ScheduleFactor: 1,
        ScheduleSleep:  10 * time.Nanosecond,
    }

    stressor := stressor.New(opts)
    defer stressor.Stop()

    // Main code
    time.Sleep(time.Second)

    // Output:
}
```
