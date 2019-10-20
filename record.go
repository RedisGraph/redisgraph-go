package redisgraph

type Record struct {
	values	[]interface{}
	keys	[]string
}

func recordNew(values []interface{}, keys []string) *Record {
	r := &Record {
		values: values,
		keys: keys,
	}

	return r
}

func (r *Record) Keys() []string {
	return r.keys
}

func (r *Record) Values() []interface{} {
	return r.values
}

func (r *Record) Get(key string) (interface{}, bool) {
	// TODO: switch from []string to map[string]int
	for i := range r.keys {
		if r.keys[i] == key {
			return r.values[i], true
		}
	}
	return nil, false
}

func (r *Record) GetByIndex(index int) interface{} {
	if(index < len(r.values)) {
		return r.values[index]
	} else {
		return nil
	}
}
