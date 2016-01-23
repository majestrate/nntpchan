var http = require('http');

req = {
  message: "test api",
  frontend: "benis.tld",
  name: "benisname",
  subject: "ayyyyyy testing api",
  /*
    file: {
      name: "benis.gif",
      type: "image/gif",
      data: // base64'd string here
    },
  */
  email: "sage",
  ip: "8.8.8.8",
  dubs: false,
  newsgroup: "overchan.test",

  // only include if we are replying to someone
  reference: "<b7dee1453564515@benis.tld>"
}

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
