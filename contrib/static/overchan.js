
function quickreply(shorthash, longhash, url) {
    if (!window.location.pathname.startsWith("/t/"))
    {
        window.location.href = url;
        return;
    }
    var elem = document.getElementById("comment");
    if(!elem) return;
    elem.value += ">>" + shorthash + "\n";
}
