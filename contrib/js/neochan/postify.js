

function _neochan_filter_boardlink(match) {
    match = match.toLowerCase();
    var a = document.createElement("a");
    a.href = "/" + match + "-0.html";
    match = ">>>/" + match + "/";
    a.appendChild(document.createTextNode(match));
    return a;
}

function _neochan_filter_postlink(match) {
    
}

var _neochan_post_filters = [
    [/>>>\/(overchan\\.[a-zA-z0-9\\.]+[a-zA-Z0-9])\//g, _neochan_filter_boardlink],
    [/>>? ([a-fA-F0-9])/g, _neochan_filter_postlink],
    [/==(.+)==/g, _neochan_filter_redtext],
    [/@@(.+)@@/g, _neochan_filter_psytext],
    [/^>/g, _neochan_filter_greentext],
];

/**
 * @brief create post body from raw text
 */
function neochan_postify(elem, text) {
    $.each(_neochan_post_filters, function(idx, ent) {
        var re = ent[0];
        var func = ent[1];
        text = text.replace(re, function(m) {
            var e = func(m);
            
            return "";
        });
    });
}
