package overpass

func (r *Result) getNode(id int64) *Node {
	node, ok := r.Nodes[id]
	if !ok {
		node = &Node{Meta: Meta{ID: id}}
		r.Nodes[id] = node
	}
	return node
}

func (r *Result) getWay(id int64) *Way {
	way, ok := r.Ways[id]
	if !ok {
		way = &Way{Meta: Meta{ID: id}}
		r.Ways[id] = way
	}
	return way
}

func (r *Result) getRelation(id int64) *Relation {
	relation, ok := r.Relations[id]
	if !ok {
		relation = &Relation{Meta: Meta{ID: id}}
		r.Relations[id] = relation
	}
	return relation
}

func (r *Result) getOldNode(id int64) *Node {
	node, ok := r.OldNodes[id]
	if !ok {
		node = &Node{Meta: Meta{ID: id}}
		r.OldNodes[id] = node
	}
	return node
}

func (r *Result) getOldWay(id int64) *Way {
	way, ok := r.OldWays[id]
	if !ok {
		way = &Way{Meta: Meta{ID: id}}
		r.OldWays[id] = way
	}
	return way
}

func (r *Result) getOldRelation(id int64) *Relation {
	relation, ok := r.OldRelations[id]
	if !ok {
		relation = &Relation{Meta: Meta{ID: id}}
		r.OldRelations[id] = relation
	}
	return relation
}
