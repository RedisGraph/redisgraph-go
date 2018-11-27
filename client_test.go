package redisgraph

import (
	"fmt"
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

func TestFlush(t *testing.T) {
	conn, _ := redis.Dial("tcp", "0.0.0.0:6379")
	defer conn.Close()
	conn.Do("FLUSHALL")
	rg := Graph{}.New("rubbles", conn)
	users := [3]string{"Barney", "Betty", "Bam-Bam"}
	for _, user := range users {

		family := Node{
			Label: "person",
			Properties: map[string]interface{}{
				"name": fmt.Sprintf("%s Rubble", user),
			},
		}
		rg.AddNode(&family)
		rg.Flush()
	}
	query := `MATCH (p:person) RETURN p.name`
	rs, _ := rg.Query(query)
	if len(rs.Results) > 4 {
		t.Errorf("There Should only be 4 entries but we get: %d", len(rs.Results))
	}

}
