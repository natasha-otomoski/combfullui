package main

import "fmt"
import "crypto/sha256"
import "net/http"
import "github.com/gorilla/mux"

func uint16_reverse(n uint16) uint16 {
	n = ((n >> 1) & 0x5555) | ((n << 1) & 0xaaaa)
	n = ((n >> 2) & 0x3333) | ((n << 2) & 0xcccc)
	n = ((n >> 4) & 0x0f0f) | ((n << 4) & 0xf0f0)
	n = ((n >> 8) & 0x00ff) | ((n << 8) & 0xff00)
	return n
}

func commits_entrenched_height(commitment1, commitment2 [32]byte) uint32 {
	commitment1 = commit(commitment1[0:])
	commitment2 = commit(commitment2[0:])

	commits_mutex.RLock()

	var tag1, ok1 = commits[commitment1]
	var tag2, ok2 = commits[commitment2]

	commits_mutex.RUnlock()

	if !(ok1 && ok2) {
		return ^uint32(0)
	}

	if utag_cmp(&tag1, &tag2) < 0 {
		return tag1.height
	}
	return tag2.height
}

func gui_contract(w http.ResponseWriter, r *http.Request) {

	fmt.Fprintf(w, `<html><head></head><body><a href="/">&larr; Back to home</a><br />`)
	defer fmt.Fprintf(w, `</body></html>`)

	fmt.Fprint(w, `
		<h1>Basic contract - enter decider</h1>

		<p>Decider can be short (to be decided) or long (already decided).</p>

		<div style="float:left">Decider: <input name='decider' id='decider' style="font-family:monospace;width:83em;font-size:80%" maxlength="132" /></div>

		<div id="buttons" style="float:left;visibility:hidden"><ul style="margin:auto 0;padding:auto 0">
			<li><a href="#" id="doadl">Amount decided later contract</a></li>
			<li><a href="#" id="dotrade">Comb trade contract</a></li>
		</ul></div>

		<script type="text/javascript">
		  var foo = function() {
		    var decider = document.getElementById("decider").value;
                    if ((decider.length == 64) || (decider.length == 132)) {
                      document.getElementById("buttons").style.visibility = 'visible';
		    } else {
                      document.getElementById("buttons").style.visibility = 'hidden';
                    }
		    document.getElementById("doadl").href="/basiccontract/amtdecidedlater/"+decider;
		    document.getElementById("dotrade").href="/basiccontract/trade/"+decider;
		    return false;
		  };
		document.getElementById("decider").oninput = foo;
		document.getElementById("decider").onpropertychange = foo;
		</script>
	`)
}

func do_decider_common(decider string) (short_decider, left_decider, right_decider, left_decider_tip, right_decider_tip [32]byte, number uint16, err1 error) {

	if len(decider) == 64 {
		err1 = checkHEXPAND32(decider)
		if err1 != nil {
			return short_decider, left_decider, right_decider, left_decider_tip, right_decider_tip, number, err1
		}

		var rawdecider = []byte(decider)

		short_decider = hexpand2byte32(rawdecider[0:64])

	} else if len(decider) == 132 {
		err1 = checkHEXPAND66(decider)
		if err1 != nil {
			return short_decider, left_decider, right_decider, left_decider_tip, right_decider_tip, number, err1
		}
		var rawdecider = []byte(decider)

		num := hexpand2byte2(rawdecider[0:4])

		left := hexpand2byte32(rawdecider[4 : 4+64])
		right := hexpand2byte32(rawdecider[4+64 : 4+64+64])

		left_decider = left
		right_decider = right

		number = uint16_reverse(uint16(256*int(num[0]) + int(num[1])))

		for i := 0; i < 65535-int(number); i++ {
			left = sha256.Sum256(left[0:])
		}
		for i := 0; i < int(number); i++ {
			right = sha256.Sum256(right[0:])
		}

		left_decider_tip = left
		right_decider_tip = right

		var buf3_a0 [96]byte

		copy(buf3_a0[32:64], left[0:])
		copy(buf3_a0[64:96], right[0:])

		short_decider = sha256.Sum256(buf3_a0[0:])

	}
	return short_decider, left_decider, right_decider, left_decider_tip, right_decider_tip, number, err1

}

func gui_adl_contract(w http.ResponseWriter, r *http.Request) {

	fmt.Fprintf(w, `<html><head></head><body><a href="/basiccontract/">&larr; Back to basic contract</a><br />`)
	defer fmt.Fprintf(w, `</body></html>`)

	vars := mux.Vars(r)
	var decider_str = vars["decider"]

	short_decider, entrenched_l, entrenched_r, _, _, num, err1 := do_decider_common(decider_str)
	if err1 != nil {
		fmt.Fprintf(w, "error: decider invalid: %s", err1)
		return
	}

	fmt.Fprintf(w, `
		<h1>Amount decided later contract</h1>
		<div>Decider: <font size="3"><tt id="decider">%s</tt></font></div>
		<div>Associated short decider: %s</div>
		<div>Merkle branch: <tt id="merklebranch">%d</tt></div>
		<div>Min amount: <input name='min_amount' id='min_amount' /></div>
		<div>Max amount: <input name='max_amount' id='max_amount' /></div>
		<div>Change COMB address: <input name='left_address' id='left_address' style="font-family:monospace;width:45em" maxlength="64" /></div>
		<div>Exact COMB address: <input name='right_address' id='right_address' style="font-family:monospace;width:45em" maxlength="64" /></div>
	`, decider_str, Hexpand32(short_decider[0:]), uint16_reverse(num))

	if len(decider_str) == 132 {

		eheight := commits_entrenched_height(entrenched_l, entrenched_r)

		if eheight < ^uint32(0) {

			fmt.Fprintf(w, `
				<div>Entrenched height: <tt id="entrenchedheight">%d</tt></div>
			`, eheight)

		} else {
			fmt.Fprintf(w, `
				<div>Entrenched height: <tt id="entrenchedheight">decider not entrenched</tt></div>
			`)
		}

	}

	fmt.Fprintf(w, `

		<div><a href="#" id="ok">ok</a></div>

		<script type="text/javascript">
		  var foo = function() {
		    var min_amount = document.getElementById("min_amount").value;
		    var max_amount = document.getElementById("max_amount").value;
		    var left_address = document.getElementById("left_address").value;
		    var right_address = document.getElementById("right_address").value;
                    min_amount = parseInt(min_amount).toString(16).toUpperCase().padStart(16, "0");
                    max_amount = parseInt(max_amount).toString(16).toUpperCase().padStart(16, "0");

		    document.getElementById("ok").href="/basiccontract/amtdecidedlatermkl/%s/"+
			min_amount+"/"+max_amount+"/"+left_address+"/"+right_address;
		    return false;
		  };
		document.getElementById("min_amount").oninput = foo;
		document.getElementById("min_amount").onpropertychange = foo;
		document.getElementById("max_amount").oninput = foo;
		document.getElementById("max_amount").onpropertychange = foo;
		document.getElementById("left_address").oninput = foo;
		document.getElementById("left_address").onpropertychange = foo;
		document.getElementById("right_address").oninput = foo;
		document.getElementById("right_address").onpropertychange = foo;
		</script>
	`, decider_str)
}

func gui_trade_contract(w http.ResponseWriter, r *http.Request) {

	fmt.Fprintf(w, `<html><head></head><body><a href="/basiccontract/">&larr; Back to basic contract</a><br />`)
	defer fmt.Fprintf(w, `</body></html>`)

	vars := mux.Vars(r)
	var decider_str = vars["decider"]

	short_decider, entrenched_l, entrenched_r, _, _, num, err1 := do_decider_common(decider_str)
	if err1 != nil {
		fmt.Fprintf(w, "error: decider invalid: %s", err1)
		return
	}

	fmt.Fprintf(w, `
		<h1>Comb trade contract</h1>
		<div>Decider: <font size="3"><tt id="decider">%s</tt></font></div>
		<div>Associated short decider: %s</div>
		<div>Merkle branch: <tt id="merklebranch">%d</tt></div>
		<div>Trade bit mask: <input name='trade_bit' id='trade_bit' /></div>
		<div>Forward COMB address: <input name='forward_address' id='forward_address' style="font-family:monospace;width:45em" maxlength="64" /></div>
		<div>Rollback COMB address: <input name='rollback_address' id='rollback_address' style="font-family:monospace;width:45em" maxlength="64" /></div>
	`, decider_str, Hexpand32(short_decider[0:]), uint16_reverse(num))

	if len(decider_str) == 132 {

		eheight := commits_entrenched_height(entrenched_l, entrenched_r)

		if eheight < ^uint32(0) {

			fmt.Fprintf(w, `
				<div>Entrenched height: <tt id="entrenchedheight">%d</tt></div>
			`, eheight)

		} else {
			fmt.Fprintf(w, `
				<div>Entrenched height: <tt id="entrenchedheight">decider not entrenched</tt></div>
			`)
		}

	}

	fmt.Fprintf(w, `

		<div><a href="#" id="ok">ok</a></div>

		<script type="text/javascript">
		  var foo = function() {
		    var trade_bit = document.getElementById("trade_bit").value;
		    var forward_address = document.getElementById("forward_address").value;
		    var rollback_address = document.getElementById("rollback_address").value;
                    trade_bit = parseInt(trade_bit).toString(16).toUpperCase().padStart(4, "0");

		    document.getElementById("ok").href="/basiccontract/trademkl/%s/"+
			trade_bit+"/"+forward_address+"/"+rollback_address;
		    return false;
		  };
		document.getElementById("trade_bit").oninput = foo;
		document.getElementById("trade_bit").onpropertychange = foo;
		document.getElementById("forward_address").oninput = foo;
		document.getElementById("forward_address").onpropertychange = foo;
		document.getElementById("rollback_address").oninput = foo;
		document.getElementById("rollback_address").onpropertychange = foo;
		</script>
	`, decider_str)
}

func gui_adl_contract_merkle(w http.ResponseWriter, r *http.Request) {

	fmt.Fprintf(w, `<html><head></head><body><a href="/basiccontract/">&larr; Back to basic contract</a><br />`)
	defer fmt.Fprintf(w, `</body></html>`)

	vars := mux.Vars(r)
	var decider_str = vars["decider"]

	short_decider, left_decider, right_decider, left_decider_tip, right_decider_tip, num, err1 := do_decider_common(decider_str)
	if err1 != nil {
		fmt.Fprintf(w, "error: decider invalid: %s", err1)
		return
	}

	var min_str, max_str, left_str, right_str = vars["min"], vars["max"], vars["left"], vars["right"]

	err2 := checkHEX32(left_str)
	if err2 != nil {
		fmt.Fprintf(w, "error: left address invalid: %s", err2)
		return
	}

	err3 := checkHEX32(right_str)
	if err3 != nil {
		fmt.Fprintf(w, "error: right address invalid: %s", err3)
		return
	}

	err4 := checkHEX8(min_str)
	if err4 != nil {
		fmt.Fprintf(w, "error: min amount invalid: %s", err4)
		return
	}

	err5 := checkHEX8(max_str)
	if err5 != nil {
		fmt.Fprintf(w, "error: max amount invalid: %s", err5)
		return
	}

	left := hex2byte32([]byte(left_str))
	right := hex2byte32([]byte(right_str))
	min_bytes := hex2byte8([]byte(min_str))
	max_bytes := hex2byte8([]byte(max_str))

	var min uint64
	min += uint64(min_bytes[0])
	min <<= 8
	min += uint64(min_bytes[1])
	min <<= 8
	min += uint64(min_bytes[2])
	min <<= 8
	min += uint64(min_bytes[3])
	min <<= 8
	min += uint64(min_bytes[4])
	min <<= 8
	min += uint64(min_bytes[5])
	min <<= 8
	min += uint64(min_bytes[6])
	min <<= 8
	min += uint64(min_bytes[7])

	var max uint64
	max += uint64(max_bytes[0])
	max <<= 8
	max += uint64(max_bytes[1])
	max <<= 8
	max += uint64(max_bytes[2])
	max <<= 8
	max += uint64(max_bytes[3])
	max <<= 8
	max += uint64(max_bytes[4])
	max <<= 8
	max += uint64(max_bytes[5])
	max <<= 8
	max += uint64(max_bytes[6])
	max <<= 8
	max += uint64(max_bytes[7])

	var branch = amount_decided_later_generate_branch(min, max, num, left, right)

	var address = merkle(short_decider[0:], branch[17][0:])

	fmt.Fprintf(w, `
		<h1>Amount decided later contract</h1>
		<div>Decider: <font size="3"><tt id="decider">%s</tt></font></div>
		<div>Associated short decider: %s</div>
		<div>Min amount: <tt name='min_amount' id='min_amount'>%d.%08d COMB</tt></div>
		<div>Max amount: <tt name='max_amount' id='max_amount'>%d.%08d COMB</tt></div>
		<div>Change COMB address: <tt name='left_address' id='left_address' style="font-family:monospace;width:45em">%X</tt></div>
		<div>Exact COMB address: <tt name='right_address' id='right_address' style="font-family:monospace;width:45em">%X</tt></div>
		<div>Merkle root: <tt id="merkleroot">%x</tt></div>
		<div>Address: <tt id="address">%X</tt></div>
	`, decider_str, Hexpand32(short_decider[0:]), combs(min), nats(min), combs(max), nats(max), left, right, branch[17], address)

	if len(decider_str) == 132 {

		var exact = amount_decided_later_exact(min, max, num)

		var adl_stack = amount_decided_later_generate_stack(min, max, num, left, right)

		if wallet_selfmining_links {

			fmt.Fprintf(w, `<h1><a href="/mining/mine/%X/%s">%X</a></h1>`, commit(left_decider[:]), serializeutxotag(forcecoinbasefirst(makefaketag())), left_decider)
			fmt.Fprintf(w, `<h1><a href="/mining/mine/%X/%s">%X</a></h1>`, commit(right_decider[:]), serializeutxotag(forcecoinbasefirst(makefaketag())), right_decider)
		}

		fmt.Fprintf(w, `
			<div>Exact amount: <tt name='max_amount' id='max_amount'>%d.%08d COMB</tt></div>
			<div>Merkle branch: <tt id="merklebranch">%d</tt></div>
			<div>FullBranch: <a href="/merkle/data/%X%X%X%X%X%X%X%X%X%X%X%X%X%X%X%X%X%X%X%X%X%X">load</a></div>
			<div>FullStack: <a href="/stack/data/%X">load</a></div>
		`, combs(exact), nats(exact), uint16_reverse(num),
			left_decider_tip, right_decider_tip, left_decider, right_decider,
			branch[1], branch[2], branch[3], branch[4], branch[5], branch[6],
			branch[7], branch[8], branch[9], branch[10], branch[11], branch[12], branch[13], branch[14], branch[15], branch[16], branch[0], [32]byte{},
			adl_stack)
	}
}

func gui_trade_contract_merkle(w http.ResponseWriter, r *http.Request) {

	fmt.Fprintf(w, `<html><head></head><body><a href="/basiccontract/">&larr; Back to basic contract</a><br />`)
	defer fmt.Fprintf(w, `</body></html>`)

	vars := mux.Vars(r)
	var decider_str = vars["decider"]

	short_decider, left_decider, right_decider, left_decider_tip, right_decider_tip, num, err1 := do_decider_common(decider_str)
	if err1 != nil {
		fmt.Fprintf(w, "error: decider invalid: %s", err1)
		return
	}

	var trade_str, forward_str, rollback_str = vars["trade"], vars["forward"], vars["rollback"]

	err2 := checkHEX32(forward_str)
	if err2 != nil {
		fmt.Fprintf(w, "error: forward address invalid: %s", err2)
		return
	}

	err3 := checkHEX32(rollback_str)
	if err3 != nil {
		fmt.Fprintf(w, "error: rollback address invalid: %s", err3)
		return
	}

	err4 := checkHEX2(trade_str)
	if err4 != nil {
		fmt.Fprintf(w, "error: trade bits invalid: %s", err4)
		return
	}

	forward := hex2byte32([]byte(forward_str))
	rollback := hex2byte32([]byte(rollback_str))
	trade := hex2uint16([]byte(trade_str))

	var branch = comb_trade_generate_branch(uint16_reverse(trade), num, forward, rollback)

	var address = merkle(short_decider[0:], branch[17][0:])

	fmt.Fprintf(w, `
		<h1>Comb trade contract</h1>
		<div>Decider: <font size="3"><tt id="decider">%s</tt></font></div>
		<div>Associated short decider: %s</div>
		<div>Trade bit mask: <tt name='trade_bit' id='trade_bit'>%d</tt></div>
		<div>Forward COMB address: <tt name='forward_address' id='forward_address' style="font-family:monospace;width:45em">%X</tt></div>
		<div>Rollback COMB address: <tt name='rollback_address' id='rollback_address' style="font-family:monospace;width:45em">%X</tt></div>
		<div>Merkle root: <tt id="merkleroot">%x</tt></div>
		<div>Address: <tt id="address">%X</tt></div>
	`, decider_str, Hexpand32(short_decider[0:]), trade, forward, rollback, branch[17], address)

	if len(decider_str) == 132 {

		if wallet_selfmining_links {

			fmt.Fprintf(w, `<h1><a href="/mining/mine/%X/%s">%X</a></h1>`, commit(left_decider[:]), serializeutxotag(forcecoinbasefirst(makefaketag())), left_decider)
			fmt.Fprintf(w, `<h1><a href="/mining/mine/%X/%s">%X</a></h1>`, commit(right_decider[:]), serializeutxotag(forcecoinbasefirst(makefaketag())), right_decider)
		}

		fmt.Fprintf(w, `
			<div>Merkle branch: <tt id="merklebranch">%d</tt></div>
			<div>FullBranch: <a href="/merkle/data/%X%X%X%X%X%X%X%X%X%X%X%X%X%X%X%X%X%X%X%X%X%X">load</a></div>
		`, uint16_reverse(num), left_decider_tip, right_decider_tip, left_decider, right_decider,
			branch[1], branch[2], branch[3], branch[4], branch[5], branch[6],
			branch[7], branch[8], branch[9], branch[10], branch[11], branch[12], branch[13], branch[14], branch[15], branch[16], branch[0], [32]byte{})
	}
}

func merkle_tree_generate_branch(tree *[65536][32]byte, branch_id uint16) (branch [18][32]byte) {
	branch[0] = (*tree)[branch_id]
	for j := 0; j < 16; j++ {

		branch[j+1] = (*tree)[branch_id^1]

		for i := 0; i < 1<<uint(15-j); i++ {
			(*tree)[i] = merkle((*tree)[2*i][0:], (*tree)[2*i+1][0:])
		}
		branch_id >>= 1
	}
	branch[17] = (*tree)[0]
	return branch
}

func amount_decided_later_exact(min_amount, max_amount uint64, branch_id uint16) uint64 {
	var amt uint64

	if (min_amount < (1 << 47)) && (max_amount < (1 << 47)) {

		amt = uint64(int64(min_amount<<16)+(int64(branch_id)*int64((max_amount-min_amount)))) >> 16

	} else {
		amt = uint64(int64(min_amount<<12)+(int64(branch_id)*int64((max_amount-min_amount)>>4))) >> 12
	}
	return amt
}

func amount_decided_later_generate_stack(min_amount, max_amount uint64, branch_id uint16, left_address, right_address [32]byte) (stack [72]byte) {

	copy(stack[0:32], left_address[0:])
	copy(stack[32:64], right_address[0:])

	var amt = amount_decided_later_exact(min_amount, max_amount, branch_id)

	stack[64] = byte(amt >> 56)
	stack[65] = byte(amt >> 48)
	stack[66] = byte(amt >> 40)
	stack[67] = byte(amt >> 32)
	stack[68] = byte(amt >> 24)
	stack[69] = byte(amt >> 16)
	stack[70] = byte(amt >> 8)
	stack[71] = byte(amt >> 0)

	return stack
}

func amount_decided_later_generate_branch(min_amount, max_amount uint64, branch_id uint16, left_address, right_address [32]byte) [18][32]byte {
	var buf [65536][32]byte
	for i := range buf {

		var stack = amount_decided_later_generate_stack(min_amount, max_amount, uint16(i), left_address, right_address)

		buf[i] = sha256.Sum256(stack[0:])

	}
	return merkle_tree_generate_branch(&buf, branch_id)
}

func comb_trade_generate_branch(trade_bit uint16, branch_id uint16, forward_address, rollback_address [32]byte) [18][32]byte {
	var buf [65536][32]byte
	for i := range buf {

		if (uint16(i) & trade_bit) == trade_bit {
			buf[i] = forward_address
		} else {
			buf[i] = rollback_address
		}
	}
	return merkle_tree_generate_branch(&buf, branch_id)
}
