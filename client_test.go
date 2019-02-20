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
	err := rg.AddNode(&john)
	if err != nil {
		t.Error(err)
		t.Fail()
	}

	japan := Node{
		Label: "country",
		Properties: map[string]interface{}{
			"name": "Japan",
		},
	}
	err = rg.AddNode(&japan)
	if err != nil {
		t.Error(err)
		t.Fail()
	}

	edge := Edge{
		Source:      &john,
		Relation:    "visited",
		Destination: &japan,
	}
	err = rg.AddEdge(&edge)
	if err != nil {
		t.Error(err)
		t.Fail()
	}

	_, err = rg.Commit()
	if err != nil {
		t.Error(err)
		t.Fail()
	}

	query := `MATCH (p:person)-[v:visited]->(c:country)
		   RETURN p.name, p.age, v.purpose, c.name`
	rs, err := rg.Query(query)
	if err != nil {
		t.Error(err)
		t.Fail()
	}

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
		err := rg.AddNode(&family)
		if err != nil {
			t.Error(err)
			t.Fail()
		}
		_, err = rg.Flush()
		if err != nil {
			t.Error(err)
			t.Fail()
		}
	}
	query := `MATCH (p:person) RETURN p.name`
	rs, err := rg.Query(query)
	if err != nil {
		t.Error(err)
		t.Fail()
	}
	if len(rs.Results) > 4 {
		t.Errorf("There Should only be 4 entries but we get: %d", len(rs.Results))
		t.Fail()
	}

}
