/** hooks.js */

onready_callbacks = [];
function onready(fnc) {
	onready_callbacks.push(fnc);
}

function ready() {
	for (var i = 0; i < onready_callbacks.length; i++) {
		onready_callbacks[i]();
	}
}

function forEachInClass(clazz, cb) {
  var elems = document.getElementsByClassName(clazz);
  for (var idx = 0; idx < elems.length; idx ++) {
    cb(elems[idx]);
  }
}
