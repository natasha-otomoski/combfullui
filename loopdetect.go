package main

func loopdetect(norecursion, loopkiller map[[32]byte]struct{}, to [32]byte) bool {
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
