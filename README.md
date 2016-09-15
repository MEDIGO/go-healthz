# go-healthz

This package provides an HTTP handler that returns information about the health status of the application. If the application is healthy and all the registered check pass, it returns a `200 OK` HTTP status, otherwise, it fails with a `503 Service Unavailable`. All responses contain a JSON encoded payload with information about the runtime system, current checks statuses and some configurable metadata.

## Copyright and license

Copyright © 2016 MEDIGO GmbH. go-healthz is licensed under the Apache License, Version 2.0. See LICENSE for the full license text.
