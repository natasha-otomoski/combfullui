package main

const reset_whole_graph_on_reorg bool = true

var graph_dirty bool

func resetgraph() {

	if !graph_dirty {
		return
	}

	balance_mutex.Lock()

	balance_edge = make(map[[32]byte]balance)
	balance_node = make(map[[32]byte]balance)
	balance_loop = make(map[[32]byte]balance)

	balance_mutex.Unlock()

	for commitment := range combbases {

		basetag := commits[commitment]

		var btag = basetag

		var bheight = uint64(btag.height)

		segments_coinbase_mine(commitment, bheight)

	}

	graph_dirty = false
}
