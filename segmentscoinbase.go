package main

import "fmt"

func segments_coinbase_untrickle_auto(previous [32]byte, next [32]byte) {
	var tag, ok = commits[previous]
	if !ok {
		return
	}

	var t = tag

	var bal = Coinbase(uint64(t.height))

	if balance_check(previous, next) == bal {
		segments_coinbase_untrickle(previous, next, bal)
	}
}

func segments_coinbase_untrickle(previous [32]byte, next [32]byte, bal balance) {
	graph_dirty = true
}

func segments_coinbase_trickle_auto(commitment [32]byte, to [32]byte) {
	var tag, ok = commits[commitment]
	if !ok {
		return
	}
	if various_debug_prints_and_self_checking {
		fmt.Printf("segments_coinbase_trickle_auto: %X %X\n", commitment, to)
	}
	var t = tag

	var bal = Coinbase(uint64(t.height))

	if balance_check(commitment, to) != bal {
		segments_coinbase_trickle(commitment, to, bal)
	}
}

func segments_coinbase_trickle(commitment [32]byte, to [32]byte, bal balance) {
	if various_debug_prints_and_self_checking {
		fmt.Printf("segments_coinbase_trickle: %X %X % 21d\n", commitment, to, bal)
	}
	balance_do(commitment, to, bal)

	var type3 = segments_stack_type(to)
	if type3 == SEGMENT_STACK_TRICKLED {
		segments_stack_trickle(make(map[[32]byte]struct{}), to)
	}

	var type2 = segments_merkle_type(to)
	if type2 == SEGMENT_MERKLE_TRICKLED {
		segments_merkle_trickle(make(map[[32]byte]struct{}), to)
	}

	var type1 = segments_transaction_type(to)
	if type1 == SEGMENT_TX_TRICKLED {
		segments_transaction_trickle(make(map[[32]byte]struct{}), to)
	} else if type1 == SEGMENT_ANY_UNTRICKLED {
	} else if type1 == SEGMENT_UNKNOWN {
	}

}

func segments_coinbase_mine(commitment [32]byte, height uint64) {

	var amount = Coinbase(height)

	wallet_mutex.RLock()

	var mykey, is_mykey = wallet_commitments[commitment]

	wallet_mutex.RUnlock()

	balance_create_coinbase(commitment, amount)

	if is_mykey {
		segments_coinbase_trickle(commitment, mykey, amount)
	}

	segments_stack_mutex.RLock()

	if hash, ok3 := segments_stack_uncommit[commitment]; ok3 {
		segments_stack_mutex.RUnlock()
		segments_coinbase_trickle(commitment, hash, amount)
	} else {

		segments_stack_mutex.RUnlock()
	}

	segments_merkle_mutex.RLock()

	if hash, ok3 := segments_merkle_uncommit[commitment]; ok3 {
		segments_merkle_mutex.RUnlock()
		segments_coinbase_trickle(commitment, hash, amount)
	} else {

		segments_merkle_mutex.RUnlock()
	}

	segments_transaction_mutex.RLock()

	if hash, ok3 := segments_transaction_uncommit[commitment]; ok3 {
		segments_transaction_mutex.RUnlock()

		if mykey != hash {

			segments_coinbase_trickle(commitment, hash, amount)

		}
	} else {

		segments_transaction_mutex.RUnlock()
	}
}

func segments_coinbase_unmine(commitment [32]byte, height uint64) {

	wallet_mutex.RLock()

	var mykey, is_mykey = wallet_commitments[commitment]

	wallet_mutex.RUnlock()

	var amount = Coinbase(height)

	segments_transaction_mutex.RLock()

	if hash, ok3 := segments_transaction_uncommit[commitment]; ok3 {
		segments_transaction_mutex.RUnlock()

		if mykey != hash {

			segments_coinbase_untrickle(commitment, hash, amount)

		}
	} else {

		segments_transaction_mutex.RUnlock()
	}

	segments_merkle_mutex.RLock()

	if hash, ok3 := segments_merkle_uncommit[commitment]; ok3 {
		segments_merkle_mutex.RUnlock()
		segments_coinbase_untrickle(commitment, hash, amount)
	} else {

		segments_merkle_mutex.RUnlock()
	}

	segments_stack_mutex.RLock()

	if hash, ok3 := segments_stack_uncommit[commitment]; ok3 {
		segments_stack_mutex.RUnlock()
		segments_coinbase_untrickle(commitment, hash, amount)
	} else {

		segments_stack_mutex.RUnlock()
	}

	if is_mykey {
		segments_coinbase_untrickle(commitment, mykey, amount)
	}

	if !reset_whole_graph_on_reorg {
		balance_destroy_coinbase(commitment, amount)
	} else {
		graph_dirty = true
	}
}

type segment_type = byte

const SEGMENT_UNKNOWN byte = 0
const SEGMENT_COINBASE_TRICKLED byte = 1
const SEGMENT_COINBASE_UNTRICKLED byte = 2
const SEGMENT_COINBASE byte = 3

func segments_coinbase_backgraph(backgraph map[[32]byte][][32]byte, norecursion map[[32]byte]struct{}, target, commitment [32]byte) {

	segments_transaction_mutex.RLock()

	if hash, ok3 := segments_transaction_uncommit[commitment]; ok3 {
		segments_transaction_mutex.RUnlock()
		segments_transaction_backgraph(backgraph, norecursion, target, hash)
	} else {
		segments_transaction_mutex.RUnlock()
	}

	segments_merkle_mutex.RLock()

	if hash, ok3 := segments_merkle_uncommit[commitment]; ok3 {
		segments_merkle_mutex.RUnlock()
		segments_merkle_backgraph(backgraph, norecursion, target, hash)
	} else {

		segments_merkle_mutex.RUnlock()
	}

	segments_stack_mutex.RLock()

	if hash, ok3 := segments_stack_uncommit[commitment]; ok3 {
		segments_stack_mutex.RUnlock()
		segments_stack_backgraph(backgraph, norecursion, target, hash)
	} else {

		segments_stack_mutex.RUnlock()
	}
}
