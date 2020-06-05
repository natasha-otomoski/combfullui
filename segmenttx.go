package main

import "sync"

var segments_transaction_mutex sync.RWMutex

var segments_transaction_uncommit map[[32]byte][32]byte
var segments_transaction_data map[[32]byte][22][32]byte
var segments_transaction_next map[[32]byte][2][32]byte
var segments_transaction_doublespends map[[32]byte][2][32]byte

func init() {
	segments_transaction_uncommit = make(map[[32]byte][32]byte)
	segments_transaction_data = make(map[[32]byte][22][32]byte)
	segments_transaction_next = make(map[[32]byte][2][32]byte)
	segments_transaction_doublespends = make(map[[32]byte][2][32]byte)
}

const SEGMENT_TX_TRICKLED byte = 4
const SEGMENT_ANY_UNTRICKLED byte = 8

func segments_transaction_trickle(loopkiller map[[32]byte]struct{}, commitment [32]byte) {

	if balance_try_increase_loop(commitment) {
		return
	}

	if _, ok2 := loopkiller[commitment]; ok2 {

		balance_create_loop(commitment)
		return
	}
	loopkiller[commitment] = struct{}{}

	var txidandto, ok = segments_transaction_next[commitment]
	var to = txidandto[1]

	balance_do(commitment, to, 0xffffffffffffffff)

	if !ok {
		println("trickle non existent tx")
	}

	var type3 = segments_stack_type(to)
	if type3 == SEGMENT_STACK_TRICKLED {
		segments_stack_trickle(loopkiller, to)
	}

	var type2 = segments_merkle_type(to)
	if type2 == SEGMENT_MERKLE_TRICKLED {
		segments_merkle_trickle(loopkiller, to)
	}
	var type1 = segments_transaction_type(to)
	if type1 == SEGMENT_TX_TRICKLED {
		segments_transaction_trickle(loopkiller, to)
	} else if type1 == SEGMENT_ANY_UNTRICKLED {
	} else if type1 == SEGMENT_UNKNOWN {
	}

}
func segments_transaction_untrickle(loopkiller *[32]byte, commitment [32]byte, bal balance) {
	graph_dirty = true
}

func segments_transaction_type(commit [32]byte) segment_type {
	_, ok1 := segments_transaction_next[commit]

	if ok1 {
		return SEGMENT_TX_TRICKLED
	}

	return SEGMENT_UNKNOWN
}

func segments_transaction_loopdetect(norecursion, loopkiller map[[32]byte]struct{}, commitment [32]byte) bool {
	if _, ok2 := loopkiller[commitment]; ok2 {
		return true
	}
	loopkiller[commitment] = struct{}{}

	var txidandto, ok = segments_transaction_next[commitment]
	var to = txidandto[1]

	if !ok {
		return false
	}
	if _, ok2 := loopkiller[to]; ok2 {

		return true
	}
	var type3 = segments_stack_type(to)
	if type3 == SEGMENT_STACK_TRICKLED {
		return segments_stack_loopdetect(norecursion, loopkiller, to)
	}
	var type2 = segments_merkle_type(to)
	if type2 == SEGMENT_MERKLE_TRICKLED {
		return segments_merkle_loopdetect(norecursion, loopkiller, to)
	}
	var type1 = segments_transaction_type(to)
	if type1 == SEGMENT_TX_TRICKLED {
		return segments_transaction_loopdetect(norecursion, loopkiller, to)
	} else if type1 == SEGMENT_ANY_UNTRICKLED {
	} else if type1 == SEGMENT_UNKNOWN {
	}

	return false
}

func segments_transaction_backgraph(backgraph map[[32]byte][][32]byte, norecursion map[[32]byte]struct{}, target, commitment [32]byte) {

	_, is_stack_recursion := norecursion[commitment]

	if is_stack_recursion {
		return
	}

	norecursion[commitment] = struct{}{}

	var txidandto, ok = segments_transaction_next[commitment]
	var to = txidandto[1]

	if !ok {
		return
	}

	add_to_backgraph(backgraph, to, commitment)

	var type3 = segments_stack_type(to)
	if type3 == SEGMENT_STACK_TRICKLED {
		segments_stack_backgraph(backgraph, norecursion, target, to)
	}
	var type2 = segments_merkle_type(to)
	if type2 == SEGMENT_MERKLE_TRICKLED {
		segments_merkle_backgraph(backgraph, norecursion, target, to)
	}
	var type1 = segments_transaction_type(to)
	if type1 == SEGMENT_TX_TRICKLED {
		segments_transaction_backgraph(backgraph, norecursion, target, to)
	} else if type1 == SEGMENT_ANY_UNTRICKLED {
	} else if type1 == SEGMENT_UNKNOWN {
	}

	return
}
