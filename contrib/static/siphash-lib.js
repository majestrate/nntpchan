var SipHash = (function() {
	function u8to64(p , i) {
		return { l: (p[i] | p[i+1]<<8 | p[i+2]<<16 | p[i+3]<<24) , h: (p[i+4] | p[i+5]<<8 | p[i+6]<<16 | p[i+7]<<24) };
	}

	
    function _add(a, b) {
        var rl = a.l + b.l,
            a2 = { h: a.h + b.h + (rl / 2 >>> 31) >>> 0,
                   l: rl >>> 0 };
        a.h = a2.h; a.l = a2.l;
    }

    function _xor(a, b) {
        a.h ^= b.h; a.h >>>= 0;
        a.l ^= b.l; a.l >>>= 0;
    }

    function _rotl(a, n) {
        var a2 = {
            h: a.h << n | a.l >>> (32 - n),
            l: a.l << n | a.h >>> (32 - n)
        };
        a.h = a2.h; a.l = a2.l;
    }

    function _rotl32(a) {
        var al = a.l;
        a.l = a.h; a.h = al;
    }

    function _compress(v0, v1, v2, v3) {
        _add(v0, v1);
        _add(v2, v3);
        _rotl(v1, 13);
        _rotl(v3, 16);
        _xor(v1, v0);
        _xor(v3, v2);
        _rotl32(v0);
        _add(v2, v1);
        _add(v0, v3);
        _rotl(v1, 17);
        _rotl(v3, 21);
        _xor(v1, v2);
        _xor(v3, v0);
        _rotl32(v2);
    }


    function hash(v0h, v0l, v1h, v1l, v2h, v2l, v3h, v3l, mh, ml) {
    	var mi = { h: mh, l: ml };
    	
    	var v0 = { h: v0h, l: v0l };
    	var v1 = { h: v1h, l: v1l };
    	var v2 = { h: v2h, l: v2l };
    	var v3 = { h: v3h, l: v3l };
    	
        /*var k0 = u8to64(key, 0);
        var k1 = u8to64(key, 8);
        
        var v0 = { h: k0.h, l: k0.l }, v2 = k0;
        var v1 = { h: k1.h, l: k1.l }, v3 = k1;

        _xor(v0, { h: 0x736f6d65, l: 0x70736575 });
        _xor(v1, { h: 0x646f7261, l: 0x6e646f6d });
        _xor(v2, { h: 0x6c796765, l: 0x6e657261 });
        _xor(v3, { h: 0x74656462, l: 0x79746573 });*/
        
        _xor(v3, mi); 
        _compress(v0, v1, v2, v3);
        _compress(v0, v1, v2, v3);
        _xor(v0, mi);
        _xor(v2, { h: 0, l: 0xff }); 
        _compress(v0, v1, v2, v3);
        _compress(v0, v1, v2, v3);
        _compress(v0, v1, v2, v3);
        _compress(v0, v1, v2, v3);

        var h = v0;
        _xor(h, v1);
        _xor(h, v2);
        _xor(h, v3);
		
		var res = new Uint32Array(2);
		res[0]=h.h; res[1]=h.l;
        return res;
	};

    return {
        hash: hash,
    };
})();

var module = module || { }, exports = module.exports = SipHash;
