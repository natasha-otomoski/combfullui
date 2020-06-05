package main

//import "fmt"

func add_to_backgraph(backgraph map[[32]byte][][32]byte, from, to [32]byte) {
	if len(backgraph[from]) == 0 {
		backgraph[from] = append(backgraph[from], to)
		return
	}

	for _, val := range backgraph[from] {
		if val == to {
			return
		}
	}
	backgraph[from] = append(backgraph[from], to)
}

func anonymize(bases map[[32]byte]struct{}, target [32]byte) map[[32]byte]struct{} {

	// build backgraph
	var backgraph = make(map[[32]byte][][32]byte)

	for combbase := range bases {

		segments_coinbase_backgraph(backgraph, make(map[[32]byte]struct{}), target, combbase)

	}

	// walk the backgraph from target

	var pred [][32]byte
	var saw = make(map[[32]byte]struct{})

	pred = append(pred, target)

	for i := 0; i < len(pred); i++ {
		if _, ok := saw[pred[i]]; !ok {
			saw[pred[i]] = struct{}{}
			//fmt.Printf("%X\n", pred[i])
			pred = append(pred, backgraph[pred[i]]...)
		}
	}
	backgraph = nil
	pred = nil
	return saw
}
