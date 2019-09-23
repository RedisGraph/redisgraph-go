package redisgraph

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/gomodule/redigo/redis"
	"github.com/olekukonko/tablewriter"
)

const (
	LABELS_ADDED            string = "Labels added"
	NODES_CREATED           string = "Nodes created"
	NODES_DELETED           string = "Nodes deleted"
	RELATIONSHIPS_DELETED   string = "Relationships deleted"
	PROPERTIES_SET          string = "Properties set"
	RELATIONSHIPS_CREATED   string = "Relationships created"
	INTERNAL_EXECUTION_TIME string = "internal execution time"
)

type ResultSetColumnTypes int

const (
	COLUMN_UNKNOWN ResultSetColumnTypes = iota
	COLUMN_SCALAR
	COLUMN_NODE
	COLUMN_RELATION
)

type ResultSetScalarTypes int

const (
	VALUE_UNKNOWN ResultSetScalarTypes = iota
	VALUE_NULL
	VALUE_STRING
	VALUE_INTEGER
	VALUE_BOOLEAN
	VALUE_DOUBLE
	VALUE_ARRAY
	VALUE_EDGE
	VALUE_NODE
)

type QueryResultHeader struct {
	ColumnNames []string
	ColumnTypes []ResultSetColumnTypes
}

// QueryResult represents the Results of a query.
type QueryResult struct {
	Results    [][]interface{}
	Statistics map[string]float64
	Header     QueryResultHeader
	graph      *Graph
}

func QueryResultNew(g *Graph, response interface{}) (*QueryResult, error) {
	qr := &QueryResult{
		Results:    nil,
		Statistics: nil,
		Header: QueryResultHeader{
			ColumnNames: make([]string, 0),
			ColumnTypes: make([]ResultSetColumnTypes, 0),
		},
		graph: g,
	}

	r, _ := redis.Values(response, nil)

	// Check to see if we're encountered a run-time error.
	if err, ok := r[len(r)-1].(redis.Error); ok {
		return nil, err
	}

	if len(r) == 1 {
		qr.parseStatistics(r[0])
	} else {
		qr.parseResults(r)
		qr.parseStatistics(r[2])
	}

	return qr, nil
}

func (qr *QueryResult) Empty() bool {
	return len(qr.Results) == 0
}

func (qr *QueryResult) parseResults(raw_result_set []interface{}) {
	header := raw_result_set[0]
	qr.parseHeader(header)
	qr.parseRecords(raw_result_set)
}

func (qr *QueryResult) parseStatistics(raw_statistics interface{}) {
	statistics, _ := redis.Strings(raw_statistics, nil)
	qr.Statistics = make(map[string]float64)

	for _, rs := range statistics {
		v := strings.Split(rs, ": ")
		f, _ := strconv.ParseFloat(strings.Split(v[1], " ")[0], 64)
		qr.Statistics[v[0]] = f
	}
}

func (qr *QueryResult) parseHeader(raw_header interface{}) {
	header, _ := redis.Values(raw_header, nil)

	for _, col := range header {
		c, _ := redis.Values(col, nil)
		ct, _ := redis.Int(c[0], nil)
		cn, _ := redis.String(c[1], nil)

		qr.Header.ColumnTypes = append(qr.Header.ColumnTypes, ResultSetColumnTypes(ct))
		qr.Header.ColumnNames = append(qr.Header.ColumnNames, cn)
	}
}

func (qr *QueryResult) parseRecords(raw_result_set []interface{}) {
	records, _ := redis.Values(raw_result_set[1], nil)

	qr.Results = make([][]interface{}, len(records))

	for i, r := range records {
		cells, _ := redis.Values(r, nil)
		record := make([]interface{}, len(cells))

		for idx, c := range cells {
			t := qr.Header.ColumnTypes[idx]
			switch t {
			case COLUMN_SCALAR:
				s, _ := redis.Values(c, nil)
				record[idx] = qr.parseScalar(s)
				break
			case COLUMN_NODE:
				record[idx] = qr.parseNode(c)
				break
			case COLUMN_RELATION:
				record[idx] = qr.parseEdge(c)
				break
			default:
				panic("Unknown column type.")
			}
		}
		qr.Results[i] = record
	}
}

func (qr *QueryResult) parseProperties(props []interface{}) map[string]interface{} {
	// [[name, value type, value] X N]
	properties := make(map[string]interface{})
	for _, prop := range props {
		p, _ := redis.Values(prop, nil)
		idx, _ := redis.Int(p[0], nil)
		prop_name := qr.graph.getProperty(idx)
		prop_value := qr.parseScalar(p[1:])
		properties[prop_name] = prop_value
	}

	return properties
}

func (qr *QueryResult) parseNode(cell interface{}) *Node {
	// Node ID (integer),
	// [label string offset (integer)],
	// [[name, value type, value] X N]

	var label string
	c, _ := redis.Values(cell, nil)
	id, _ := redis.Uint64(c[0], nil)
	labels, _ := redis.Ints(c[1], nil)
	if len(labels) > 0 {
		label = qr.graph.getLabel(labels[0])
	}

	rawProps, _ := redis.Values(c[2], nil)
	properties := qr.parseProperties(rawProps)

	n := NodeNew(label, "", properties)
	n.ID = id
	return n
}

func (qr *QueryResult) parseEdge(cell interface{}) *Edge {
	// Edge ID (integer),
	// reltype string offset (integer),
	// src node ID offset (integer),
	// dest node ID offset (integer),
	// [[name, value, value type] X N]

	c, _ := redis.Values(cell, nil)
	id, _ := redis.Uint64(c[0], nil)
	r, _ := redis.Int(c[1], nil)
	relation := qr.graph.getRelation(r)

	src_node_id, _ := redis.Uint64(c[2], nil)
	dest_node_id, _ := redis.Uint64(c[3], nil)
	rawProps, _ := redis.Values(c[4], nil)
	properties := qr.parseProperties(rawProps)
	e := EdgeNew(relation, nil, nil, properties)

	e.ID = id
	e.srcNodeID = src_node_id
	e.destNodeID = dest_node_id
	return e
}

func (qr *QueryResult) parseArray(cell interface{}) []interface{} {
	var array = cell.([]interface{})
	var arrayLength = len(array)
	for i := 0; i < arrayLength; i++ {
		array[i] = qr.parseScalar(array[i].([]interface{}))
	}
	return array
}

func (qr *QueryResult) parseScalar(cell []interface{}) interface{} {
	t, _ := redis.Int(cell[0], nil)
	v := cell[1]
	var s interface{}
	switch ResultSetScalarTypes(t) {
	case VALUE_NULL:
		return nil

	case VALUE_STRING:
		s, _ = redis.String(v, nil)

	case VALUE_INTEGER:
		s, _ = redis.Int(v, nil)

	case VALUE_BOOLEAN:
		s, _ = redis.Bool(v, nil)

	case VALUE_DOUBLE:
		s, _ = redis.Float64(v, nil)

	case VALUE_ARRAY:
		s = qr.parseArray(v)

	case VALUE_EDGE:
		s = qr.parseEdge(v)

	case VALUE_NODE:
		s = qr.parseNode(v)

	case VALUE_UNKNOWN:
		panic("Unknown scalar type\n")
	}

	return s
}

// PrettyPrint prints the QueryResult to stdout, pretty-like.
func (qr *QueryResult) PrettyPrint() {
	if qr.Empty() {
		return
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetAutoFormatHeaders(false)
	table.SetHeader(qr.Header.ColumnNames)

	if len(qr.Results) > 0 {
		// Convert to [][]string.
		results := make([][]string, len(qr.Results))
		for i, record := range qr.Results {
			results[i] = make([]string, len(record))
			for j, elem := range record {
				results[i][j] = fmt.Sprint(elem)
			}
		}
		table.AppendBulk(results)
	} else {
		table.Append([]string{"No data returned."})
	}
	table.Render()

	for k, v := range qr.Statistics {
		fmt.Fprintf(os.Stdout, "\n%s %f", k, v)
	}

	fmt.Fprintf(os.Stdout, "\n")
}

func (qr *QueryResult) getStat(stat string) int {
	if val, ok := qr.Statistics[stat]; ok {
		return int(val)
	} else {
		return 0
	}
}

func (qr *QueryResult) LabelsAdded() int {
	return qr.getStat(LABELS_ADDED)
}

func (qr *QueryResult) NodesCreated() int {
	return qr.getStat(NODES_CREATED)
}

func (qr *QueryResult) NodesDeleted() int {
	return qr.getStat(NODES_DELETED)
}

func (qr *QueryResult) PropertiesSet() int {
	return qr.getStat(PROPERTIES_SET)
}

func (qr *QueryResult) RelationshipsCreated() int {
	return qr.getStat(RELATIONSHIPS_CREATED)
}

func (qr *QueryResult) RelationshipsDeleted() int {
	return qr.getStat(RELATIONSHIPS_DELETED)
}

func (qr *QueryResult) RunTime() int {
	return qr.getStat(INTERNAL_EXECUTION_TIME)
}
