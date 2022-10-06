package redisgraph

import (
	"fmt"
	"strings"
	"sync"

	"github.com/gomodule/redigo/redis"
)

// QueryMode is the query mode
type QueryMode string

const (
	// ReadQuery is a read query
	ReadQuery QueryMode = "READ"
	// WriteQuery is a write query
	WriteQuery QueryMode = "WRITE"
)

// QueryOptions are a set of additional arguments to be emitted with a query.
type QueryOptions struct {
	timeout int
}

// Graph represents a graph, which is a collection of nodes and edges.
type Graph struct {
	Id                string
	Nodes             map[string]*Node
	Edges             []*Edge
	Conn              redis.Conn
	labels            []string   // List of node labels.
	relationshipTypes []string   // List of relation types.
	properties        []string   // List of properties.
	mutex             sync.Mutex // Lock, used for updating internal state.
}

// New creates a new graph.
func GraphNew(Id string, conn redis.Conn) Graph {
	return Graph{
		Id:                Id,
		Nodes:             make(map[string]*Node, 0),
		Edges:             make([]*Edge, 0),
		Conn:              conn,
		labels:            make([]string, 0),
		relationshipTypes: make([]string, 0),
		properties:        make([]string, 0),
	}
}

// AddNode adds a node to the graph.
func (g *Graph) AddNode(n *Node) {
	if n.Alias == "" {
		n.Alias = RandomString(10)
	}
	n.graph = g
	g.Nodes[n.Alias] = n
}

// AddEdge adds an edge to the graph.
func (g *Graph) AddEdge(e *Edge) error {
	// Verify that the edge has source and destination
	if e.Source == nil || e.Destination == nil {
		return fmt.Errorf("Both source and destination nodes should be defined")
	}

	// Verify that the edge's nodes have been previously added to the graph
	if _, ok := g.Nodes[e.Source.Alias]; !ok {
		return fmt.Errorf("Source node neeeds to be added to the graph first")
	}
	if _, ok := g.Nodes[e.Destination.Alias]; !ok {
		return fmt.Errorf("Destination node neeeds to be added to the graph first")
	}

	e.graph = g
	g.Edges = append(g.Edges, e)
	return nil
}

// ExecutionPlan gets the execution plan for given query.
func (g *Graph) ExecutionPlan(q string) (string, error) {
	return redis.String(g.Conn.Do("GRAPH.EXPLAIN", g.Id, q))
}

// Delete removes the graph.
func (g *Graph) Delete() error {
	_, err := g.Conn.Do("GRAPH.DELETE", g.Id)

	// clear internal mappings
	g.labels = g.labels[:0]
	g.properties = g.properties[:0]
	g.relationshipTypes = g.relationshipTypes[:0]

	return err
}

// Flush will create the graph and clear it
func (g *Graph) Flush() (*QueryResult, error) {
	res, err := g.Commit()
	if err == nil {
		g.Nodes = make(map[string]*Node)
		g.Edges = make([]*Edge, 0)
	}
	return res, err
}

// Commit creates the entire graph, but will re-add nodes if called again.
func (g *Graph) Commit() (*QueryResult, error) {
	items := make([]string, 0, len(g.Nodes)+len(g.Edges))
	for _, n := range g.Nodes {
		items = append(items, n.Encode())
	}
	for _, e := range g.Edges {
		items = append(items, e.Encode())
	}
	q := "CREATE " + strings.Join(items, ",")
	return g.Query(q)
}

// NewQueryOptions instantiates a new QueryOptions struct.
func NewQueryOptions() *QueryOptions {
	return &QueryOptions{
		timeout: -1,
	}
}

// SetTimeout sets the timeout member of the QueryOptions struct
func (options *QueryOptions) SetTimeout(timeout int) *QueryOptions {
	options.timeout = timeout
	return options
}

// GetTimeout retrieves the timeout of the QueryOptions struct
func (options *QueryOptions) GetTimeout() int {
	return options.timeout
}

// Query executes a query against the graph.
func (g *Graph) Query(q string) (*QueryResult, error) {
	r, err := g.Conn.Do("GRAPH.QUERY", g.Id, q, "--compact")
	if err != nil {
		return nil, err
	}

	return QueryResultNew(g, r)
}

// ROQuery executes a read only query against the graph.
func (g *Graph) ROQuery(q string) (*QueryResult, error) {

	r, err := g.Conn.Do("GRAPH.RO_QUERY", g.Id, q, "--compact")
	if err != nil {
		return nil, err
	}

	return QueryResultNew(g, r)
}

func (g *Graph) ParameterizedQuery(q string, params map[string]interface{}) (*QueryResult, error) {
	if params != nil {
		q = BuildParamsHeader(params) + q
	}
	return g.Query(q)
}

// QueryWithOptions issues a query with the given timeout
func (g *Graph) QueryWithOptions(q string, options *QueryOptions) (*QueryResult, error) {
	var r interface{}
	var err error
	if options.timeout >= 0 {
		r, err = g.Conn.Do("GRAPH.QUERY", g.Id, q, "--compact", "timeout", options.timeout)
	} else {
		r, err = g.Conn.Do("GRAPH.QUERY", g.Id, q, "--compact")
	}
	if err != nil {
		return nil, err
	}

	return QueryResultNew(g, r)
}

// ParameterizedQueryWithOptions issues a parameterized query with the given timeout
func (g *Graph) ParameterizedQueryWithOptions(q string, params map[string]interface{}, options *QueryOptions) (*QueryResult, error) {
	if params != nil {
		q = BuildParamsHeader(params) + q
	}
	return g.QueryWithOptions(q, options)
}

// ROQueryWithOptions issues a read-only query with the given timeout
func (g *Graph) ROQueryWithOptions(q string, options *QueryOptions) (*QueryResult, error) {
	var r interface{}
	var err error
	if options.timeout >= 0 {
		r, err = g.Conn.Do("GRAPH.RO_QUERY", g.Id, q, "--compact", "timeout", options.timeout)
	} else {
		r, err = g.Conn.Do("GRAPH.RO_QUERY", g.Id, q, "--compact")
	}
	if err != nil {
		return nil, err
	}

	return QueryResultNew(g, r)
}

// Merge pattern
func (g *Graph) Merge(p string) (*QueryResult, error) {
	q := fmt.Sprintf("MERGE %s", p)
	return g.Query(q)
}

func (g *Graph) getLabel(lblIdx int) string {
	if lblIdx >= len(g.labels) {
		// Missing label, refresh label mapping table.
		g.mutex.Lock()

		// Recheck now that we've got the lock.
		if lblIdx >= len(g.labels) {
			g.labels = g.Labels()
			// Retry.
			if lblIdx >= len(g.labels) {
				// Error!
				panic("Unknown label index.")
			}
		}
		g.mutex.Unlock()
	}

	return g.labels[lblIdx]
}

func (g *Graph) getRelation(relIdx int) string {
	if relIdx >= len(g.relationshipTypes) {
		// Missing relation type, refresh relation type mapping table.
		g.mutex.Lock()

		// Recheck now that we've got the lock.
		if relIdx >= len(g.relationshipTypes) {
			g.relationshipTypes = g.RelationshipTypes()
			// Retry.
			if relIdx >= len(g.relationshipTypes) {
				// Error!
				panic("Unknown relation type index.")
			}
		}
		g.mutex.Unlock()
	}

	return g.relationshipTypes[relIdx]
}

func (g *Graph) getProperty(propIdx int) string {
	if propIdx >= len(g.properties) {
		// Missing property, refresh property mapping table.
		g.mutex.Lock()

		// Recheck now that we've got the lock.
		if propIdx >= len(g.properties) {
			g.properties = g.PropertyKeys()

			// Retry.
			if propIdx >= len(g.properties) {
				// Error!
				panic("Unknown property index.")
			}
		}
		g.mutex.Unlock()
	}

	return g.properties[propIdx]
}

// Procedures

// CallProcedure invokes procedure.
func (g *Graph) CallProcedure(procedure string, yield []string, mode QueryMode, args ...interface{}) (*QueryResult, error) {
	q := fmt.Sprintf("CALL %s(", procedure)

	tmp := make([]string, 0, len(args))
	for arg := range args {
		tmp = append(tmp, ToString(arg))
	}
	q += fmt.Sprintf("%s)", strings.Join(tmp, ","))

	if len(yield) > 0 {
		q += fmt.Sprintf(" YIELD %s", strings.Join(yield, ","))
	}

	if mode == ReadQuery {
		return g.ROQuery(q)
	}

	return g.Query(q)
}

// Labels, retrieves all node labels.
func (g *Graph) Labels() []string {
	qr, _ := g.CallProcedure("db.labels", nil, ReadQuery)

	l := make([]string, len(qr.results))

	for idx, r := range qr.results {
		l[idx] = r.GetByIndex(0).(string)
	}
	return l
}

// RelationshipTypes, retrieves all edge relationship types.
func (g *Graph) RelationshipTypes() []string {
	qr, _ := g.CallProcedure("db.relationshipTypes", nil, ReadQuery)

	rt := make([]string, len(qr.results))

	for idx, r := range qr.results {
		rt[idx] = r.GetByIndex(0).(string)
	}
	return rt
}

// PropertyKeys, retrieves all properties names.
func (g *Graph) PropertyKeys() []string {
	qr, _ := g.CallProcedure("db.propertyKeys", nil, ReadQuery)

	p := make([]string, len(qr.results))

	for idx, r := range qr.results {
		p[idx] = r.GetByIndex(0).(string)
	}
	return p
}
