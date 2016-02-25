var http = require('http');

var makeIpBans = function(cidrs, privkey, cb) {
  cb({
    message: cidrs.join("\noverchan-inet-ban "),
    name: "mod#"+privkey,
    frontend: "memegod.censor",
    newsgroup: "ctl"
  })
}

var makeDeletePosts = function(msgids, privkey, cb) {
  cb({
    message: msgids.join("\ndelete "),
    name: "mod#"+privkey,
    frontend: "memegod.censor",
    newsgroup: "ctl",
  })
}

var moderate = function(req) {

  j = JSON.stringify(req);

  var r = http.request({
    port: 8800,
    method: "POST",
    path: "/api/post",
    auth: "user:pass",
    headers: {
      "Content-Type": "text/json",
      "Content-Length": j.length
    }
  }, function (res) {
    res.on('data', function (chunk) {
      var r = chunk.toString();
      var rj = JSON.parse(r);
      console.log(rj.id);
    });
  });

  r.write(j);
  r.end();
}

var privateKey = "longhexgoestripcodegoeshere";

// ban 192.168.0.1/16 and sign with private key
makeIpBans(["192.168.0.1/16"], privateKey, moderate);
// delete <msg1@place.tld> and <msg2@otherplace.tld> and sign with private key
makeDeletPosts(["<msg1@place.tld>", "<msg2@otherplace.tld>"], privateKey, moderate);
