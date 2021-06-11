package redisgraph

import (
	"os"
	"testing"

	"github.com/gomodule/redigo/redis"
	"github.com/stretchr/testify/assert"
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

	checkQueryResults(t, res)
}

func TestMatchROQuery(t *testing.T) {
	q := "MATCH (s)-[e]->(d) RETURN s,e,d"
	res, err := graph.ROQuery(q)
	if err != nil {
		t.Error(err)
	}

	checkQueryResults(t, res)
}

func checkQueryResults(t *testing.T, res *QueryResult) {
	assert.Equal(t, len(res.results), 1, "expecting 1 result record")

	res.Next()
	r := res.Record()

	s, ok := r.GetByIndex(0).(*Node)
	assert.True(t, ok, "First column should contain nodes.")
	e, ok := r.GetByIndex(1).(*Edge)
	assert.True(t, ok, "Second column should contain edges.")
	d, ok := r.GetByIndex(2).(*Node)
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
	res.Next()
	r := res.Record()
	w := r.GetByIndex(0).(*Node)
	assert.Equal(t, w.Label, "WorkPlace", "Unexpected node label.")
}

func TestCreateROQueryFailure(t *testing.T) {
	q := "CREATE (w:WorkPlace {name:'RedisLabs'})"
	_, err := graph.ROQuery(q)
	assert.NotNil(t, err, "error should not be nil")
}

func TestErrorReporting(t *testing.T) {
	q := "RETURN toupper(5)"
	res, err := graph.Query(q)
	assert.Nil(t, res)
	assert.NotNil(t, err)

	q = "MATCH (p:Person) RETURN toupper(p.age)"
	res, err = graph.Query(q)
	assert.Nil(t, res)
	assert.NotNil(t, err)
}

func TestArray(t *testing.T) {
	graph.Flush()
	graph.Query("MATCH (n) DELETE n")

	q := "CREATE (:person{name:'a',age:32,array:[0,1,2]})"
	res, err := graph.Query(q)
	if err != nil {
		t.Error(err)
	}

	q = "CREATE (:person{name:'b',age:30,array:[3,4,5]})"
	res, err = graph.Query(q)
	if err != nil {
		t.Error(err)
	}

	q = "WITH [0,1,2] as x return x"
	res, err = graph.Query(q)
	if err != nil {
		t.Error(err)
	}

	res.Next()
	r := res.Record()
	assert.Equal(t, len(res.results), 1, "expecting 1 result record")
	assert.Equal(t, []interface{}{0, 1, 2}, r.GetByIndex(0))

	q = "unwind([0,1,2]) as x return x"
	res, err = graph.Query(q)
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, len(res.results), 3, "expecting 3 result record")

	i := 0
	for res.Next() {
		r = res.Record()
		assert.Equal(t, i, r.GetByIndex(0))
		i++
	}

	q = "MATCH(n) return collect(n) as x"
	res, err = graph.Query(q)
	if err != nil {
		t.Error(err)
	}

	a := NodeNew("person", "", nil)
	b := NodeNew("person", "", nil)

	a.SetProperty("name", "a")
	a.SetProperty("age", 32)
	a.SetProperty("array", []interface{}{0, 1, 2})

	b.SetProperty("name", "b")
	b.SetProperty("age", 30)
	b.SetProperty("array", []interface{}{3, 4, 5})

	assert.Equal(t, 1, len(res.results), "expecting 1 results record")

	res.Next()
	r = res.Record()
	arr := r.GetByIndex(0).([]interface{})

	assert.Equal(t, 2, len(arr))

	resA := arr[0].(*Node)
	resB := arr[1].(*Node)
	// the order of values in the array returned by collect operation is not defined
	// check for the node that contains the name "a" and set it to be resA
	if resA.GetProperty("name") != "a" {
		resA = arr[1].(*Node)
		resB = arr[0].(*Node)
	}

	assert.Equal(t, a.GetProperty("name"), resA.GetProperty("name"), "Unexpected property value.")
	assert.Equal(t, a.GetProperty("age"), resA.GetProperty("age"), "Unexpected property value.")
	assert.Equal(t, a.GetProperty("array"), resA.GetProperty("array"), "Unexpected property value.")

	assert.Equal(t, b.GetProperty("name"), resB.GetProperty("name"), "Unexpected property value.")
	assert.Equal(t, b.GetProperty("age"), resB.GetProperty("age"), "Unexpected property value.")
	assert.Equal(t, b.GetProperty("array"), resB.GetProperty("array"), "Unexpected property value.")
}

func TestMap(t *testing.T) {
	createGraph()

	q := "RETURN {val_1: 5, val_2: 'str', inner: {x: [1]}}"
	res, err := graph.Query(q)
	if err != nil {
		t.Error(err)
	}
	res.Next()
	r := res.Record()
	mapval := r.GetByIndex(0).(map[string]interface{})

	inner_map := map[string]interface{}{"x": []interface{}{1}}
	expected := map[string]interface{}{"val_1": 5, "val_2": "str", "inner": inner_map}
	assert.Equal(t, mapval, expected, "expecting a map literal")

	q = "MATCH (a:Country) RETURN a { .name }"
	res, err = graph.Query(q)
	if err != nil {
		t.Error(err)
	}
	res.Next()
	r = res.Record()
	mapval = r.GetByIndex(0).(map[string]interface{})

	expected = map[string]interface{}{"name": "Japan"}
	assert.Equal(t, mapval, expected, "expecting a map projection")
}

func TestPoint(t *testing.T) {
	createGraph()

	q := "RETURN point({latitude: -33.8567844, longitude: 151.213108})"
	res, err := graph.Query(q)
	if err != nil {
		t.Error(err)
	}
	res.Next()
	r := res.Record()
	pointval := r.GetByIndex(0).(map[string]float64)

	expected := map[string]float64{"latitude": -33.8567844, "longitude": 151.213108}
	assert.InDeltaMapValues(t, pointval, expected, 0.001, "expecting a point map")
}

func TestPath(t *testing.T) {
	createGraph()
	q := "MATCH p = (:Person)-[:Visited]->(:Country) RETURN p"
	res, err := graph.Query(q)
	if err != nil {
		t.Error(err)
	}

	assert.Equal(t, len(res.results), 1, "expecting 1 result record")

	res.Next()
	r := res.Record()

	p, ok := r.GetByIndex(0).(Path)
	assert.True(t, ok, "First column should contain path.")

	assert.Equal(t, 2, p.NodesCount(), "Path should contain two nodes")
	assert.Equal(t, 1, p.EdgeCount(), "Path should contain one edge")

	s := p.FirstNode()
	e := p.GetEdge(0)
	d := p.LastNode()

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

func TestParameterizedQuery(t *testing.T) {
	createGraph()
	params := []interface{}{1, 2.3, "str", true, false, nil, []interface {}{0, 1, 2}, []interface {}{"0", "1", "2"}}
	q := "RETURN $param"
	params_map := make(map[string]interface{})
	for index, param := range params {
		params_map["param"] = param
		res, err := graph.ParameterizedQuery(q, params_map);
		if err != nil {
			t.Error(err)
		}
		res.Next()
		assert.Equal(t, res.Record().GetByIndex(0), params[index], "Unexpected parameter value")
	}
}

func TestCreateIndex(t *testing.T) {
	res, err := graph.Query("CREATE INDEX ON :user(name)")
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, 1, res.IndicesCreated(), "Expecting 1 index created")

	res, err = graph.Query("CREATE INDEX ON :user(name)")
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, 0, res.IndicesCreated(), "Expecting 0 index created")

	res, err = graph.Query("DROP INDEX ON :user(name)")
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, 1, res.IndicesDeleted(), "Expecting 1 index deleted")

	_, err = graph.Query("DROP INDEX ON :user(name)")
	assert.Equal(t, err.Error(), "ERR Unable to drop index on :user(name): no such index.")
}

func TestQueryStatistics(t *testing.T) {
	graph.Flush()
	err := graph.Delete()
	assert.Nil(t,err)

	q := "CREATE (:Person{name:'a',age:32,array:[0,1,2]})"
	res, err := graph.Query(q)
	assert.Nil(t,err)

	assert.Equal(t, 1, res.NodesCreated(), "Expecting 1 node created")
	assert.Equal(t, 0, res.NodesDeleted(), "Expecting 0 nodes deleted")
	assert.Greater(t, res.InternalExecutionTime(),0.0, "Expecting internal execution time not to be 0.0")
	assert.Equal(t, true, res.Empty(), "Expecting empty resultset")

	res,err = graph.Query("MATCH (n) DELETE n")
	assert.Nil(t,err)
	assert.Equal(t, 1, res.NodesDeleted(), "Expecting 1 nodes deleted")

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
	res, err = graph.Commit()
	assert.Nil(t,err)
	assert.Equal(t, 2, res.NodesCreated(), "Expecting 2 node created")
	assert.Equal(t, 0, res.NodesDeleted(), "Expecting 0 nodes deleted")
	assert.Equal(t, 7, res.PropertiesSet(), "Expecting 7 properties set")
	assert.Equal(t, 1, res.RelationshipsCreated(), "Expecting 1 relationships created")
	assert.Equal(t, 0, res.RelationshipsDeleted(), "Expecting 0 relationships deleted")
	assert.Greater(t, res.InternalExecutionTime(),0.0, "Expecting internal execution time not to be 0.0")
	assert.Equal(t, true, res.Empty(), "Expecting empty resultset")
	q = "MATCH p = (:Person)-[:Visited]->(:Country) RETURN p"
	res, err = graph.Query(q)
	assert.Nil(t,err)
	assert.Equal(t, len(res.results), 1, "expecting 1 result record")
	assert.Equal(t, false, res.Empty(), "Expecting resultset to have records")
	res,err = graph.Query("MATCH ()-[r]-() DELETE r")
	assert.Nil(t,err)
	assert.Equal(t, 1, res.RelationshipsDeleted(), "Expecting 1 relationships deleted")
}

func TestUtils(t *testing.T) {
	res := RandomString(10)
	assert.Equal(t, len(res), 10)

	res = ToString("test_string")
	assert.Equal(t, res, "\"test_string\"")
	
	res = ToString(10)
	assert.Equal(t, res, "10")	

	res = ToString(1.2)
	assert.Equal(t, res, "1.2")

	res = ToString(true)
	assert.Equal(t, res, "true")

	var arr = []interface{}{1,2,3,"boom"}
	res = ToString(arr)
	assert.Equal(t, res, "[1,2,3,\"boom\"]")
	
	jsonMap := make(map[string]interface{})
	jsonMap["object"] = map[string]interface{} {"foo": 1}
	res = ToString(jsonMap)
	assert.Equal(t, res, "{object: {foo: 1}}")
}

func TestNodeMapDatatype(t *testing.T) {
	graph.Flush()
	err := graph.Delete()
	assert.Nil(t, err)

	// Create 2 nodes connect via a single edge.
	japan := NodeNew("Country", "j",
		map[string]interface{}{
			"name":       "Japan",
			"population": 126800000,
			"states":     []string{"Kanto", "Chugoku"},
		})
	john := NodeNew("Person", "p",
		map[string]interface{}{
			"name":   "John Doe",
			"age":    33,
			"gender": "male",
			"status": "single",
		})
	edge := EdgeNew("Visited", john, japan, map[string]interface{}{"year": 2017})
	// Introduce entities to graph.
	graph.AddNode(john)
	graph.AddNode(japan)
	graph.AddEdge(edge)

	// Flush graph to DB.
	res, err := graph.Commit()
	assert.Nil(t, err)
	assert.Equal(t, 2, res.NodesCreated(), "Expecting 2 node created")
	assert.Equal(t, 0, res.NodesDeleted(), "Expecting 0 nodes deleted")
	assert.Equal(t, 8, res.PropertiesSet(), "Expecting 8 properties set")
	assert.Equal(t, 1, res.RelationshipsCreated(), "Expecting 1 relationships created")
	assert.Equal(t, 0, res.RelationshipsDeleted(), "Expecting 0 relationships deleted")
	assert.Greater(t, res.InternalExecutionTime(), 0.0, "Expecting internal execution time not to be 0.0")
	assert.Equal(t, true, res.Empty(), "Expecting empty resultset")
	res, err = graph.Query("MATCH p = (:Person)-[:Visited]->(:Country) RETURN p")
	assert.Nil(t, err)
	assert.Equal(t, len(res.results), 1, "expecting 1 result record")
	assert.Equal(t, false, res.Empty(), "Expecting resultset to have records")
	res, err = graph.Query("MATCH ()-[r]-() DELETE r")
	assert.Nil(t, err)
	assert.Equal(t, 1, res.RelationshipsDeleted(), "Expecting 1 relationships deleted")
}
