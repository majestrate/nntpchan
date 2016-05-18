var easiness = 55.0;
var miner_threads = 4;
var randoffs = 64;

onready(function(){
	document.getElementById("start_miner").onclick = function() {
		var btn = document.getElementById("start_miner");
		var label = btn.value;
		btn.value = "..."
		btn.disabled = true;
		
		
		var b = new Uint8Array(randoffs);
		window.crypto.getRandomValues(b);
		var b_cur = 0;
		var b_i = 0;
		var tmp = new Uint8Array(randoffs+b_i+1);
		tmp.set(b)
		tmp[b.length]=0
		b = tmp;
		
		var workers = new Array(miner_threads);
		
		var worker_cb =  function(e) {
			if (e.data[0] == "ok") {
			  miner_cb(e.data[1]);
			  btn.value=label;
			  btn.disabled = false;
			  for (i=0; i<miner_threads; i++) {
			  	workers[i].terminate();
			  }
			} else {
				if (b_cur >= 256) {
					var tmp = new Uint8Array(randoffs+b_i+1);
					tmp.set(b)
					tmp[b.length]=0
					b = tmp;
					b_i++;
					b_cur=0;
				}
				b[randoffs+b_i]=b_cur;
				b_cur++;
				var params = [b, easiness, e.data[2]];
				workers[e.data[2]].postMessage(params);
			}	
		
		}
		
		
		for (i=0; i<miner_threads; i++) {
			b[randoffs+b_i]=b_cur;
			b_cur++;
			var params = [b, easiness, i];
			workers[i] = new Worker("./static/mineworker.js");
			workers[i].onmessage = worker_cb;
			workers[i].postMessage(params); // Start the worker.
		}
		b_cur=4;
	};
});
function miner_cb(s) {
    document.getElementById("miner_result").value = s;
}
