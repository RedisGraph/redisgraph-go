[![license](https://img.shields.io/github/license/RedisGraph/redisgraph-go.svg)](https://github.com/RedisGraph/redisgraph-go)
[![CircleCI](https://circleci.com/gh/RedisGraph/redisgraph-go/tree/master.svg?style=svg)](https://circleci.com/gh/RedisGraph/redisgraph-go/tree/master)
[![GitHub issues](https://img.shields.io/github/release/RedisGraph/redisgraph-go.svg)](https://github.com/RedisGraph/redisgraph-go/releases/latest)
[![Codecov](https://codecov.io/gh/RedisGraph/redisgraph-go/branch/master/graph/badge.svg)](https://codecov.io/gh/RedisGraph/redisgraph-go)
[![Go Report Card](https://goreportcard.com/badge/github.com/RedisGraph/redisgraph-go)](https://goreportcard.com/report/github.com/RedisGraph/redisgraph-go)
[![GoDoc](https://godoc.org/github.com/RedisGraph/redisgraph-go?status.svg)](https://godoc.org/github.com/RedisGraph/redisgraph-go)

[![Mailing List](https://img.shields.io/badge/Mailing%20List-RedisGraph-blue)](https://groups.google.com/forum/#!forum/redisgraph)
[![Gitter](https://badges.gitter.im/RedisLabs/RedisGraph.svg)](https://gitter.im/RedisLabs/RedisGraph?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge)

# redisgraph-go

`redisgraph-go` is a Golang client for the [RedisGraph](https://oss.redislabs.com/redisgraph/) module. It relies on [`redigo`](https://github.com/gomodule/redigo) for Redis connection management and provides support for RedisGraph's QUERY, EXPLAIN, and DELETE commands.

## Installation

Simply do:
```sh
$ go get github.com/redislabs/redisgraph-go
```

## Usage

The complete `redisgraph-go` API is documented on [GoDoc](https://godoc.org/github.com/RedisGraph/redisgraph-go).

```go
package main

import (
    "fmt"
    "github.com/gomodule/redigo/redis"
    rg "github.com/redislabs/redisgraph-go"
)

func main() {
    conn, _ := redis.Dial("tcp", "0.0.0.0:6379")
    defer conn.Close()

    graph := rg.GraphNew("social", conn)

    john := rg.Node{
        Label: "person",
        Properties: map[string]interface{}{
            "name":   "John Doe",
            "age":    33,
            "gender": "male",
            "status": "single",
        },
    }
    graph.AddNode(&john)

    japan := rg.Node{
        Label: "country",
        Properties: map[string]interface{}{
            "name": "Japan",
        },
    }
    graph.AddNode(&japan)

    edge := rg.Edge{
        Source:      &john,
        Relation:    "visited",
        Destination: &japan,
    }
    graph.AddEdge(&edge)

    graph.Commit()

    query := `MATCH (p:person)-[v:visited]->(c:country)
           RETURN p.name, p.age, c.name`

    // result is a QueryResult struct containing the query's generated records and statistics.
    result, _ := graph.Query(query)

    // Pretty-print the full result set as a table.
    result.PrettyPrint()

    // Iterate over each individual Record in the result.
    for result.Next() { // Next returns true until the iterator is depleted.
        // Get the current Record.
        r := result.Record()

        // Entries in the Record can be accessed by index or key.
        p_name := r.GetByIndex(0)
        fmt.Printf("\nName: %s\n", p_name)
        p_age, _ := r.Get("p.age")
        fmt.Printf("\nAge: %d\n", p_age)
    }

    // Path matching example.
    query = "MATCH p = (:Person)-[:Visited]->(:Country) RETURN p"
    result, _ = graph.Query(query)
    res.Next()
    r := res.Record()
    p, ok := r.GetByIndex(0).(Path)
    fmt.Printf("%s", p)
}
```

Running the above produces the output:

```sh
+----------+-------+--------+
|  p.name  | p.age | c.name |
+----------+-------+--------+
| John Doe |    33 | Japan  |
+----------+-------+--------+

Query internal execution time 1.623063

Name: John Doe

Age: 33
```

## Running tests

A simple test suite is provided, and can be run with:

```sh
$ go test
```

The tests expect a Redis server with the RedisGraph module loaded to be available at localhost:6379

## License

redisgraph-go is distributed under the BSD3 license - see [LICENSE](LICENSE)
