# gopool

[![GoDoc](https://godoc.org/github.com/icub3d/gop/gopool?status.svg)](https://godoc.org/github.com/icub3d/gop/gopol)

Package gopool implements a concurrent work processing model. It is a
similar to thread pools in other languages, but it uses goroutines and
channels. A pool is formed wherein several goroutines get tasks from a
channel. Various sources can be used to schedule tasks and given some
coordination workgroups on various systems can work from the same
source.
