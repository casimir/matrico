package data

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/gomodule/redigo/redis"
	rg "github.com/redislabs/redisgraph-go"
)

func toString(i interface{}) string {
	switch i.(type) {
	case int32, int64:
		return strconv.FormatInt(i.(int64), 10)
	default:
		return rg.ToString(i)
	}
}

type EventStream chan Event

type ClosableResult struct {
	conn redis.Conn
	*rg.QueryResult
}

type ClosableResults struct {
	conn    redis.Conn
	Results []*rg.QueryResult
}

func (cr *ClosableResult) Close() error {
	return cr.conn.Close()
}

type DataGraph struct {
	graphName    string
	pool         *redis.Pool
	eventStreams map[EventStream]struct{}
}

func New(graph string) *DataGraph {
	pool := redis.Pool{
		MaxIdle: 3,
		Dial:    func() (redis.Conn, error) { return redis.Dial("tcp", "0.0.0.0:6379") },
	}
	return &DataGraph{
		graphName:    graph,
		pool:         &pool,
		eventStreams: make(map[EventStream]struct{}),
	}
}

func (d *DataGraph) DELETEME() rg.Graph {
	conn := d.pool.Get()
	return rg.GraphNew(d.graphName, conn)
}

func (d *DataGraph) Query(q string) (ClosableResult, error) {
	// TODO trace log queries
	conn := d.pool.Get()
	graph := rg.GraphNew(d.graphName, conn)
	res, err := graph.Query(q)
	return ClosableResult{conn, res}, err
}

func (d *DataGraph) Exists(q string) (bool, error) {
	res, err := d.Query(q)
	defer res.Close()
	if err != nil {
		return false, err
	}
	return !res.Empty(), nil
}

func (d *DataGraph) Queries(qs ...string) (ClosableResults, error) {
	// TODO trace log queries
	conn := d.pool.Get()
	graph := rg.GraphNew(d.graphName, conn)
	ret := ClosableResults{conn: conn}
	for _, q := range qs {
		res, err := graph.Query(q)
		if err != nil {
			return ret, err
		}
		ret.Results = append(ret.Results, res)
	}
	return ret, nil
}

type Noder interface {
	Label() string
	Key() string
	KeyVal() interface{}
	Props() map[string]interface{}
}

func nodeSource(n Noder, name string, withProps bool) string {
	props := []string{fmt.Sprintf("%s:%v", n.Key(), toString(n.KeyVal()))}
	if withProps {
		for k, v := range n.Props() {
			props = append(props, fmt.Sprintf("%s:%v", k, rg.ToString(v)))
		}
	}
	return fmt.Sprintf("(%s:%s {%s})", name, n.Label(), strings.Join(props, ","))
}

func (d *DataGraph) NodeCreate(n Noder) error {
	q := fmt.Sprintf(`CREATE %s`, nodeSource(n, "", true))
	res, err := d.Query(q)
	defer res.Close()
	return err
}

func (d *DataGraph) NodeExists(n Noder) (bool, error) {
	q := fmt.Sprintf(`MATCH %s RETURN n`, nodeSource(n, "n", false))
	return d.Exists(q)
}

func (d *DataGraph) NodeDelete(n Noder) (bool, error) {
	q := fmt.Sprintf(`MATCH %s DELETE n`, nodeSource(n, "n", false))
	res, err := d.Query(q)
	defer res.Close()
	return !res.Empty(), err
}

func (d *DataGraph) NodeGet(n Noder) (map[string]interface{}, bool, error) {
	q := fmt.Sprintf(`MATCH %s RETURN n`, nodeSource(n, "n", false))
	res, err := d.Query(q)
	defer res.Close()
	if err != nil {
		return nil, false, err
	}
	if !res.Next() {
		return nil, false, nil
	}
	r := res.Record()
	node := r.GetByIndex(0).(*rg.Node)
	return node.Properties, true, nil
}

func (d *DataGraph) NodeSet(n Noder, props map[string]interface{}) error {
	var setProps []string
	for k, v := range props {
		setProps = append(setProps, fmt.Sprintf("n.%s=%v", k, toString(v)))
	}
	q := fmt.Sprintf(
		`MATCH %s SET %s`,
		nodeSource(n, "n", false), strings.Join(setProps, ","),
	)
	res, err := d.Query(q)
	defer res.Close()
	return err
}

func (d *DataGraph) LinkNodes(src, dst Noder, label string) error {
	q := fmt.Sprintf(
		`MATCH %s, %s MERGE (n)-[:%s]->(m)`,
		nodeSource(src, "n", false), nodeSource(dst, "m", false), label,
	)
	res, err := d.Query(q)
	defer res.Close()
	return err
}

func (d *DataGraph) ListenNextEvent() EventStream {
	stream := make(EventStream, 1)
	d.eventStreams[stream] = struct{}{}
	return stream
}

func (d *DataGraph) CancelEventListener(es EventStream) {
	if _, ok := d.eventStreams[es]; ok {
		delete(d.eventStreams, es)
	}
}

func (d *DataGraph) BroadcastEvent(event *Event) {
	for stream := range d.eventStreams {
		stream <- *event
		delete(d.eventStreams, stream)
	}
}
