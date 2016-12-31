

function neochan_post_new(j) {
    var post = document.createElement("div");
    post.id = j.PostHash;
    post.setAttribute("class", "neochan-post-wrapper");
    var header = document.createElement("div");
    header.setAttribute("class", "neochan-post-header");
    post.appendChild(header);

    var body = document.createElement("div");
    body.setAttribute("class", "neochan-post-body");

    neochan_postify(body, j.Message);

    return post;
}

function neochan_post_fadein(elem, j) {
    var post = neochan_post_new(j);
    $(post).fadein();
    elem.appendChild(post);
}
