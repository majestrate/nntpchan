
function inject_imghide(elem) {
    elem.onerror = function() {
        $(elem).fadeOut(200, function() {
            elem.remove();
        });
    }
}

onready(function(){
    var imgs = document.getElementsByClassName("thumbnail");
    for (var idx = 0; idx < imgs.length; idx ++) {
        var elem = imgs[idx];
        inject_imghide(elem);
    }
});
