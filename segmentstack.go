package main

import "sync"

var segments_stack_mutex sync.RWMutex
var segments_stack map[[32]byte][72]byte
var segments_stack_uncommit map[[32]byte][32]byte

func init() {
	segments_stack = make(map[[32]byte][72]byte)
	segments_stack_uncommit = make(map[[32]byte][32]byte)
}

const SEGMENT_STACK_TRICKLED byte = 32

func segments_stack_trickle(loopkiller map[[32]byte]struct{}, commitment [32]byte) {

	if balance_try_increase_loop(commitment) {
		return
	}

	if _, ok2 := loopkiller[commitment]; ok2 {

		balance_create_loop(commitment)
		return
	}
	loopkiller[commitment] = struct{}{}
	segments_stack_mutex.RLock()
	var txstackandbranchandamount, ok = segments_stack[commitment]
	segments_stack_mutex.RUnlock()
	var to, sumto, sum = stack_decode(txstackandbranchandamount[0:])

	if !ok {
		println("trickle non existent tx")
	}

	var did_split = balance_split_if_enough(commitment, to, sumto, sum)

	if did_split == 2 && sumto != to && sum > 0 {
		var newloopkiller map[[32]byte]struct{}
		newloopkiller = make(map[[32]byte]struct{})

		var type3 = segments_stack_type(sumto)
		if type3 == SEGMENT_STACK_TRICKLED {
			segments_stack_trickle(newloopkiller, sumto)
		}

		var type2 = segments_merkle_type(sumto)
		if type2 == SEGMENT_MERKLE_TRICKLED {
			segments_merkle_trickle(newloopkiller, sumto)
		}

		var type1 = segments_transaction_type(sumto)
		if type1 == SEGMENT_TX_TRICKLED {
			segments_transaction_trickle(newloopkiller, sumto)
		} else if type1 == SEGMENT_ANY_UNTRICKLED {
		} else if type1 == SEGMENT_UNKNOWN {
		}
	}
	if did_split > 0 {
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

}
func segments_stack_untrickle(loopkiller *[32]byte, commitment [32]byte, bal balance) {
	graph_dirty = true
}

func segments_stack_type(commit [32]byte) segment_type {
	segments_stack_mutex.RLock()
	_, ok1 := segments_stack[commit]
	segments_stack_mutex.RUnlock()
	if ok1 {
		return SEGMENT_STACK_TRICKLED
	}

	return SEGMENT_UNKNOWN
}

func segments_stack_loopdetect(norecursion, loopkiller map[[32]byte]struct{}, commitment [32]byte) bool {
	if _, ok2 := loopkiller[commitment]; ok2 {

		return true
	}

	_, is_stack_recursion := norecursion[commitment]

	norecursion[commitment] = struct{}{}

	loopkiller[commitment] = struct{}{}
	segments_stack_mutex.RLock()
	var txstackandbranchandamount, ok = segments_stack[commitment]
	segments_stack_mutex.RUnlock()
	var to, sumto, _ = stack_decode(txstackandbranchandamount[0:])

	if !ok {
		return false
	}
	if _, ok2 := loopkiller[to]; ok2 {

		return true
	}
	if _, ok2 := loopkiller[sumto]; ok2 {

		return true
	}
	{
		var type3 = segments_stack_type(to)
		if type3 == SEGMENT_STACK_TRICKLED {
			if segments_stack_loopdetect(norecursion, loopkiller, to) {
				return true
			}
		}
		var type2 = segments_merkle_type(to)
		if type2 == SEGMENT_MERKLE_TRICKLED {
			if segments_merkle_loopdetect(norecursion, loopkiller, to) {
				return true
			}
		}
		var type1 = segments_transaction_type(to)
		if type1 == SEGMENT_TX_TRICKLED {
			if segments_transaction_loopdetect(norecursion, loopkiller, to) {
				return true
			}
		} else if type1 == SEGMENT_ANY_UNTRICKLED {
		} else if type1 == SEGMENT_UNKNOWN {
		}
	}
	if !is_stack_recursion {
		var newloopkiller = make(map[[32]byte]struct{})

		var type3 = segments_stack_type(sumto)
		if type3 == SEGMENT_STACK_TRICKLED {
			return segments_stack_loopdetect(norecursion, newloopkiller, sumto)
		}
		var type2 = segments_merkle_type(sumto)
		if type2 == SEGMENT_MERKLE_TRICKLED {
			return segments_merkle_loopdetect(norecursion, newloopkiller, sumto)
		}
		var type1 = segments_transaction_type(sumto)
		if type1 == SEGMENT_TX_TRICKLED {
			return segments_transaction_loopdetect(norecursion, newloopkiller, sumto)
		} else if type1 == SEGMENT_ANY_UNTRICKLED {
		} else if type1 == SEGMENT_UNKNOWN {
		}
	}

	return false
}

func segments_stack_backgraph(backgraph map[[32]byte][][32]byte, norecursion map[[32]byte]struct{}, target, commitment [32]byte) {

	_, is_stack_recursion := norecursion[commitment]

	if is_stack_recursion {
		return
	}

	norecursion[commitment] = struct{}{}

	segments_stack_mutex.RLock()
	var txstackandbranchandamount, ok = segments_stack[commitment]
	segments_stack_mutex.RUnlock()
	var to, sumto, _ = stack_decode(txstackandbranchandamount[0:])

	if !ok {
		return
	}
	add_to_backgraph(backgraph, to, commitment)
	{
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
	}
	add_to_backgraph(backgraph, sumto, commitment)
	{
		var type3 = segments_stack_type(sumto)
		if type3 == SEGMENT_STACK_TRICKLED {
			segments_stack_backgraph(backgraph, norecursion, target, sumto)
		}
		var type2 = segments_merkle_type(sumto)
		if type2 == SEGMENT_MERKLE_TRICKLED {
			segments_merkle_backgraph(backgraph, norecursion, target, sumto)
		}
		var type1 = segments_transaction_type(sumto)
		if type1 == SEGMENT_TX_TRICKLED {
			segments_transaction_backgraph(backgraph, norecursion, target, sumto)
		} else if type1 == SEGMENT_ANY_UNTRICKLED {
		} else if type1 == SEGMENT_UNKNOWN {
		}
	}

	return
}
