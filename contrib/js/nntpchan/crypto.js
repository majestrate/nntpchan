
var nntpchan_ToBase64 = function (u8) {
    return btoa(String.fromCharCode.apply(null, u8));
};

var nntpchan_FromBase64 = function (str) {
    return atob(str).split('').map(function (c) { return c.charCodeAt(0); });
};


function nntpchan_keygen() {
    var crypto = window.crypto || window.msCrypto;
    if(!crypto) {
        throw "no crypto";
    }
    var key = new Uint8Array(32);
    crypto.getRandomValues(key);
    var nonce = new Uint8Array(12);
    crypto.getRandomValues(nonce);
    return nntpchan_ToBase64(key) + ":" + nntpchan_ToBase64(nonce);
}

function nntpchan_chacha_nonce(blob) {
    var parts = blob.split(":");
    return nntpchan_FromBase64(parts[1]);
}

function nntpchan_chacha_key(blob) {
    var parts = blob.split(":");
    return nntpchan_FromBase64(parts[0]);
}

// encrypt text from an element using symettric key and nonce
// output to element.innerHTML as base64
function nntpchan_encrypt_element(inelem, outelem, key, nonce) {
    var plaintext = inelem.innerHTML;
    console.log(plaintext);
    var ciphertext = chacha_encrypt(key, nonce, plaintext);
    outelem.innerHTML = nntpchan_ToBase64(ciphertext);
}

// encrypt text from an element using symettric key and nonce
// output to element.innerHTML as base64
function nntpchan_decrypt_element(inelem, outelem, key, nonce) {
    var ciphertext = nntpchan_FromBase64(inelem.innerHTML);
    console.log(ciphertext);
    var plaintext = chacha_decrypt(key, nonce, ciphertext);
    console.log(plaintext);
    while(outelem.children.length > 0) {
        outelem.children[0].remove();
    }
    outelem.appendChild(document.createTextNode(plaintext));
}



function nntpchan_crypto_test() {
    console.log("begin crypto test");
    var genelem = $("#test_keygen");
    genelem[0].innerHTML = nntpchan_keygen();
    var inelem = $("#encrypt_in")[0];
    var outelem = $("#encrypt_out")[0];
    var decryptelem = $("#decrypt_out")[0];
    var key = nntpchan_chacha_key(genelem[0].innerHTML);
    var nonce = nntpchan_chacha_nonce(genelem[0].innerHTML);
    nntpchan_encrypt_element(inelem, outelem, key, nonce);
    nntpchan_decrypt_element(outelem, decryptelem, key, nonce);
}
