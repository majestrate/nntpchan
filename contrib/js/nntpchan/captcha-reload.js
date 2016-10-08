function setSrcQuery(e, q) {
	var src  = e.src;
	var p = src.indexOf('?');
	if (p >= 0) {
		src = src.substr(0, p);
	}
	e.src = src + "?" + q
}


function reload(el) {
	setSrcQuery(el, "reload=" + (new Date()).getTime());
	return false;
}

onready(function(){
	document.getElementById("captcha_img").onclick = function() {
		reload(document.getElementById("captcha_img"));
	};
});
