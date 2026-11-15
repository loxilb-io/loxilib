![sanity workflow](https://github.com/loxilb-io/loxilib/actions/workflows/basic-sanity.yml/badge.svg)  [![Go Report Card](https://goreportcard.com/badge/github.com/loxilb-io/loxilib?style=flat)](https://goreportcard.com/report/github.com/loxilb-io/loxilib)


## What is loxilib

loxilib is a standalone golang package providing utility routines like lpm, networking utils, counters, logging etc. Originally designed to work only with [loxilb](https://github.com/loxilb-io/loxilb), it can now be used with any golang project. It exports functionality not found usually in other standard golang packages.

## Concurrent logging

`LogItInit` initializes a file logger and installs it if there is no default.
Use `NewLogger` to construct a logger for an `io.Writer` and `SwapDefaultLogger`
to replace the default. The swap publishes the new logger before waiting for
calls using the previous publication to finish. Close the previous output only
after any other uses have finished as well.

Use `LogItSetLevel`, `LogItGetLevel`, and `LogItSetTTY` during concurrent use.
Direct access to `DefaultLogger`, `CurrLogLevel`, or `LogTTY` requires exclusive
access. Do not call `SwapDefaultLogger` from a log writer or formatter. Such a
call would wait for the log operation that invoked it.
