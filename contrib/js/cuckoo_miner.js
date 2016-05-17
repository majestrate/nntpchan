onready(function(){
	document.getElementById("start_miner").onclick = function() {
		var worker = new Worker("./static/mineworker.js");
		worker.onmessage = function(e) {
		  miner_cb(e.data);
		  worker.terminate();
		}
		worker.postMessage(55.0); // Start the worker.
	};
});
function miner_cb(s) {
    document.getElementById("miner_result").value = s;
}
