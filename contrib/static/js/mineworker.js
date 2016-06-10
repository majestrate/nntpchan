importScripts('./siphash-lib.js');
importScripts('./miner-js.js');

onmessage = function(e) {
	var s = cuckoo["mine_cuckoo"](e.data[0], e.data[1]);
	s.push(e.data[2]);
	postMessage(s);
}
