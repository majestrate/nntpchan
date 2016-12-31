

/**
 * @brief initialize a board page
 */
function neochan_board_init(root, board) {

    var thread_init = function (j) {
        var elem = document.createElement("div");
        for(var idx = 0; idx < j.length; idx ++) {
            var post = j[idx];
            neochan_post_fadein(elem, post);
        }
        return elem;
    }
    // inject threads
    onready(function() {
        for (var idx = 0; idx < board.posts.length; idx ++) {
            var posts = board.posts[idx];
            var elem = thread_init(posts);
            root.appendChild(elem);
        }
    });
}
