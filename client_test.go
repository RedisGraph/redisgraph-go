package redisgraph

import (
	"testing"

	"github.com/gomodule/redigo/redis"
)

func TestExample(t *testing.T) {
	conn, _ := redis.Dial("tcp", "0.0.0.0:6379")
	defer conn.Close()

	conn.Do("FLUSHALL")
	rg := Graph{}.New("social", conn)

	john := Node{
		Label: "person",
		Properties: map[string]interface{}{
			"name":   "John Doe",
			"age":    33,
			"gender": "male",
			"status": "single",
		},
	}
	rg.AddNode(&john)

	japan := Node{
		Label: "country",
		Properties: map[string]interface{}{
			"name": "Japan",
		},
	}
	rg.AddNode(&japan)

	edge := Edge{
		Source:      &john,
		Relation:    "visited",
		Destination: &japan,
	}
	rg.AddEdge(&edge)

	rg.Commit()

	query := `MATCH (p:person)-[v:visited]->(c:country)
		   RETURN p.name, p.age, v.purpose, c.name`
	rs, _ := rg.Query(query)

	rs.PrettyPrint()
}
