package redisgraph

import (
	"fmt"
	"os"
	"testing"

	"github.com/gomodule/redigo/redis"
	"github.com/stretchr/testify/assert"
)

var graph Graph

func createGraph() {
	conn, _ := redis.Dial("tcp", os.Getenv("REDISGRAPH_URL"))
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

	assert.Equal(t, len(res.Results), 1, "expecting 1 result record")

	s, ok := (res.Results[0][0]).(*Node)
	assert.True(t, ok, "First column should contain nodes.")
	e, ok := (res.Results[0][1]).(*Edge)
	assert.True(t, ok, "Second column should contain edges.")
	d, ok := (res.Results[0][2]).(*Node)
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

	// Validate Statistics.
	assert.Equal(t, res.NodesCreated(), 1, "Expecting a single node to be created.")
	assert.Equal(t, res.PropertiesSet(), 1, "Expecting a songle property to be added.")

	q = "MATCH (w:WorkPlace) RETURN w"
	res, err = graph.Query(q)
	if err != nil {
		t.Error(err)
	}

	assert.False(t, res.Empty(), "Expecting resultset to include a single node.")
	w := (res.Results[0][0]).(*Node)
	assert.Equal(t, w.Label, "WorkPlace", "Unexpected node label.")
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

	assert.Equal(t, len(res.Results), 1, "expecting 1 result record")

	assert.Equal(t, []interface{}{0, 1, 2}, res.Results[0][0])

	q = "unwind([0,1,2]) as x return x"
	res, err = graph.Query(q)
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, len(res.Results), 3, "expecting 3 result record")

	for i := 0; i < 3; i++ {
		assert.Equal(t, i, res.Results[i][0])
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

	assert.Equal(t, 1, len(res.Results), "expecting 1 Results record")

	arr := res.Results[0][0].([]interface{})

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

func TestLinkedLookup(t *testing.T) {

	usertemplate := "CREATE (:user {uuid: '%s'})"

	res, err := graph.Query(fmt.Sprintf(usertemplate, "1"))
	assert.NoError(t, err)
	_, err = graph.Query(fmt.Sprintf(usertemplate, "2"))
	assert.NoError(t, err)
	_, err = graph.Query(fmt.Sprintf(usertemplate, "3"))
	assert.NoError(t, err)
	org := "CREATE (:organization {uuid: '15'})"
	_, err = graph.Query(org)
	assert.NoError(t, err)

	//res, err = graph.Query("MATCH (u:user) return u.uuid AS userids")
	//assert.NoError(t, err)
	//fmt.Println(fmt.Sprintf("%+v", res))

	// Group to the org
	group := "MATCH (o:organization {uuid: '15'}) CREATE (o)<-[:ispartof]-(:group {uuid: '4'})"
	_, err = graph.Query(group)
	assert.NoError(t, err)

	//res, err = graph.Query("MATCH (g:group) return g.uuid AS groupIds")
	//assert.NoError(t, err)
	//fmt.Println(fmt.Sprintf("%+v", res))

	//res, err = graph.Query("MATCH (o:organization)<-[i:ispartof]-(g:group) return o.uuid AS orgid, i AS parts, g.uuid AS groupId")
	//assert.NoError(t, err)
	//fmt.Println(fmt.Sprintf("%+v", res))

	// Add org members
	template := "MATCH (o:organization {uuid: '15'}), (u:user {uuid: '%s'}) CREATE (o)<-[:ismemberof]-(u)"
	// Add all users to orgs
	_, err = graph.Query(fmt.Sprintf(template, "1"))
	assert.NoError(t, err)
	_, err = graph.Query(fmt.Sprintf(template, "2"))
	assert.NoError(t, err)
	_, err = graph.Query(fmt.Sprintf(template, "3"))
	assert.NoError(t, err)

	//res, err = graph.Query("MATCH (o:organization)<-[m:ismemberof]-(u:user) RETURN o.uuid AS orgid, m AS membership, u.uuid AS userid")
	//assert.NoError(t, err)
	//fmt.Println(fmt.Sprintf("%+v", res))

	// make user a member of the group
	_, err = graph.Query("MATCH (g:group {uuid: '4'}), (u:user {uuid: '3'}) CREATE (g)<-[:ismemberof]-(u)")
	assert.NoError(t, err)

	//res, err = graph.Query("MATCH (g:group)<-[m:ismemberof]-(u:user) RETURN g.uuid AS groupid, m AS membership, u.uuid AS userid")
	//assert.NoError(t, err)
	//fmt.Println(fmt.Sprintf("%+v", res))

	// Make some resources
	resourcetemplate := "CREATE (:resource {uuid: '%s'})"
	_, err = graph.Query(fmt.Sprintf(resourcetemplate, "aaa"))
	assert.NoError(t, err)
	_, err = graph.Query(fmt.Sprintf(resourcetemplate, "bbb"))
	assert.NoError(t, err)
	_, err = graph.Query(fmt.Sprintf(resourcetemplate, "ccc"))
	assert.NoError(t, err)

	//res, err = graph.Query("MATCH (r:resource) RETURN r.uuid as resourceid")
	//assert.NoError(t, err)
	//fmt.Println(fmt.Sprintf("%+v", res))

	// Grant access to resources
	// user access
	res, err = graph.Query("MATCH (u:user {uuid: '1'}), (r:resource {uuid: 'aaa'}) CREATE (u)<-[:grantsaccessto {uuid: '1'}]-(r)")
	assert.NoError(t, err)
	// fmt.Println(fmt.Sprintf("%+v", res))
	res, err = graph.Query("MATCH (u:user {uuid: '3'}), (r:resource {uuid: 'aaa'}) CREATE (u)<-[:grantsaccessto {uuid: '2'}]-(r)")
	assert.NoError(t, err)
	// fmt.Println(fmt.Sprintf("%+v", res))
	res, err = graph.Query("MATCH (u:user {uuid: '1'}), (r:resource {uuid: 'ccc'}) CREATE (u)<-[:grantsaccessto {uuid: '3'}]-(r)")
	assert.NoError(t, err)
	// fmt.Println(fmt.Sprintf("%+v", res))
	res, err = graph.Query("MATCH (u:user {uuid: '2'}), (r:resource {uuid: 'ccc'}) CREATE (u)<-[:grantsaccessto {uuid: '4'}]-(r)")
	assert.NoError(t, err)
	// fmt.Println(fmt.Sprintf("%+v", res))

	// organization access
	_, err = graph.Query("MATCH (o:organization {uuid: '15'}), (r:resource {uuid: 'bbb'}) CREATE (o)<-[:grantsaccessto {uuid: '5'}]-(r)")
	assert.NoError(t, err)
	_, err = graph.Query("MATCH (o:organization {uuid: '15'}), (r:resource {uuid: 'ccc'}) CREATE (o)<-[:grantsaccessto {uuid: '6'}]-(r)")
	assert.NoError(t, err)

	// group access
	_, err = graph.Query("MATCH (g:group {uuid: '4'}), (r:resource {uuid: 'aaa'}) CREATE (g)<-[:grantsaccessto {uuid: '7'}]-(r)")
	assert.NoError(t, err)

	// Query all grants
	// Query all users directly granted access to each resource
	query := "MATCH (u:user)<-[gr:grantsaccessto]-(r:resource) RETURN u.uuid AS userid, gr.uuid AS grantid, r.uuid AS resourceid"
	res, err = graph.Query(query)
	fmt.Println(fmt.Sprintf("%+v", res))
	assert.Equal(t, 4, len(res.Results))

	// Query all grants through groups
	query = "MATCH (r:resource)-[gr:grantsaccessto]->(g:group)<-[:ismemberof]-(u:user) RETURN u.uuid AS userid, gr.uuid AS grantid, g.uuid AS groupid, r.uuid AS resourceid"
	res, err = graph.Query(query)
	fmt.Println(fmt.Sprintf("%+v", res))
	assert.Equal(t, 1, len(res.Results))
}
