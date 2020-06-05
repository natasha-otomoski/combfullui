package main

import (
	"fmt"
	//"github.com/gorilla/mux"
	"net/http"
)

func dump_liq_stack_item(w http.ResponseWriter, key [32]byte) {
	var val, ok = segments_stack[key]

	if !ok {
		return
	}

	var ckey = commit(key[:])

	var to, sumto, sum = stack_decode(val[0:])

	balance_mutex.RLock()

	var bal = balance_node[key]
	var balto = balance_node[to]
	var balsumto = balance_node[sumto]

	balance_mutex.RUnlock()

	fmt.Fprintf(w, `<h2>%X %d.%08d COMB</h2>`+"\n", key, combs(bal), nats(bal))
	if wallet_selfmining_links {
		fmt.Fprintf(w, `<a href="/mining/mine/%X/%s">mine</a>`+"\n", ckey, serializeutxotag(forcecoinbasefirst(makefaketag())))
	} else {
		fmt.Fprintf(w, `<ul><li>%s mine address <b>any Bitcoin paid here will be lost forever</b></li></ul>`+"\n", bech32get(ckey[0:]))
	}
	fmt.Fprintf(w, ` <ul><li>%X  %d.%08d COMB</li><li>%X %d.%08d COMB / %d.%08d COMB </li></ul>`+"\n", to, combs(balto), nats(balto), sumto, combs(balsumto), nats(balsumto), combs(sum), nats(sum))

	fmt.Fprintf(w, "<ul>\n")

	dump_liq_stack_item(w, to)
	dump_liq_stack_item(w, sumto)

	fmt.Fprintf(w, "</ul>\n")
}

func stacks_view(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, `<html><head></head><body>`)
	defer fmt.Fprintf(w, `</body></html>`)
	fmt.Fprintf(w, `<a href="/wallet/">&larr; Back to wallet</a><br />`)
	var seeds = make(map[[32]byte]struct{})
	var seeds_remove = make(map[[32]byte]struct{})

	segments_stack_mutex.RLock()

	for key, val := range segments_stack {
		var to, sumto, _ = stack_decode(val[0:])
		seeds[key] = struct{}{}

		seeds_remove[to] = struct{}{}
		seeds_remove[sumto] = struct{}{}
	}

	for key := range seeds_remove {
		delete(seeds, key)
	}
	seeds_remove = nil

	for key := range seeds {

		dump_liq_stack_item(w, key)

	}

	segments_stack_mutex.RUnlock()

	seeds = nil
	fmt.Fprintf(w, `
		<hr />
		To: <input id="to" style="font-family:monospace;width:45em" />
		To2: <input id="toamt" style="font-family:monospace;width:45em" />
		Amount2: <input id="amt" style="font-family:monospace;width:10em" />
		<a href="#" id="dostack">do liquidity stack</a>

		<script type="text/javascript">
		  var foo = function() {
		    var number = parseInt(document.getElementById("amt").value);
		    document.getElementById("dostack").href="/stack/data/"+
			document.getElementById("to").value.toUpperCase()+
			document.getElementById("toamt").value.toUpperCase()+""+number.toString(16).toUpperCase().padStart(16, "0");
		    return false;
		  };
		document.getElementById("to").oninput = foo;
		document.getElementById("toamt").oninput = foo;
		document.getElementById("amt").oninput = foo;
		document.getElementById("to").onpropertychange = foo;
		document.getElementById("toamt").onpropertychange = foo;
		document.getElementById("amt").onpropertychange = foo;
		</script>
		<hr />

	`)
	if wallet_selfmining_links {
		fmt.Fprintf(w, `
<a href="/mining/mine/FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF/9999999999999999">flush (finalize) block</a>
	`)
	}

}
