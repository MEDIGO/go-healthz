# go-healthz

[![CircleCI](https://circleci.com/gh/MEDIGO/go-healthz.svg?style=shield)](https://circleci.com/gh/MEDIGO/go-healthz)

This package provides an HTTP handler that returns information about the health status of the application. If the application is healthy and all the registered check pass, it returns a `200 OK` HTTP status, otherwise, it fails with a `503 Service Unavailable`. All responses contain a JSON encoded payload with information about the runtime system, current checks statuses and some configurable metadata.

This is a fork of [github.com/MEDIGO/go-healthz](https://github.com/MEDIGO/go-healthz) with breaking changes. Changes include:

- Added the ability to return warnings from checkers that are not considered service failures.
- Renamed Set and Delete to SetMeta and DeleteMeta (this is a breaking change)
- Added Set to statically set a check status without callbacks with optional timeout. This is useful if you already have an event loop.
- Several locking fixes.


### Usage

```go
package main

import (
	"errors"
	"net/http"
	"time"

	"github.com/wojas/go-healthz"
)

const version = "1.0.0"

func main() {
	healthz.SetMeta("version", version)

	healthz.Register("important_check", time.Second*5, func() error {
		return errors.New("fail fail fail")
	})

	healthz.Register("some_warning", time.Second*5, func() error {
		return healthz.Warn("some warning")
	})

	healthz.Set("other_warning", healthz.Warn("static value"), time.Hour)

	http.Handle("/healthz", healthz.Handler())
	http.ListenAndServe(":8000", nil)
}
```

```
$ http GET localhost:8000/healthz
HTTP/1.1 503 Service Unavailable
Content-Length: 317
Content-Type: application/json
Date: Fri, 23 Sep 2016 08:55:16 GMT

{
    "ok": false,
    "has_warnings": true,
    "status": "Unavailable",
    "time": "2016-09-23T10:55:16.781538256+02:00",
    "since": "2016-09-23T10:55:14.268149643+02:00",
    "metadata": {
        "version": "1.0.0"
    },
    "failures": {
        "important_check": "fail fail fail"
    },
    "warnings": {
        "some_warning": "some warning",
        "other_warning": "static value"
    },
    "runtime": {
        "alloc_bytes": 314048,
        "arch": "amd64",
        "goroutines_count": 4,
        "heap_objects_count": 4575,
        "os": "darwin",
        "total_alloc_bytes": 314048,
        "version": "go1.19"
    }
}
```

## Copyright and license

Copyright © 2016 MEDIGO GmbH.
Copyright © 2022-present other contributors, see git history.

go-healthz is licensed under the Apache License, Version 2.0. See [LICENSE](LICENSE) for the full license text.
