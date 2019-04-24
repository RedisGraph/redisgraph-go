package redisgraph

import (
	"testing"

	"github.com/gomodule/redigo/redis"
)

func TestGraphCreation(t *testing.T) {
	// Setup.
	conn, _ := redis.Dial("tcp", "0.0.0.0:6379")
	defer conn.Close()

	conn.Do("FLUSHALL")
	rg := GraphNew("social", conn)

	// Create 2 nodes connect via a single edge.
	japan := NodeNew(0, "country", "j", nil)
	john := NodeNew(0, "person", "p", nil)
	edge := EdgeNew(0, "visited", john, japan, nil)

	// Set node properties.
	john.SetProperty("name", "John Doe")
	john.SetProperty("age", 33)
	john.SetProperty("gender", "male")
	john.SetProperty("status", "single")

	// Introduce entities to graph.
	rg.AddNode(john)
	rg.AddNode(japan)
	rg.AddEdge(edge)

	// Flush graph to DB.
	resp, err := rg.Commit()
	if err != nil {
		t.Error(err)
	}

	// Validate response.
	if(resp.results != nil) {
		t.FailNow()
	}
	if(resp.statistics["Labels added"] != 2) {
		t.FailNow()
	}
   	if(resp.statistics["Nodes created"] != 2) {
   		t.FailNow()
   	}
   	if(resp.statistics["Properties set"] != 4) {
   		t.FailNow()
   	}
   	if(resp.statistics["Relationships created"] != 1) {
   		t.FailNow()
   	}
}
