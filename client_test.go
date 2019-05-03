package redisgraph

import (
	"testing"
	"os"
	"time"
	"github.com/stretchr/testify/assert"
	"github.com/gomodule/redigo/redis"
)

var graph Graph

func createGraph() {
	conn, _ := redis.Dial("tcp", "0.0.0.0:6379")
	conn.Do("FLUSHALL")
	graph = GraphNew("social", conn)

	// Create 2 nodes connect via a single edge.
	japan := NodeNew("Country", "j", nil)
	john := NodeNew("Person", "p", nil)
	edge := EdgeNew("Visited", john, japan, nil)

	// Set node properties.
	john.SetProperty("name", "John Doe")
	john.SetProperty("age", 33)
	john.SetProperty("gender", "male")
	john.SetProperty("status", "single")
	
	japan.SetProperty("name", "Japan")
	japan.SetProperty("population", 126800000)

	edge.SetProperty("year", 2017)

	// Introduce entities to graph.
	graph.AddNode(john)
	graph.AddNode(japan)
	graph.AddEdge(edge)

	// Flush graph to DB.
	_, err := graph.Commit()
	if err != nil {
		panic(err)
	}
}

func setup() {
	createGraph()
}

func shutdown() {
	graph.Conn.Close()
}

func TestMain(m *testing.M) {
    setup()
    code := m.Run() 
    shutdown()
    os.Exit(code)
}

func TestMatchQuery(t *testing.T) {
	q := "MATCH (s)-[e]->(d) RETURN s,e,d"
	res, err := graph.Query(q)
	if err != nil {
		t.Error(err)
	}
	
	assert.Equal(t, len(res.results), 1, "expecting 1 result record")

	s, ok := (res.results[0][0]).(*Node)
	assert.True(t, ok, "First column should contain nodes.")
	e, ok := (res.results[0][1]).(*Edge)
	assert.True(t, ok, "Second column should contain edges.")
	d, ok := (res.results[0][2]).(*Node)
	assert.True(t, ok, "Third column should contain nodes.")

	assert.Equal(t, s.Label, "Person", "Node should be of type 'Person'")
	assert.Equal(t, e.Relation, "Visited", "Edge should be of relation type 'Visited'")
	assert.Equal(t, d.Label, "Country", "Node should be of type 'Country'")

	assert.Equal(t, len(s.Properties), 4, "Person node should have 4 properties")

	assert.Equal(t, s.GetProperty("name"), "John Doe", "Unexpected property value.")
	assert.Equal(t, s.GetProperty("age"), 33, "Unexpected property value.")
	assert.Equal(t, s.GetProperty("gender"), "male", "Unexpected property value.")
	assert.Equal(t, s.GetProperty("status"), "single", "Unexpected property value.")
	
	assert.Equal(t, e.GetProperty("year"), 2017, "Unexpected property value.")

	assert.Equal(t, d.GetProperty("name"), "Japan", "Unexpected property value.")
	assert.Equal(t, d.GetProperty("population"), 126800000, "Unexpected property value.")
}

func TestCreateQuery(t *testing.T) {
	q := "CREATE (w:WorkPlace {name:'RedisLabs'})"
	res, err := graph.Query(q)
	if err != nil {
		t.Error(err)
	}
	
	assert.True(t, res.Empty(), "Expecting empty result-set")
	
	// Validate statistics.
	assert.Equal(t, res.NodesCreated(), 1, "Expecting a single node to be created.")
	assert.Equal(t, res.PropertiesSet(), 1, "Expecting a songle property to be added.")

	q = "MATCH (w:WorkPlace) RETURN w"
	res, err = graph.Query(q)
	if err != nil {
		t.Error(err)
	}
	
	assert.False(t, res.Empty(), "Expecting resultset to include a single node.")
	w := (res.results[0][0]).(*Node)
	assert.Equal(t, w.Label, "WorkPlace", "Unexpected node label.")
}
