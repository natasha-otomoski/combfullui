package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
)

func prettyprint_liq_stack_items(w http.ResponseWriter, key [32]byte, until *[32]byte) (bool, uint64, int) {
	var val, ok = segments_stack[key]

	if !ok {
		return false, 0, 0
	}

	var to, sumto, sum = stack_decode(val[0:])

	//fmt.Fprintf(w, `<h2>%X COMB</h2>`+"\n", key)

	fmt.Fprintf(w, ` <ul><li>Pay to %X: %d.%08d COMB </li></ul>`+"\n", sumto, combs(sum), nats(sum))

	if to == *until {
		return true, sum, 1
	}

	var goodstack, totalnats, totalentries = prettyprint_liq_stack_items(w, to, until)

	return goodstack, totalnats + sum, totalentries + 1
}

func stack_load_multipay_data(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	var walletdata, changedata, data = vars["wallet"], vars["change"], vars["data"]

	fmt.Fprintf(w, `<html><head></head><body>`)
	defer fmt.Fprintf(w, `</body></html>`)

	err1 := checkHEX32(walletdata)
	if err1 != nil {
		fmt.Fprintf(w, "error stack building from walletdata: %s", err1)
		return
	}
	err2 := checkHEX32(changedata)
	if err2 != nil {
		fmt.Fprintf(w, "error stack building targeting changedata: %s", err2)
		return
	}

	var rawwalletdata = hex2byte32([]byte(walletdata))
	var rawchangedata = hex2byte32([]byte(changedata))

	var stackhash = stack_load_data_internal(w, data)
	fmt.Fprintf(w, `<a href="/sign/multipay/%X/%X/%X">&larr; Back to your multipayment</a><br />`, rawwalletdata, rawchangedata, stackhash)
}

func stackbuilder(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, `<html><head></head><body onload="javascript:bar();"><a href="/wallet/">&larr; Back to wallet</a><br />`)
	defer fmt.Fprintf(w, `</body></html>`)

	vars := mux.Vars(r)
	var walletkey, change, stackbottom = vars["walletkey"], vars["change"], vars["stackbottom"]

	_ = walletkey

	err1 := checkHEX32(walletkey)
	if err1 != nil {
		fmt.Fprintf(w, "error stack building from wallet: %s", err1)
		return
	}
	err2 := checkHEX32(change)
	if err2 != nil {
		fmt.Fprintf(w, "error stack building targeting change address: %s", err2)
		return
	}
	err3 := checkHEX32(stackbottom)
	if err3 != nil {
		fmt.Fprintf(w, "error stack building with stack bottom: %s", err3)
		return
	}

	var rawwalletkey = hex2byte32([]byte(walletkey))
	var rawchange = hex2byte32([]byte(change))
	var rawstackbottom = hex2byte32([]byte(stackbottom))
	// print entries below stackbottom until change is reached
	fmt.Fprintf(w, "<h1>From address %X</h1>\n", rawwalletkey)

	fmt.Fprintf(w, `
		<hr />
		To: <input id="to" style="font-family:monospace;width:45em" />
		Amount: <input id="amt" style="font-family:monospace;width:10em" />
		<a href="#" id="dostack">Add destination</a>
		<a href="#" id="cleardest" onclick="javascript:bar();">Clear destination</a>
		
		<script type="text/javascript">
		  var bar = function() {
		    document.getElementById("amt").value = "";
		    document.getElementById("to").value = "";
		  }
		
		  var foo = function() {
		    var len = (document.getElementById("amt").value.length)+(document.getElementById("to").value.length);
		    if (len == 0) {
		        document.getElementById("activator").style.visibility = 'visible';
		        document.getElementById("advice").style.visibility = 'hidden';
		    } else {
		        document.getElementById("activator").style.visibility = 'hidden';
		        document.getElementById("advice").style.visibility = 'visible';
		    }
		  
		    var number = parseInt(document.getElementById("amt").value);
		    document.getElementById("dostack").href="/stack/multipaydata/%X/%X/%X"+
			document.getElementById("to").value.toUpperCase()+
			""+number.toString(16).toUpperCase().padStart(16, "0");
		    return false;
		  };
		document.getElementById("to").oninput = foo;
		document.getElementById("amt").oninput = foo;
		document.getElementById("to").onpropertychange = foo;
		document.getElementById("amt").onpropertychange = foo;
		</script>
		<hr />

	`, rawwalletkey, rawchange, rawstackbottom)

	var goodstack, totalnats, totalentries = prettyprint_liq_stack_items(w, rawstackbottom, &rawchange)

	if !goodstack {
		if rawchange != rawstackbottom {
			fmt.Fprintf(w, `<h1>THE BOTTOM OF THE STACK IS NOT YOUR CHANGE ADDRESS. IT IS NOT SAFE TO CONTINUE!</h1>`)
			return
		}
	}
	fmt.Fprintf(w, "<h1>%d.%08d COMB will be spent to %d destination(s).</h1>\n", combs(totalnats), nats(totalnats), totalentries)
	fmt.Fprintf(w, "<p>The remainder will go to the change address <small><tt>%X</tt></small></p>", rawchange)

	fmt.Fprintf(w, "<a href=\"/stack/multipaydata/%X/%X/%X%X%016X\">Fold the stack</a>\n", rawwalletkey, rawchange, rawchange, rawstackbottom, totalnats)
	fmt.Fprintf(w, "<a id=\"activator\" href=\"/sign/pay/%X/%X\">Do payment</a>\n", rawwalletkey, rawstackbottom)
	fmt.Fprintf(w, `<span id="advice" style="visibility:hidden">Please add or clear destination to complete your payment</span>`)
}
