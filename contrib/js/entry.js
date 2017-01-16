onready_callbacks = [];
function onready(fnc) {
	onready_callbacks.push(fnc);
}

function ready(prefix) {
  configRoot = prefix || "/";
	for (var i = 0; i < onready_callbacks.length; i++) {
		onready_callbacks[i]();
	}
}

