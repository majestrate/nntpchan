importScripts('./miner-js.js');

onmessage = function(e) {
	var s = cuckoo["mine_cuckoo"](e.data);
	postMessage(s);
}
