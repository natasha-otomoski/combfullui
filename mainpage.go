package main

import (
	"fmt"
	"net/http"
)

func main_page(w http.ResponseWriter, r *http.Request) {

	fmt.Fprintf(w, `
		<html><head>
		<title>Haircomb Core Wallet</title>

</head><body>
		<table style="width:100%%"><tr style="width:100%%"><td><font face="Arial,Helvetica,sans-serif"><small syle="letter-spacing: -0.5em;">|_|_|_|_|_|_|_|_|_|_|_|_|_|_|_|_|_|_|_|_|__________</small></font>
		<h1 style="padding-top:0; margin-top:0"> Haircomb Core</h1></td>

		<td><a href="/wallet/">&#x1f45b; Wallet</a></td>
		<td><a href="/import/">&#x1F4C1; Import</a></td>
		<td><a href="/utxo/">&#x1F5DE; View utxo</a></td>
<td> <a id="saver" href="/export/save/secret"> &#x1F4BE; Fully save as:</a> File name: <input id="filename" name="filename" value="secret" pattern="^[a-z0-9A-Z\.]*$" /> </td>

		<td><a href="/export/index.html">&#x1F320; Export coin history only</a></td>
<td> <a href="/shutdown"><svg style="width:1em;height:1em" version="1.2" baseProfile="tiny"
	 xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink" xmlns:a="http://ns.adobe.com/AdobeSVGViewerExtensions/3.0/"
	 x="0px" y="0px" width="177px" height="202px" viewBox="-0.8 -0.5 177 202" xml:space="preserve">
<defs>
</defs>
<path fill="none" stroke="#000000" stroke-width="30" stroke-linecap="round" d="M33.7,64.3C22.1,77.2,15,94.3,15,113
	c0,40.1,32.5,72.7,72.7,72.7c40.1,0,72.7-32.5,72.7-72.7c0-18.7-7.1-35.8-18.7-48.7"/>
<line fill="none" stroke="#000000" stroke-width="30" stroke-linecap="round" x1="87.8" y1="15" x2="87.8" y2="113"/>
</svg> Safe shutdown</a> </td> </tr> </table>

		<table style="width:20%%"><tr>
		<td> <a href="/purse/">&#x1F511; Deciders purse</a></td>
		<td> <a href="/basiccontract/">&#129309; Basic contract</a></td>
		</tr> </table>

		<script type="text/javascript">
		  var foo = function() {
			var val = document.getElementById("filename").value.replace(/\//g, "");

			if ((val == "..") || (val == "")) {
				val = "secret";
			}

		    document.getElementById("saver").href="/export/save/"+
			val; 
		    return false;
		  };
		document.getElementById("filename").oninput = foo;
		document.getElementById("filename").onpropertychange = foo;

		</script>
`)
	defer fmt.Fprintf(w, `

		</body></html>
	`)

	var sum uint64

	commits_mutex.RLock()
	balance_mutex.RLock()

	for _, val := range balance_node {
		sum += val
	}

	var size = len(commits)
	var height = uint64(commit_currently_loaded.height)
	var cryptofingerprint = database_aes

	balance_mutex.RUnlock()
	commits_mutex.RUnlock()

	var sum2 = RemainingAfter(height)

	var fakeheight = makefaketag().height
	if height < uint64(fakeheight) {
		fmt.Fprintf(w, `<h1>Wallet appears to be out of sync. Displayed balances are incorrect until wallet is fully synced:</h1>`)
	}

	fmt.Fprintf(w, "<p>%x Fingerprint of commitment set of size %d</p>", cryptofingerprint, size)
	fmt.Fprintf(w, "<p>%d.%08d COMB Tokens in existence as of Bitcoin block %d </p>", combs(sum), nats(sum), height)
	fmt.Fprintf(w, "<p>%d.%08d COMB Tokens remaining to be mined after Bitcoin block %d </p>", combs(sum2), nats(sum2), height)
	fmt.Fprintf(w, "<p><a href=\"/utxo/bisect/b%08d\">&#128272; Validate database integrity with peer</a></p>", height&0xFFFFFC00)
}

func shutdown_page(w http.ResponseWriter, r *http.Request) {

	var wallet_flushed_ok bool

	wallet_mutex.RLock()

	wallet_flushed_ok = len(wallet) == wallet_saved
	wallet_mutex.RUnlock()

	if !wallet_flushed_ok {

		fmt.Fprintf(w, `
			<html><head></head><body><a href="/">Back</a><h1>Wallet is not saved! You may not shutdown!</h1></body></html>
		`)
		return
	}

	fmt.Fprintf(w, `
		<html><head></head><body>Shutdown successfull. You may now close the window.</body></html>
	`)
	go CommitDbCtrlC(fmt.Errorf("SHUTDOWN SIGNAL"))
}

func version_page(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Disposition", `inline; filename="version.json"`)
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"version_major":"`+string((version_major/10%10)+'0')+string((version_major%10)+'0')+
		`","version_minor":"`+string((version_minor/10%10)+'0')+string((version_minor%10)+'0')+
		`","version_patch":"`+string((version_patch/10%10)+'0')+string((version_patch%10)+'0')+`"}`)

}
