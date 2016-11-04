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
  var e = document.getElementById("captcha_img");
  if (e) {
    e.onclick = function() {
		  reload(e);
	  };
  }
});
