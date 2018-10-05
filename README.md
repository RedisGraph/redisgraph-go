# redisgraph-go

`redisgraph-go` is a Golang client for the [RedisGraph](https://oss.redislabs.com/redisgraph/) module. It relies on [`redigo`](https://github.com/gomodule/redigo) for Redis connection management and provides support for RedisGraph's QUERY, EXPLAIN, and DELETE commands.

## Installation

Simply do:
```sh
$ go get github.com/redislabs/redisgraph-go
```

## Usage

```go
package main

import (
	"github.com/gomodule/redigo/redis"
	rg "github.com/redislabs/redisgraph-go"
)

func main() {
	conn, _ := redis.Dial("tcp", "0.0.0.0:6379")
	defer conn.Close()

	graph := rg.Graph{}.New("social", conn)

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
		   RETURN p.name, p.age, v.purpose, c.name`
	rs, _ := graph.Query(query)

	rs.PrettyPrint()
}
```

Running the above should output:

```sh
$ go run main.go
+----------+-----------+-----------+--------+
|  p.name  |   p.age   | v.purpose | c.name |
+----------+-----------+-----------+--------+
| John Doe | 33.000000 | NULL      | Japan  |
+----------+-----------+-----------+--------+
```

## Running tests

A simple test suite is provided, and can be run with:

```sh
$ go test
```

The tests expect a Redis server with the RedisGraph module loaded to be available at localhost:6379

## License

redisgraph-go is distributed under the BSD3 license - see [LICENSE](LICENSE)
