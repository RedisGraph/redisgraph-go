[![license](https://img.shields.io/github/license/RedisGraph/redisgraph-go.svg)](https://github.com/RedisGraph/redisgraph-go)
[![CircleCI](https://circleci.com/gh/RedisGraph/redisgraph-go/tree/master.svg?style=svg)](https://circleci.com/gh/RedisGraph/redisgraph-go/tree/master)
[![GitHub issues](https://img.shields.io/github/release/RedisGraph/redisgraph-go.svg)](https://github.com/RedisGraph/redisgraph-go/releases/latest)
[![Codecov](https://codecov.io/gh/RedisGraph/redisgraph-go/branch/master/graph/badge.svg)](https://codecov.io/gh/RedisGraph/redisgraph-go)
[![Go Report Card](https://goreportcard.com/badge/github.com/RedisGraph/redisgraph-go)](https://goreportcard.com/report/github.com/RedisGraph/redisgraph-go)
[![GoDoc](https://godoc.org/github.com/RedisGraph/redisgraph-go?status.svg)](https://godoc.org/github.com/RedisGraph/redisgraph-go)

# redisgraph-go
[![Forum](https://img.shields.io/badge/Forum-RedisGraph-blue)](https://forum.redislabs.com/c/modules/redisgraph)
[![Gitter](https://badges.gitter.im/RedisLabs/RedisGraph.svg)](https://gitter.im/RedisLabs/RedisGraph?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge)

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
	"os"

	"github.com/gomodule/redigo/redis"
	rg "github.com/redislabs/redisgraph-go"
)

func main() {
	conn, _ := redis.Dial("tcp", "127.0.0.1:6379")
	defer conn.Close()

	graph := rg.GraphNew("social", conn)

	graph.Delete()

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
	fmt.Println("Visited countries by person:")
	for result.Next() { // Next returns true until the iterator is depleted.
		// Get the current Record.
		r := result.Record()

		// Entries in the Record can be accessed by index or key.
		pName := r.GetByIndex(0)
		fmt.Printf("\nName: %s\n", pName)
		pAge, _ := r.Get("p.age")
		fmt.Printf("\nAge: %d\n", pAge)
	}

	// Path matching example.
	query = "MATCH p = (:person)-[:visited]->(:country) RETURN p"
	result, err := graph.Query(query)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Println("Pathes of persons visiting countries:")
	for result.Next() {
		r := result.Record()
		p, ok := r.GetByIndex(0).(rg.Path)
		fmt.Printf("%s %v\n", p, ok)
	}
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
