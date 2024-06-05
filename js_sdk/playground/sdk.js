var Rimdian = (function () {

  function _typeof(obj) {
    "@babel/helpers - typeof";

    return _typeof = "function" == typeof Symbol && "symbol" == typeof Symbol.iterator ? function (obj) {
      return typeof obj;
    } : function (obj) {
      return obj && "function" == typeof Symbol && obj.constructor === Symbol && obj !== Symbol.prototype ? "symbol" : typeof obj;
    }, _typeof(obj);
  }

  /******************************************************************************
  Copyright (c) Microsoft Corporation.

  Permission to use, copy, modify, and/or distribute this software for any
  purpose with or without fee is hereby granted.

  THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
  REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
  AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
  INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
  LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
  OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
  PERFORMANCE OF THIS SOFTWARE.
  ***************************************************************************** */

  var __assign = function() {
      __assign = Object.assign || function __assign(t) {
          for (var s, i = 1, n = arguments.length; i < n; i++) {
              s = arguments[i];
              for (var p in s) if (Object.prototype.hasOwnProperty.call(s, p)) t[p] = s[p];
          }
          return t;
      };
      return __assign.apply(this, arguments);
  };

  var commonjsGlobal = typeof globalThis !== 'undefined' ? globalThis : typeof window !== 'undefined' ? window : typeof global !== 'undefined' ? global : typeof self !== 'undefined' ? self : {};

  var md5 = {exports: {}};

  /*
   * JavaScript MD5
   * https://github.com/blueimp/JavaScript-MD5
   *
   * Copyright 2011, Sebastian Tschan
   * https://blueimp.net
   *
   * Licensed under the MIT license:
   * https://opensource.org/licenses/MIT
   *
   * Based on
   * A JavaScript implementation of the RSA Data Security, Inc. MD5 Message
   * Digest Algorithm, as defined in RFC 1321.
   * Version 2.2 Copyright (C) Paul Johnston 1999 - 2009
   * Other contributors: Greg Holt, Andrew Kepert, Ydnar, Lostinet
   * Distributed under the BSD License
   * See http://pajhome.org.uk/crypt/md5 for more info.
   */

  (function (module) {
  (function ($) {

  	  /**
  	   * Add integers, wrapping at 2^32.
  	   * This uses 16-bit operations internally to work around bugs in interpreters.
  	   *
  	   * @param {number} x First integer
  	   * @param {number} y Second integer
  	   * @returns {number} Sum
  	   */
  	  function safeAdd(x, y) {
  	    var lsw = (x & 0xffff) + (y & 0xffff);
  	    var msw = (x >> 16) + (y >> 16) + (lsw >> 16);
  	    return (msw << 16) | (lsw & 0xffff)
  	  }

  	  /**
  	   * Bitwise rotate a 32-bit number to the left.
  	   *
  	   * @param {number} num 32-bit number
  	   * @param {number} cnt Rotation count
  	   * @returns {number} Rotated number
  	   */
  	  function bitRotateLeft(num, cnt) {
  	    return (num << cnt) | (num >>> (32 - cnt))
  	  }

  	  /**
  	   * Basic operation the algorithm uses.
  	   *
  	   * @param {number} q q
  	   * @param {number} a a
  	   * @param {number} b b
  	   * @param {number} x x
  	   * @param {number} s s
  	   * @param {number} t t
  	   * @returns {number} Result
  	   */
  	  function md5cmn(q, a, b, x, s, t) {
  	    return safeAdd(bitRotateLeft(safeAdd(safeAdd(a, q), safeAdd(x, t)), s), b)
  	  }
  	  /**
  	   * Basic operation the algorithm uses.
  	   *
  	   * @param {number} a a
  	   * @param {number} b b
  	   * @param {number} c c
  	   * @param {number} d d
  	   * @param {number} x x
  	   * @param {number} s s
  	   * @param {number} t t
  	   * @returns {number} Result
  	   */
  	  function md5ff(a, b, c, d, x, s, t) {
  	    return md5cmn((b & c) | (~b & d), a, b, x, s, t)
  	  }
  	  /**
  	   * Basic operation the algorithm uses.
  	   *
  	   * @param {number} a a
  	   * @param {number} b b
  	   * @param {number} c c
  	   * @param {number} d d
  	   * @param {number} x x
  	   * @param {number} s s
  	   * @param {number} t t
  	   * @returns {number} Result
  	   */
  	  function md5gg(a, b, c, d, x, s, t) {
  	    return md5cmn((b & d) | (c & ~d), a, b, x, s, t)
  	  }
  	  /**
  	   * Basic operation the algorithm uses.
  	   *
  	   * @param {number} a a
  	   * @param {number} b b
  	   * @param {number} c c
  	   * @param {number} d d
  	   * @param {number} x x
  	   * @param {number} s s
  	   * @param {number} t t
  	   * @returns {number} Result
  	   */
  	  function md5hh(a, b, c, d, x, s, t) {
  	    return md5cmn(b ^ c ^ d, a, b, x, s, t)
  	  }
  	  /**
  	   * Basic operation the algorithm uses.
  	   *
  	   * @param {number} a a
  	   * @param {number} b b
  	   * @param {number} c c
  	   * @param {number} d d
  	   * @param {number} x x
  	   * @param {number} s s
  	   * @param {number} t t
  	   * @returns {number} Result
  	   */
  	  function md5ii(a, b, c, d, x, s, t) {
  	    return md5cmn(c ^ (b | ~d), a, b, x, s, t)
  	  }

  	  /**
  	   * Calculate the MD5 of an array of little-endian words, and a bit length.
  	   *
  	   * @param {Array} x Array of little-endian words
  	   * @param {number} len Bit length
  	   * @returns {Array<number>} MD5 Array
  	   */
  	  function binlMD5(x, len) {
  	    /* append padding */
  	    x[len >> 5] |= 0x80 << len % 32;
  	    x[(((len + 64) >>> 9) << 4) + 14] = len;

  	    var i;
  	    var olda;
  	    var oldb;
  	    var oldc;
  	    var oldd;
  	    var a = 1732584193;
  	    var b = -271733879;
  	    var c = -1732584194;
  	    var d = 271733878;

  	    for (i = 0; i < x.length; i += 16) {
  	      olda = a;
  	      oldb = b;
  	      oldc = c;
  	      oldd = d;

  	      a = md5ff(a, b, c, d, x[i], 7, -680876936);
  	      d = md5ff(d, a, b, c, x[i + 1], 12, -389564586);
  	      c = md5ff(c, d, a, b, x[i + 2], 17, 606105819);
  	      b = md5ff(b, c, d, a, x[i + 3], 22, -1044525330);
  	      a = md5ff(a, b, c, d, x[i + 4], 7, -176418897);
  	      d = md5ff(d, a, b, c, x[i + 5], 12, 1200080426);
  	      c = md5ff(c, d, a, b, x[i + 6], 17, -1473231341);
  	      b = md5ff(b, c, d, a, x[i + 7], 22, -45705983);
  	      a = md5ff(a, b, c, d, x[i + 8], 7, 1770035416);
  	      d = md5ff(d, a, b, c, x[i + 9], 12, -1958414417);
  	      c = md5ff(c, d, a, b, x[i + 10], 17, -42063);
  	      b = md5ff(b, c, d, a, x[i + 11], 22, -1990404162);
  	      a = md5ff(a, b, c, d, x[i + 12], 7, 1804603682);
  	      d = md5ff(d, a, b, c, x[i + 13], 12, -40341101);
  	      c = md5ff(c, d, a, b, x[i + 14], 17, -1502002290);
  	      b = md5ff(b, c, d, a, x[i + 15], 22, 1236535329);

  	      a = md5gg(a, b, c, d, x[i + 1], 5, -165796510);
  	      d = md5gg(d, a, b, c, x[i + 6], 9, -1069501632);
  	      c = md5gg(c, d, a, b, x[i + 11], 14, 643717713);
  	      b = md5gg(b, c, d, a, x[i], 20, -373897302);
  	      a = md5gg(a, b, c, d, x[i + 5], 5, -701558691);
  	      d = md5gg(d, a, b, c, x[i + 10], 9, 38016083);
  	      c = md5gg(c, d, a, b, x[i + 15], 14, -660478335);
  	      b = md5gg(b, c, d, a, x[i + 4], 20, -405537848);
  	      a = md5gg(a, b, c, d, x[i + 9], 5, 568446438);
  	      d = md5gg(d, a, b, c, x[i + 14], 9, -1019803690);
  	      c = md5gg(c, d, a, b, x[i + 3], 14, -187363961);
  	      b = md5gg(b, c, d, a, x[i + 8], 20, 1163531501);
  	      a = md5gg(a, b, c, d, x[i + 13], 5, -1444681467);
  	      d = md5gg(d, a, b, c, x[i + 2], 9, -51403784);
  	      c = md5gg(c, d, a, b, x[i + 7], 14, 1735328473);
  	      b = md5gg(b, c, d, a, x[i + 12], 20, -1926607734);

  	      a = md5hh(a, b, c, d, x[i + 5], 4, -378558);
  	      d = md5hh(d, a, b, c, x[i + 8], 11, -2022574463);
  	      c = md5hh(c, d, a, b, x[i + 11], 16, 1839030562);
  	      b = md5hh(b, c, d, a, x[i + 14], 23, -35309556);
  	      a = md5hh(a, b, c, d, x[i + 1], 4, -1530992060);
  	      d = md5hh(d, a, b, c, x[i + 4], 11, 1272893353);
  	      c = md5hh(c, d, a, b, x[i + 7], 16, -155497632);
  	      b = md5hh(b, c, d, a, x[i + 10], 23, -1094730640);
  	      a = md5hh(a, b, c, d, x[i + 13], 4, 681279174);
  	      d = md5hh(d, a, b, c, x[i], 11, -358537222);
  	      c = md5hh(c, d, a, b, x[i + 3], 16, -722521979);
  	      b = md5hh(b, c, d, a, x[i + 6], 23, 76029189);
  	      a = md5hh(a, b, c, d, x[i + 9], 4, -640364487);
  	      d = md5hh(d, a, b, c, x[i + 12], 11, -421815835);
  	      c = md5hh(c, d, a, b, x[i + 15], 16, 530742520);
  	      b = md5hh(b, c, d, a, x[i + 2], 23, -995338651);

  	      a = md5ii(a, b, c, d, x[i], 6, -198630844);
  	      d = md5ii(d, a, b, c, x[i + 7], 10, 1126891415);
  	      c = md5ii(c, d, a, b, x[i + 14], 15, -1416354905);
  	      b = md5ii(b, c, d, a, x[i + 5], 21, -57434055);
  	      a = md5ii(a, b, c, d, x[i + 12], 6, 1700485571);
  	      d = md5ii(d, a, b, c, x[i + 3], 10, -1894986606);
  	      c = md5ii(c, d, a, b, x[i + 10], 15, -1051523);
  	      b = md5ii(b, c, d, a, x[i + 1], 21, -2054922799);
  	      a = md5ii(a, b, c, d, x[i + 8], 6, 1873313359);
  	      d = md5ii(d, a, b, c, x[i + 15], 10, -30611744);
  	      c = md5ii(c, d, a, b, x[i + 6], 15, -1560198380);
  	      b = md5ii(b, c, d, a, x[i + 13], 21, 1309151649);
  	      a = md5ii(a, b, c, d, x[i + 4], 6, -145523070);
  	      d = md5ii(d, a, b, c, x[i + 11], 10, -1120210379);
  	      c = md5ii(c, d, a, b, x[i + 2], 15, 718787259);
  	      b = md5ii(b, c, d, a, x[i + 9], 21, -343485551);

  	      a = safeAdd(a, olda);
  	      b = safeAdd(b, oldb);
  	      c = safeAdd(c, oldc);
  	      d = safeAdd(d, oldd);
  	    }
  	    return [a, b, c, d]
  	  }

  	  /**
  	   * Convert an array of little-endian words to a string
  	   *
  	   * @param {Array<number>} input MD5 Array
  	   * @returns {string} MD5 string
  	   */
  	  function binl2rstr(input) {
  	    var i;
  	    var output = '';
  	    var length32 = input.length * 32;
  	    for (i = 0; i < length32; i += 8) {
  	      output += String.fromCharCode((input[i >> 5] >>> i % 32) & 0xff);
  	    }
  	    return output
  	  }

  	  /**
  	   * Convert a raw string to an array of little-endian words
  	   * Characters >255 have their high-byte silently ignored.
  	   *
  	   * @param {string} input Raw input string
  	   * @returns {Array<number>} Array of little-endian words
  	   */
  	  function rstr2binl(input) {
  	    var i;
  	    var output = [];
  	    output[(input.length >> 2) - 1] = undefined;
  	    for (i = 0; i < output.length; i += 1) {
  	      output[i] = 0;
  	    }
  	    var length8 = input.length * 8;
  	    for (i = 0; i < length8; i += 8) {
  	      output[i >> 5] |= (input.charCodeAt(i / 8) & 0xff) << i % 32;
  	    }
  	    return output
  	  }

  	  /**
  	   * Calculate the MD5 of a raw string
  	   *
  	   * @param {string} s Input string
  	   * @returns {string} Raw MD5 string
  	   */
  	  function rstrMD5(s) {
  	    return binl2rstr(binlMD5(rstr2binl(s), s.length * 8))
  	  }

  	  /**
  	   * Calculates the HMAC-MD5 of a key and some data (raw strings)
  	   *
  	   * @param {string} key HMAC key
  	   * @param {string} data Raw input string
  	   * @returns {string} Raw MD5 string
  	   */
  	  function rstrHMACMD5(key, data) {
  	    var i;
  	    var bkey = rstr2binl(key);
  	    var ipad = [];
  	    var opad = [];
  	    var hash;
  	    ipad[15] = opad[15] = undefined;
  	    if (bkey.length > 16) {
  	      bkey = binlMD5(bkey, key.length * 8);
  	    }
  	    for (i = 0; i < 16; i += 1) {
  	      ipad[i] = bkey[i] ^ 0x36363636;
  	      opad[i] = bkey[i] ^ 0x5c5c5c5c;
  	    }
  	    hash = binlMD5(ipad.concat(rstr2binl(data)), 512 + data.length * 8);
  	    return binl2rstr(binlMD5(opad.concat(hash), 512 + 128))
  	  }

  	  /**
  	   * Convert a raw string to a hex string
  	   *
  	   * @param {string} input Raw input string
  	   * @returns {string} Hex encoded string
  	   */
  	  function rstr2hex(input) {
  	    var hexTab = '0123456789abcdef';
  	    var output = '';
  	    var x;
  	    var i;
  	    for (i = 0; i < input.length; i += 1) {
  	      x = input.charCodeAt(i);
  	      output += hexTab.charAt((x >>> 4) & 0x0f) + hexTab.charAt(x & 0x0f);
  	    }
  	    return output
  	  }

  	  /**
  	   * Encode a string as UTF-8
  	   *
  	   * @param {string} input Input string
  	   * @returns {string} UTF8 string
  	   */
  	  function str2rstrUTF8(input) {
  	    return unescape(encodeURIComponent(input))
  	  }

  	  /**
  	   * Encodes input string as raw MD5 string
  	   *
  	   * @param {string} s Input string
  	   * @returns {string} Raw MD5 string
  	   */
  	  function rawMD5(s) {
  	    return rstrMD5(str2rstrUTF8(s))
  	  }
  	  /**
  	   * Encodes input string as Hex encoded string
  	   *
  	   * @param {string} s Input string
  	   * @returns {string} Hex encoded string
  	   */
  	  function hexMD5(s) {
  	    return rstr2hex(rawMD5(s))
  	  }
  	  /**
  	   * Calculates the raw HMAC-MD5 for the given key and data
  	   *
  	   * @param {string} k HMAC key
  	   * @param {string} d Input string
  	   * @returns {string} Raw MD5 string
  	   */
  	  function rawHMACMD5(k, d) {
  	    return rstrHMACMD5(str2rstrUTF8(k), str2rstrUTF8(d))
  	  }
  	  /**
  	   * Calculates the Hex encoded HMAC-MD5 for the given key and data
  	   *
  	   * @param {string} k HMAC key
  	   * @param {string} d Input string
  	   * @returns {string} Raw MD5 string
  	   */
  	  function hexHMACMD5(k, d) {
  	    return rstr2hex(rawHMACMD5(k, d))
  	  }

  	  /**
  	   * Calculates MD5 value for a given string.
  	   * If a key is provided, calculates the HMAC-MD5 value.
  	   * Returns a Hex encoded string unless the raw argument is given.
  	   *
  	   * @param {string} string Input string
  	   * @param {string} [key] HMAC key
  	   * @param {boolean} [raw] Raw output switch
  	   * @returns {string} MD5 output
  	   */
  	  function md5(string, key, raw) {
  	    if (!key) {
  	      if (!raw) {
  	        return hexMD5(string)
  	      }
  	      return rawMD5(string)
  	    }
  	    if (!raw) {
  	      return hexHMACMD5(key, string)
  	    }
  	    return rawHMACMD5(key, string)
  	  }

  	  if (module.exports) {
  	    module.exports = md5;
  	  } else {
  	    $.md5 = md5;
  	  }
  	})(commonjsGlobal);
  } (md5));

  var _md = md5.exports;

  /*!
   Copyright 2018 Google Inc. All Rights Reserved.
   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
  */
  /*! lifecycle.mjs v0.1.1 */
  let e;try{new EventTarget,e=!0;}catch(t){e=!1;}class t{constructor(){this.e={};}addEventListener(e,t,s=!1){this.t(e).push(t);}removeEventListener(e,t,s=!1){const i=this.t(e),a=i.indexOf(t);a>-1&&i.splice(a,1);}dispatchEvent(e){return e.target=this,Object.freeze(e),this.t(e.type).forEach(t=>t(e)),!0}t(e){return this.e[e]=this.e[e]||[]}}var s=e?EventTarget:t;class i{constructor(e){this.type=e;}}var a=e?Event:i;class n extends a{constructor(e,t){super(e),this.newState=t.newState,this.oldState=t.oldState,this.originalEvent=t.originalEvent;}}const r="active",h="passive",c="hidden",o="frozen",d="terminated",u="object"==typeof safari&&safari.pushNotification,v="onpageshow"in self,l=["focus","blur","visibilitychange","freeze","resume","pageshow",v?"pagehide":"unload"],g=e=>(e.preventDefault(),e.returnValue="Are you sure?"),f=e=>e.reduce((e,t,s)=>(e[t]=s,e),{}),b=[[r,h,c,d],[r,h,c,o],[c,h,r],[o,c],[o,r],[o,h]].map(f),p=(e,t)=>{for(let s,i=0;s=b[i];++i){const i=s[e],a=s[t];if(i>=0&&a>=0&&a>i)return Object.keys(s).slice(i,a+1)}return []},E=()=>document.visibilityState===c?c:document.hasFocus()?r:h;class m extends s{constructor(){super();const e=E();this.s=e,this.i=[],this.a=this.a.bind(this),l.forEach(e=>addEventListener(e,this.a,!0)),u&&addEventListener("beforeunload",e=>{this.n=setTimeout(()=>{e.defaultPrevented||e.returnValue.length>0||this.r(e,c);},0);});}get state(){return this.s}get pageWasDiscarded(){return document.wasDiscarded||!1}addUnsavedChanges(e){!this.i.indexOf(e)>-1&&(0===this.i.length&&addEventListener("beforeunload",g),this.i.push(e));}removeUnsavedChanges(e){const t=this.i.indexOf(e);t>-1&&(this.i.splice(t,1),0===this.i.length&&removeEventListener("beforeunload",g));}r(e,t){if(t!==this.s){const s=this.s,i=p(s,t);for(let t=0;t<i.length-1;++t){const s=i[t],a=i[t+1];this.s=a,this.dispatchEvent(new n("statechange",{oldState:s,newState:a,originalEvent:e}));}}}a(e){switch(u&&clearTimeout(this.n),e.type){case"pageshow":case"resume":this.r(e,E());break;case"focus":this.r(e,r);break;case"blur":this.s===r&&this.r(e,E());break;case"pagehide":case"unload":this.r(e,e.persisted?o:d);break;case"visibilitychange":this.s!==o&&this.s!==d&&this.r(e,E());break;case"freeze":this.r(e,o);}}}var w=new m;

  // params used to decorate URLs for cross-device / cross-domain linking

  var URLParams = {
    device_external_id: '_did',
    user_external_id: '_uid',
    user_is_authenticated: '_auth',
    user_external_id_hmac: '_uidh'
  };
  var OneYearInSeconds = 31536000;
  var PageStates = {
    active: 'active',
    passive: 'passive',
    hidden: 'hidden',
    frozen: 'frozen'
  };
  var Rimdian = {
    config: {
      workspace_id: '',
      host: 'https://collector-eu.rimdian.com',
      session_timeout: 60 * 30,
      namespace: '_rmd_',
      cross_domains: [],
      ignored_origins: [],
      version: '2.6.0',
      log_level: 'error',
      max_retry: 10,
      from_cm: false
    },
    isReady: false,
    dispatchConsent: false,
    currentUser: undefined,
    currentDevice: undefined,
    currentSession: undefined,
    currentPageview: undefined,
    currentPageviewVisibleSince: undefined,
    currentPageviewDuration: 0,
    currentCart: undefined,
    itemsQueue: {
      items: [],
      add: function add(kind, data) {
        // if current session expired, and received an interaction event, create a new session
        var sessionCookie = Rimdian.getCookie(Rimdian.config.namespace + 'session');

        if (!sessionCookie && (['pageview', 'cart', 'order'].includes(kind) || kind === 'custom_event' && data.non_interactive === true)) {
          Rimdian._startNewSession({});
        }

        var item = {
          kind: kind
        };
        item[kind] = data; // interations have a session + user + device context

        if (['pageview', 'custom_event', 'cart', 'order'].includes(kind)) {
          item.user = __assign({}, Rimdian.currentUser);
          item.session = __assign({}, Rimdian.currentSession);
        }

        Rimdian.itemsQueue.items.push(item); // persist items in local storage

        Rimdian._localStorage.set('items', JSON.stringify(Rimdian.itemsQueue.items));
      },
      // addPageviewDuration() is called when page becomes passive/not focused
      // pageview duration should not trigger a new session if the session expired
      // that's why we handle it separately
      addPageviewDuration: function addPageviewDuration() {
        // abort if we are not tracking the current pageview
        if (!Rimdian.currentPageview || !Rimdian.currentPageviewVisibleSince) {
          return;
        } // increment the time spent


        var increment = Math.round((new Date().getTime() - Rimdian.currentPageviewVisibleSince.getTime()) / 1000);
        Rimdian.currentPageviewDuration += increment;
        Rimdian.log('info', 'time spent on page is now', Rimdian.currentPageviewDuration); // add pageview to items queue

        var pageview = __assign({}, Rimdian.currentPageview);

        var now = new Date().toISOString();
        pageview.duration = Rimdian.currentPageviewDuration;
        pageview.updated_at = now; // increment session duration too

        if (!Rimdian.currentSession.duration) Rimdian.currentSession.duration = 0;
        Rimdian.currentSession.duration += increment;
        Rimdian.currentSession.updated_at = now;
        Rimdian.setSessionContext(Rimdian.currentSession); // cookie update
        // error on invalid pageview
        // if (!ValidatePageview(pageview)) {
        //   Rimdian.log('error', 'invalid pageview', ValidatePageview.errors)
        // }

        var item = {
          kind: 'pageview',
          pageview: pageview,
          user: __assign({}, Rimdian.currentUser),
          session: __assign({}, Rimdian.currentSession),
          device: __assign({}, Rimdian.currentDevice)
        };
        Rimdian.itemsQueue.items.push(item); // persist items in local storage

        Rimdian._localStorage.set('items', JSON.stringify(Rimdian.itemsQueue.items));
      }
    },
    dispatchQueue: [],
    isDispatching: false,
    onReadyQueue: [],
    log: function log(level) {
      var args = [];

      for (var _i = 1; _i < arguments.length; _i++) {
        args[_i - 1] = arguments[_i];
      }

      switch (level) {
        case 'warn':
          if (['warn', 'info', 'debug', 'trace'].includes(Rimdian.config.log_level)) {
            console.warn.apply(console, args);
          }

          break;

        case 'info':
          if (['info', 'debug', 'trace'].includes(Rimdian.config.log_level)) {
            console.info.apply(console, args);
          }

          break;

        case 'debug':
          if (['debug', 'trace'].includes(Rimdian.config.log_level)) {
            console.debug.apply(console, args);
          }

          break;

        case 'trace':
          if (Rimdian.config.log_level === 'trace') {
            console.trace.apply(console, args);
          }

          break;
        // print errors by default

        default:
          console.error.apply(console, args);
      }
    },
    // watch for DOM to be ready and call onReady()
    init: function init(cfg) {
      // continue if DOM is ready
      if (document.readyState === 'complete' || document.readyState === 'interactive') {
        Rimdian._onReady(cfg);
      } // watch for DOM readiness


      document.onreadystatechange = function () {
        if (document.readyState === 'interactive' || document.readyState === 'complete') {
          Rimdian.log('info', 'document is now', document.readyState);

          Rimdian._onReady(cfg);
        }
      };

      var logLevel = Rimdian.getCookie(Rimdian.config.namespace + 'debug');

      if (logLevel) {
        Rimdian.config.log_level = 'info';
      }
    },
    setDispatchConsent: function setDispatchConsent(consent) {
      Rimdian.log('info', 'RMD dispatch consent is now', consent);
      Rimdian.dispatchConsent = consent;
    },
    // return callback when the user is ready
    getCurrentUser: function getCurrentUser(callback) {
      // return callback when the user is ready
      if (Rimdian.currentUser) {
        callback(Rimdian.currentUser);
        return;
      }

      Rimdian._execWhenReady(function () {
        callback(Rimdian.currentUser);
      });
    },
    onReady: function onReady(fn) {
      if (Rimdian.isReady) {
        fn();
      } else {
        Rimdian.onReadyQueue.push(fn);
      }
    },
    // tracks the current pageview
    trackPageview: function trackPageview(data) {
      if (!Rimdian.isReady) {
        Rimdian.log('debug', 'RMD is not yet ready for trackPageview, queuing function...');

        Rimdian._execWhenReady(function () {
          return Rimdian.trackPageview(data);
        });

        return;
      } // persist existing pageview duration if any, or discard it if not active
      // because the pageview duration is already persisted when pages become passive


      if (Rimdian.currentPageview && !Rimdian.currentPageviewVisibleSince) {
        Rimdian.log('info', 'previous pageview has been discarded, was not active');
        Rimdian.currentPageview = undefined;
      }

      var now = new Date().toISOString(); // check if a pageview is already tracked (Single Page App)

      if (Rimdian.currentPageview) {
        Rimdian.log('info', 'pageview already tracked'); // compute pageview duration

        var increment = Math.round((new Date().getTime() - Rimdian.currentPageviewVisibleSince.getTime()) / 1000);
        Rimdian.currentPageviewDuration += increment;
        Rimdian.currentPageview.duration = Rimdian.currentPageviewDuration;
        Rimdian.currentPageview.updated_at = now; // increment session duration too

        if (!Rimdian.currentSession.duration) Rimdian.currentSession.duration = 0;
        Rimdian.currentSession.duration += increment;
        Rimdian.currentSession.updated_at = now; // session is persisted in cookie below
        // add current pageview to items queue

        Rimdian.itemsQueue.add('pageview', __assign({}, Rimdian.currentPageview)); // reset timer

        Rimdian.currentPageviewVisibleSince = undefined;
        Rimdian.currentPageviewDuration = 0;
      } // deep clone pageview


      var pageview = JSON.parse(JSON.stringify(data || {}));
      pageview.external_id = Rimdian.uuidv4();
      pageview.created_at = now;
      var referrer = Rimdian.getReferrer();

      if (referrer) {
        pageview.referrer = referrer;
      } // set defaults


      if (!pageview.title) {
        pageview.title = document.title;
      }

      if (!pageview.page_id) {
        pageview.page_id = window.location.href;
      } // escape title for JSON


      pageview.title = pageview.title.replace(/\\"/, '"'); // amount in cents

      if (pageview.product_price && pageview.product_price > 0) {
        pageview.product_price = Math.round(pageview.product_price * 100);
      } // init visibility tracking


      if (Rimdian.isPageVisible()) {
        Rimdian.currentPageviewVisibleSince = new Date();
      } // // error on invalid pageview
      // if (!ValidatePageview(pageview)) {
      //   Rimdian.log('error', 'invalid pageview', ValidatePageview.errors)
      // }
      // set current pageview


      Rimdian.currentPageview = pageview; // update user last interaction

      Rimdian.currentUser.last_interaction_at = now;
      Rimdian.setUserContext(Rimdian.currentUser); // cookie update
      // increment pageview count + interaction count of session

      if (!Rimdian.currentSession.pageviews_count) Rimdian.currentSession.pageviews_count = 0;
      if (!Rimdian.currentSession.interactions_count) Rimdian.currentSession.interactions_count = 0;
      Rimdian.currentSession.pageviews_count++;
      Rimdian.currentSession.interactions_count++;
      Rimdian.currentSession.updated_at = now;
      Rimdian.setSessionContext(Rimdian.currentSession); // cookie update
      // enqueue pageview

      Rimdian.itemsQueue.add('pageview', __assign({}, pageview));
    },
    // tracks the current customEvent
    trackCustomEvent: function trackCustomEvent(data) {
      if (!Rimdian.isReady) {
        Rimdian.log('debug', 'RMD is not yet ready for trackCustomEvent, queuing function...');

        Rimdian._execWhenReady(function () {
          return Rimdian.trackCustomEvent(data);
        });

        return;
      }

      if (data && !data.label) {
        Rimdian.log('error', 'customEvent label is required');
        return;
      } // deep clone customEvent


      var customEvent = JSON.parse(JSON.stringify(data || {}));
      var now = new Date().toISOString();
      customEvent.external_id = Rimdian.uuidv4();
      customEvent.created_at = now; // escape label for JSON

      customEvent.label = customEvent.label.replace(/\\"/, '"');

      if (customEvent.string_value) {
        // escape string_value for JSON
        customEvent.string_value = customEvent.string_value.replace(/\\"/, '"');
      } // // error on invalid customEvent
      // if (!ValidatePageview(customEvent)) {
      //   Rimdian.log('error', 'invalid customEvent', ValidatePageview.errors)
      // }
      // update user last interaction


      if (!customEvent.non_interactive) {
        Rimdian.currentUser.last_interaction_at = now;
        Rimdian.setUserContext(Rimdian.currentUser); // cookie update
        // increment customEvent count + interaction count of session

        if (!Rimdian.currentSession.interactions_count) Rimdian.currentSession.interactions_count = 0;
        Rimdian.currentSession.interactions_count++;
        Rimdian.currentSession.updated_at = now;
        Rimdian.setSessionContext(Rimdian.currentSession); // cookie update
      } // enqueue customEvent


      Rimdian.itemsQueue.add('custom_event', __assign({}, customEvent));
    },
    trackCart: function trackCart(data) {
      if (!Rimdian.isReady) {
        Rimdian.log('debug', 'RMD is not yet ready for trackCart, queuing function...');

        Rimdian._execWhenReady(function () {
          return Rimdian.trackCart(data);
        });

        return;
      } // check if data exists and is an object


      if (!data || _typeof(data) !== 'object') {
        Rimdian.log('error', 'invalid cart data');
        return;
      } // deep clone cart


      var cart = JSON.parse(JSON.stringify(data));
      Rimdian.log('info', 'cart is', cart);

      if (!cart.external_id) {
        cart.external_id = Rimdian.uuidv4();
      }

      cart.session_external_id = Rimdian.currentSession.external_id;

      if (cart.items) {
        // convert item prices to cents
        cart.items.forEach(function (item) {
          item.cart_external_id = cart.external_id;

          if (item.price && item.price > 0) {
            item.price = Math.round(item.price * 100);
          }
        });
      } // // error on invalid cart
      // if (!ValidateCart(cart)) {
      //   Rimdian.log('error', 'invalid cart', ValidateCart.errors)
      // }
      // compute a cart hash if not provided


      if (!cart.hash) {
        cart.hash = Rimdian._cartHash(cart);
      } // check if a cart is already tracked (Single Page App)


      if (Rimdian.currentCart) {
        Rimdian.log('info', 'cart already tracked', __assign({}, Rimdian.currentCart)); // skip if cart hash is the same

        if (Rimdian.currentCart.hash === cart.hash) {
          return;
        }
      }

      var now = new Date().toISOString();

      if (!cart.updated_at) {
        cart.updated_at = new Date().toISOString();
      }

      Rimdian.currentCart = cart; // increment interaction count of session
      // update user last interaction

      Rimdian.currentUser.last_interaction_at = now;

      if (!Rimdian.currentSession.interactions_count) {
        Rimdian.currentSession.interactions_count = 0;
      }

      Rimdian.currentSession.interactions_count++;
      Rimdian.currentSession.updated_at = now;
      Rimdian.setSessionContext(Rimdian.currentSession); // cookie update

      Rimdian.itemsQueue.add('cart', cart);
    },
    trackOrder: function trackOrder(data) {
      if (!Rimdian.isReady) {
        Rimdian.log('debug', 'RMD is not yet ready for trackOrder, queuing function...');

        Rimdian._execWhenReady(function () {
          return Rimdian.trackOrder(data);
        });

        return;
      } // check if data exists and is an object


      if (!data || _typeof(data) !== 'object') {
        Rimdian.log('error', 'invalid order data');
        return;
      } // deep clone order


      var order = JSON.parse(JSON.stringify(data)); // convert item prices to cents

      if (order.items) {
        order.items.forEach(function (item) {
          if (item.price && item.price > 0) {
            item.price = Math.round(item.price * 100);
          }

          item.order_external_id = order.external_id;
        });
      } // // error on invalid order, but don't abort
      // if (!ValidateOrder(order)) {
      //   Rimdian.log('error', 'invalid order', ValidateOrder.errors)
      // }


      var now = new Date().toISOString(); // update user last interaction

      Rimdian.currentUser.last_interaction_at = now;
      Rimdian.setUserContext(Rimdian.currentUser); // cookie update
      // increment interaction count of session

      if (!Rimdian.currentSession.interactions_count) Rimdian.currentSession.interactions_count = 0;
      Rimdian.currentSession.interactions_count++;
      Rimdian.currentSession.updated_at = now;
      Rimdian.setSessionContext(Rimdian.currentSession); // cookie update
      // enqueue order after user+session update

      Rimdian.itemsQueue.add('order', order);
    },
    setDeviceContext: function setDeviceContext(data) {
      if (!Rimdian.isReady) {
        Rimdian.log('debug', 'RMD is not yet ready for setDeviceContext, queuing function...');

        Rimdian._execWhenReady(function () {
          return Rimdian.setDeviceContext(data);
        });

        return;
      } // merge currentDevice with data


      var newDevice = __assign(__assign({}, Rimdian.currentDevice), data); // enriching the device data should trigger a DB update with updated_at


      newDevice.updated_at = new Date().toISOString(); // // validate current device
      // if (!ValidateDevice(Rimdian.currentDevice)) {
      //   Rimdian.log('error', 'invalid device', ValidateDevice.errors)
      //   return
      // }

      Rimdian.currentDevice = newDevice; // persist updated device

      Rimdian.setCookie(Rimdian.config.namespace + 'device', JSON.stringify(Rimdian.currentDevice), OneYearInSeconds);
    },
    setSessionContext: function setSessionContext(data) {
      if (!Rimdian.isReady) {
        Rimdian.log('debug', 'RMD is not yet ready for setSessionContext, queuing function...');

        Rimdian._execWhenReady(function () {
          return Rimdian.setSessionContext(data);
        });

        return;
      } // merge currentSession with data


      var newSession = __assign(__assign({}, Rimdian.currentSession), data); // enriching the device data should trigger a DB update with updated_at


      newSession.updated_at = new Date().toISOString();
      Rimdian.currentSession = newSession;
      Rimdian.log('info', 'RMD updated session is', newSession); // persist updated session

      Rimdian.setCookie(Rimdian.config.namespace + 'session', JSON.stringify(Rimdian.currentSession), Rimdian.config.session_timeout);
    },
    // check if user ID has changed
    // reset context if new user is also authenticated
    // or user_alias if previous was anonymous
    // set new user fields and eventually enqueue user
    setUserContext: function setUserContext(data) {
      if (!Rimdian.isReady) {
        Rimdian.log('debug', 'RMD is not yet ready for setUserContext, queuing function...');

        Rimdian._execWhenReady(function () {
          return Rimdian.setUserContext(data);
        });

        return;
      } // replace user_centric_consent by consent_all


      if (data && data.user_centric_consent !== undefined) {
        data.consent_all = data.user_centric_consent;
        delete data.user_centric_consent;
      }

      if (data && data.external_id && data.external_id !== '' && data.external_id !== Rimdian.currentUser.external_id) {
        // if previous user was authenticated, reset new session / device, as we can't alias 2 authenticated users
        if (Rimdian.currentUser.is_authenticated && data.is_authenticated === true) {
          Rimdian.log('info', 'new authenticated user detected, reset context', data); // reset device

          Rimdian.currentDevice = undefined;

          Rimdian._localStorage.remove('device');

          Rimdian._createDevice(); // reset current session after device


          Rimdian.currentSession = undefined;
          Rimdian.deleteCookie(Rimdian.config.namespace + 'session');

          Rimdian._handleSession(); // loop over keys and ignore empty strings


          Object.keys(data).forEach(function (key) {
            if (typeof data[key] === 'string' && data[key] === '') {
              delete data[key];
            }
          });
          Rimdian.currentUser = __assign({}, data); // set defaults

          if (Rimdian.currentUser.created_at === undefined) {
            Rimdian.currentUser.created_at = new Date().toISOString();
          } // // validate current user
          // if (!ValidateUser(Rimdian.currentUser)) {
          //   Rimdian.log('error', 'invalid user', ValidateUser.errors)
          // }


          return;
        } // create a user_alias if previous user was anonymous
        // and merge new user data & enqueue user


        if (!Rimdian.currentUser.is_authenticated) {
          Rimdian.log('info', 'alias previous user', Rimdian.currentUser.external_id, 'with user', data);
          Rimdian.itemsQueue.add('user_alias', {
            from_user_external_id: Rimdian.currentUser.external_id,
            to_user_external_id: data.external_id,
            to_user_is_authenticated: data.is_authenticated === true,
            to_user_created_at: data.created_at ? data.created_at : new Date().toISOString()
          });
        }
      } // update user context with new data


      var newUser = __assign(__assign({}, Rimdian.currentUser), data); // // validate current user
      // if (!ValidateUser(Rimdian.currentUser)) {
      //   // abort on error
      //   Rimdian.log('error', 'invalid user', ValidateUser.errors)
      //   return
      // }


      Rimdian.currentUser = newUser; // persist updated user

      Rimdian.setCookie(Rimdian.config.namespace + 'user', JSON.stringify(Rimdian.currentUser), OneYearInSeconds);
    },
    // enqueue the current user
    saveUserProfile: function saveUserProfile() {
      if (!Rimdian.isReady) {
        Rimdian.log('debug', 'RMD is not yet ready for saveUserProfile, queuing function...');

        Rimdian._execWhenReady(function () {
          return Rimdian.saveUserProfile();
        });

        return;
      }

      Rimdian.itemsQueue.add('user', __assign({}, Rimdian.currentUser));
    },
    isPageVisible: function isPageVisible() {
      return w.state === PageStates.active;
    },
    // dispatch items to the collector
    dispatch: function dispatch(useBeacon) {
      // wait for tracker to be ready
      if (Rimdian.isReady === false) {
        Rimdian.log('info', 'RMD is not ready, retrying dispatch soon...');
        window.setTimeout(function () {
          Rimdian.dispatch(useBeacon);
        }, 50);
        return;
      } // abort if we don't have user consent


      if (!Rimdian.dispatchConsent) {
        Rimdian.log('info', 'RMD abort dispatch, no dispatch consent');
        return;
      } // abort if we don't have items to dispatch


      if (Rimdian.itemsQueue.items.length === 0 && Rimdian.dispatchQueue.length === 0) {
        Rimdian.log('info', 'RMD abort dispatch, no items to dispatch');
        return;
      }

      var deviceCtx = __assign({}, Rimdian.currentDevice); // set updated_at only if user_agent changed, it will trigger a DB update


      if (deviceCtx.user_agent !== navigator.userAgent) {
        deviceCtx.updated_at = new Date().toISOString();
      }

      deviceCtx.user_agent = navigator.userAgent;
      deviceCtx.language = navigator.language;
      deviceCtx.ad_blocker = Rimdian.hasAdBlocker();
      deviceCtx.resolution = window.screen && window.screen.width && window.screen.height ? window.screen.height > window.screen.width ? window.screen.height + 'x' + window.screen.width : window.screen.width + 'x' + window.screen.height : undefined; // create dataImport batches of 10 items maximum

      var batches = [];
      var itemsBatch = [];

      while (Rimdian.itemsQueue.items.length > 0) {
        itemsBatch.push(Rimdian.itemsQueue.items.shift());

        if (itemsBatch.length >= 10) {
          batches.push(itemsBatch);
          itemsBatch = [];
        }
      } // add remaining items to last batch


      if (itemsBatch.length > 0) {
        batches.push(itemsBatch);
      } // convert batches into IDataImport objects


      batches.forEach(function (batch) {
        // add device to items
        batch.forEach(function (item) {
          item.device = deviceCtx;
        });
        Rimdian.dispatchQueue.push({
          id: Rimdian.uuidv4(),
          workspace_id: Rimdian.config.workspace_id,
          items: batch,
          context: {// set this field right before sending data
            // data_sent_at: new Date().toISOString()
          },
          created_at: new Date().toISOString()
        });
      }); // abort if we don't have data imports to dispatch

      if (Rimdian.dispatchQueue.length === 0) {
        Rimdian.log('info', 'RMD abort dispatch, no data imports to dispatch');
        return;
      } // persist dispatch queue in local storage


      Rimdian._localStorage.set('dispatchQueue', JSON.stringify(Rimdian.dispatchQueue)); // send data to collector


      Rimdian.log('info', 'RMD sending data to collector');

      Rimdian._initDispatchLoop(useBeacon);
    },
    _initDispatchLoop: function _initDispatchLoop(useBeacon) {
      // abort if we are already dispatching
      if (Rimdian.isDispatching) {
        Rimdian.log('info', 'RMD abort dispatch, already dispatching');
        return;
      } // dispatch items


      Rimdian.isDispatching = true; // post the hits, starting with the oldest batch first if it exists

      Rimdian.dispatchQueue.sort(function (a, b) {
        if (a.created_at < b.created_at) {
          return -1; // a listed before b
        }

        if (a.created_at > b.created_at) {
          return 1;
        }

        return 0;
      });
      var currentBatch = Rimdian.dispatchQueue[0]; // post payload

      Rimdian._postPayload(currentBatch, 0, useBeacon);
    },
    _postPayload: function _postPayload(dataImport, retryCount, useBeacon) {
      Rimdian._post(dataImport, useBeacon, function (error) {
        var success = true;

        if (error) {
          // retry or requeue
          Rimdian.log('error', 'RMD post payload error', error); // parse error, ignore error 400

          try {
            var jsonError = JSON.parse(error) || {}; // error is not a bad hit

            if (jsonError.code > 400) {
              success = false;
            }
          } catch (err) {
            success = false;
          }

          Rimdian.log('info', 'sucess', success, 'retry count', retryCount);

          if (success === false) {
            // stop after 10 retry
            if (retryCount >= Rimdian.config.max_retry) {
              Rimdian.isDispatching = false;
              Rimdian.log('info', 'max retry reached, aborting');
              return;
            } // exponential backoff


            var delay = 100; // ms

            var toWait = delay;

            if (retryCount > 1) {
              for (var i = 0; i < retryCount; i++) {
                toWait = toWait * 2;
              }
            }

            Rimdian.log('debug', 'retry in ' + toWait + 'ms'); // wait 500 ms

            window.setTimeout(function () {
              retryCount++;

              Rimdian._postPayload(dataImport, retryCount, useBeacon);
            }, toWait);
            return;
          }
        }

        if (success === true) {
          // remove from the dispatch queue
          var remainingBatches_1 = []; // keep other batches

          Rimdian.dispatchQueue.forEach(function (di) {
            if (di.id !== dataImport.id) {
              remainingBatches_1.push(di);
            }
          });
          Rimdian.dispatchQueue = remainingBatches_1; // persist dispatch queue in local storage

          Rimdian._localStorage.set('dispatchQueue', JSON.stringify(Rimdian.dispatchQueue)); // continue the loop


          Rimdian.isDispatching = false;

          if (Rimdian.dispatchQueue.length > 0) {
            Rimdian._initDispatchLoop(useBeacon);
          }
        }
      });
    },
    _post: function _post(dataImport, useBeacon, callback) {
      // set client clock right before sending data
      dataImport.context.data_sent_at = new Date().toISOString();
      var data = JSON.stringify(dataImport); // log info data

      Rimdian.log('info', 'RMD sending data', data); // send dataImport to collector using beacon

      if (useBeacon && navigator.sendBeacon) {
        var queued = navigator.sendBeacon(Rimdian.config.host + '/live', new Blob([data], {
          type: 'application/json'
        }));
        return callback(queued ? null : 'sendBeacon failed');
      }

      var xhr = new XMLHttpRequest();

      xhr.onload = function () {
        var body = 'response' in xhr ? xhr.response : xhr['responseText'];

        if (xhr.status >= 300) {
          return callback(body);
        }

        return callback(null);
      };

      xhr.onerror = function () {
        return callback('network request failed');
      };

      xhr.ontimeout = function () {
        return callback('network request timeout');
      };

      xhr.open('POST', Rimdian.config.host + '/live', true);
      xhr.setRequestHeader('Content-Type', 'application/json');
      xhr.withCredentials = true;
      xhr.send(data);
    },
    // retrieve existing user ID or create a new one
    // 1. id found in URL
    // 2. OR id exists in a cookie
    // 3. OR id is created
    _handleUser: function _handleUser() {
      // sometimes user IDs are injected by template engines and are not parsed - ie: {{ user_id }}
      // to avoid using/merging invalid user IDs, we forbid some patterns
      var forbiddenPatterns = ['{{', '}}', '{%', '%}', '{#', '#}', '*|', '|*'];
      var userCookie = Rimdian.getCookie(Rimdian.config.namespace + 'user');
      var previousUser = {};

      if (userCookie && userCookie !== '') {
        previousUser = JSON.parse(userCookie); // ignore cookies where the user ID was a forbidden pattern
        // this is used to erase legacy cookies that allowed fordidden patterns in the past

        if (forbiddenPatterns.some(function (pattern) {
          return previousUser.external_id && previousUser.external_id.indexOf(pattern) !== -1;
        })) {
          previousUser = {};
          userCookie = '';
        } // we have a legacy cookie, migrate it


        if (Rimdian.config.from_cm === true && previousUser.id !== undefined && previousUser.external_id === undefined) {
          previousUser = {
            external_id: previousUser.id,
            is_authenticated: previousUser.is_authenticated || false,
            created_at: new Date().toISOString(),
            hmac: previousUser.hmac || undefined
          };
        }
      }

      var userId = Rimdian.getQueryParam(document.URL, URLParams.user_external_id); // 1. ID found in URL
      // check if the user ID does not contain forbidden patterns

      if (userId && userId !== '' && forbiddenPatterns.every(function (pattern) {
        return userId.indexOf(pattern) === -1;
      })) {
        Rimdian.log('info', 'found user ID in URL', userId);
        var isAuthenticated = Rimdian.getQueryParam(document.URL, URLParams.user_is_authenticated);
        Rimdian.log('info', 'user authenticated URL value', isAuthenticated);
        Rimdian.currentUser = {
          external_id: userId,
          is_authenticated: isAuthenticated === 'true' || isAuthenticated === '1',
          created_at: new Date().toISOString(),
          hmac: Rimdian.getQueryParam(document.URL, URLParams.user_external_id_hmac)
        }; // alias previous user if was unknown

        if (userCookie && userCookie !== '') {
          Rimdian.log('info', 'found another user ID in cookie', userCookie);

          if (previousUser.is_authenticated === false && previousUser.external_id !== Rimdian.currentUser.external_id) {
            Rimdian.log('info', 'alias previous user', previousUser.external_id, 'with user', Rimdian.currentUser.external_id);
            Rimdian.itemsQueue.add('user_alias', {
              from_user_external_id: previousUser.external_id,
              to_user_external_id: Rimdian.currentUser.external_id,
              to_user_is_authenticated: Rimdian.currentUser.is_authenticated,
              to_user_created_at: Rimdian.currentUser.created_at
            });
          }
        }

        Rimdian.setCookie(Rimdian.config.namespace + 'user', JSON.stringify(Rimdian.currentUser), OneYearInSeconds);
        return;
      } // 2. ID found in cookie


      if (userCookie && userCookie !== '') {
        Rimdian.log('info', 'found user ID in cookie', userCookie);
        Rimdian.currentUser = previousUser;
        Rimdian.setCookie(Rimdian.config.namespace + 'user', JSON.stringify(Rimdian.currentUser), OneYearInSeconds);
        return;
      } // 3. ID is created, with anonymous user


      Rimdian._createUser(Rimdian.uuidv4(), false, new Date().toISOString());
    },
    _createUser: function _createUser(userExternalId, isAuthenticated, createdAt) {
      Rimdian.currentUser = {
        external_id: userExternalId,
        is_authenticated: isAuthenticated,
        created_at: createdAt
      };
      Rimdian.log('info', 'creating new user', __assign({}, Rimdian.currentUser));
      Rimdian.setCookie(Rimdian.config.namespace + 'user', JSON.stringify(Rimdian.currentUser), OneYearInSeconds);
    },
    // retrieve eventual user details provided in URL
    _enrichUserContext: function _enrichUserContext() {
      // update current user profile with other parameters
      var otherUserIds = [{
        key: 'email',
        value: Rimdian.getQueryParam(document.URL, '_email')
      }, {
        key: 'email_md5',
        value: Rimdian.getQueryParam(document.URL, '_emailmd5')
      }, {
        key: 'email_sha1',
        value: Rimdian.getQueryParam(document.URL, '_emailsha1')
      }, {
        key: 'email_sha256',
        value: Rimdian.getQueryParam(document.URL, '_emailsha256')
      }, {
        key: 'telephone',
        value: Rimdian.getQueryParam(document.URL, '_telephone')
      }]; // enrich user context with other parameters

      otherUserIds.forEach(function (x, i) {
        if (x.value && x.value !== '') {
          Rimdian.currentUser[x.key] = x.value;
        }
      }); // update user cookie

      Rimdian.setCookie(Rimdian.config.namespace + 'user', JSON.stringify(Rimdian.currentUser), OneYearInSeconds);
    },
    // retrieve existing device ID or create a new one
    // 1. id found in URL
    // 2. id exists in a cookie
    // 3. id is created
    // to save cookie space, device context is enriched while dispatching data
    _handleDevice: function _handleDevice() {
      var deviceId = Rimdian.getQueryParam(document.URL, URLParams.device_external_id); // 1. ID found in URL

      if (deviceId && deviceId !== '') {
        Rimdian.log('info', 'found device ID in URL', deviceId);
        Rimdian.currentDevice = {
          external_id: deviceId,
          created_at: new Date().toISOString(),
          user_agent: navigator.userAgent
        };
        Rimdian.setCookie(Rimdian.config.namespace + 'device', JSON.stringify(Rimdian.currentDevice), OneYearInSeconds);
        return;
      } // 2. ID found in cookie


      var deviceCookie = Rimdian.getCookie(Rimdian.config.namespace + 'device');

      if (deviceCookie && deviceCookie !== '') {
        Rimdian.log('info', 'found device ID in cookie', deviceCookie);
        Rimdian.currentDevice = JSON.parse(deviceCookie); // check if we already had a legacy client ID in a cookie

        if (Rimdian.config.from_cm === true) {
          var legacyClientID = Rimdian.getCookie('_cm_cid');

          if (legacyClientID && legacyClientID !== '') {
            Rimdian.currentDevice.external_id = legacyClientID; // extract its creation timestamp

            var legacyClientTimestamp = Rimdian.getCookie('_cm_cidat');

            if (legacyClientTimestamp && legacyClientTimestamp !== '') {
              Rimdian.currentDevice.created_at = new Date(parseInt(legacyClientTimestamp, 10)).toISOString();
            }
          }
        }

        Rimdian.setCookie(Rimdian.config.namespace + 'device', JSON.stringify(Rimdian.currentDevice), OneYearInSeconds);
        return;
      } // 3. ID is created


      Rimdian._createDevice();
    },
    _createDevice: function _createDevice() {
      Rimdian.log('info', 'creating new device ID');
      Rimdian.currentDevice = {
        external_id: Rimdian.uuidv4(),
        created_at: new Date().toISOString(),
        user_agent: navigator.userAgent
      };
      Rimdian.setCookie(Rimdian.config.namespace + 'device', JSON.stringify(Rimdian.currentDevice), OneYearInSeconds);
    },
    // polyfill
    _addEventListener: function _addEventListener(element, eventType, eventHandler, useCapture) {
      if (element.addEventListener) {
        element.addEventListener(eventType, eventHandler, useCapture);
        return true;
      }

      if (element.attachEvent) {
        return element.attachEvent('on' + eventType, eventHandler);
      }

      element['on' + eventType] = eventHandler;
    },
    // load config, existing data and start tracking
    _onReady: function _onReady(cfg) {
      // avoid calling onReady twice
      if (Rimdian.isReady) return;
      Rimdian.isReady = true;
      Rimdian.log('info', 'onReady() called'); // check if browser is legit

      if (!Rimdian.isBrowserLegit()) {
        Rimdian.log('warn', 'Browser is not legit');
        return;
      } // merge cfg with default config & validate


      var config = __assign(__assign({}, Rimdian.config), cfg); // if (!ValidateConfig(config)) {
      //   Rimdian.log('error', 'RMD Config error:', ValidateConfig.errors)
      //   return
      // }
      // save config


      Rimdian.config = config;
      Rimdian.log('info', 'RMD Config is:', Rimdian.config); // load items and batches from localStorage

      if (window.localStorage) {
        Rimdian.itemsQueue.items = JSON.parse(localStorage.getItem(Rimdian.config.namespace + 'items') || '[]');
        Rimdian.dispatchQueue = JSON.parse(localStorage.getItem(Rimdian.config.namespace + 'dispatchQueue') || '[]');
      } // init user context


      Rimdian._handleUser();

      Rimdian._enrichUserContext(); // init device context


      Rimdian._handleDevice(); // init session context after device


      Rimdian._handleSession(); // execute queued functions


      if (Rimdian.onReadyQueue.length > 0) {
        Rimdian.onReadyQueue.forEach(function (x, i) {
          Rimdian.log('debug', 'executing queued function', x);
          x();
        });
        Rimdian.onReadyQueue = [];
      } // validate user / device / session
      // if (!ValidateUser(Rimdian.currentUser)) {
      //   Rimdian.log('error', 'invalid user', ValidateUser.errors)
      // }
      // if (!ValidateDevice(Rimdian.currentDevice)) {
      //   Rimdian.log('error', 'invalid device', ValidateDevice.errors)
      // }
      // if (!ValidateSession(Rimdian.currentSession)) {
      //   Rimdian.log('error', 'invalid session', ValidateSession.errors)
      // }
      // extend session expiration time every minute while pageview is visible


      window.setInterval(function () {
        if (Rimdian.currentPageview && Rimdian.isPageVisible()) {
          // get current session from cookie to see if it expired
          var cookieSession = Rimdian.getCookie(Rimdian.config.namespace + 'session');

          if (cookieSession) {
            // extend lifetime of session
            Rimdian.setCookie(Rimdian.config.namespace + 'session', cookieSession, Rimdian.config.session_timeout);
          }
        }
      }, 60000); // every minute
      // decorate cross domains links

      if (Rimdian.config.cross_domains.length > 0) {
        // loop over every <a> tag and add the decorateURL function
        for (var i = 0; i < document.links.length; i++) {
          var elt = document.links[i];
          Rimdian.config.cross_domains.forEach(function (d) {
            // only decorate links to the matching domain
            if (elt.href.indexOf(d) !== -1) {
              Rimdian._addEventListener(elt, 'click', Rimdian._decorateURL, true);

              Rimdian._addEventListener(elt, 'mousedown', Rimdian._decorateURL, true);
            }
          });
        }
      } // use visibilitychange event to send pageview secs in beacon
      // use https://github.com/GoogleChromeLabs/page-lifecycle
      // https://developer.chrome.com/blog/page-lifecycle-api/#advice-hidden
      // watch page state changes


      w.addEventListener('statechange', function (event) {
        Rimdian.log('info', 'page state changed from', event.oldState, 'to', event.newState);

        if (event.oldState === PageStates.active && event.newState === PageStates.passive) {
          Rimdian._onPagePassive();
        } else if (event.oldState === PageStates.passive && event.newState === PageStates.active) {
          Rimdian._onPageActive();
        }
      }); // dispatch items  automatically every 5 secs
      // window.setInterval(function () {
      //   Rimdian.dispatch(false)
      // }, 5000)
    },
    _execWhenReady: function _execWhenReady(fn) {
      if (Rimdian.isReady) {
        fn();
      } else {
        Rimdian.onReadyQueue.push(fn);
      }
    },
    // session is stored in a cookie and will expire with its cookie
    // scenarii:
    // 1. no existing session -> create new session
    // 2. existing session with same referrer origin -> continue session
    // 3. existing session with different referrer origin -> create new session
    _handleSession: function _handleSession() {
      // extract current utm params from url
      var utm_source = Rimdian.getQueryParam(document.URL, 'utm_source') || Rimdian.getHashParam(window.location.hash, 'utm_source');
      var utm_medium = Rimdian.getQueryParam(document.URL, 'utm_medium') || Rimdian.getHashParam(window.location.hash, 'utm_medium');
      var utm_campaign = Rimdian.getQueryParam(document.URL, 'utm_campaign') || Rimdian.getHashParam(window.location.hash, 'utm_campaign');
      var utm_content = Rimdian.getQueryParam(document.URL, 'utm_content') || Rimdian.getHashParam(window.location.hash, 'utm_content');
      var utm_term = Rimdian.getQueryParam(document.URL, 'utm_term') || Rimdian.getHashParam(window.location.hash, 'utm_term');
      var utm_id = Rimdian.getQueryParam(document.URL, 'utm_id') || Rimdian.getHashParam(window.location.hash, 'utm_id');
      var utm_id_from = Rimdian.getQueryParam(document.URL, 'utm_id_from') || Rimdian.getHashParam(window.location.hash, 'utm_id_from'); // extract referrer from url

      var referrer = Rimdian.getReferrer();

      if (referrer) {
        // parse referrer URL with <a> tag
        var referrerURL = document.createElement('a');
        referrerURL.href = referrer; // check if referrer came from another domain

        var fromAnotherDomain = false; // check if different domain is not is the cross_domain config

        if (referrerURL.hostname && referrerURL.hostname !== window.location.hostname) {
          var isCrossDomain_1 = false;

          if (Rimdian.config.cross_domains && Rimdian.config.cross_domains.length) {
            Rimdian.config.cross_domains.forEach(function (dom) {
              if (referrerURL.href.indexOf(dom) !== -1) {
                isCrossDomain_1 = true;
              }
            });
          }

          if (isCrossDomain_1 === false) {
            fromAnotherDomain = true;
          }
        }

        if (fromAnotherDomain) {
          utm_source = referrerURL.hostname; // utm_medium is referral by default

          if (!utm_medium || utm_medium === '') {
            utm_medium = 'referral';
          } // check if it comes from known search engines and extract search query if possible


          if (referrer.search('https?://(.*)google.([^/?]*)') === 0) {
            utm_term = Rimdian.getQueryParam(referrer, 'q');
          } else if (referrer.search('https?://(.*)bing.com') === 0) {
            utm_term = Rimdian.getQueryParam(referrer, 'q');
          } else if (referrer.search('https?://(.*)search.yahoo.com') === 0) {
            utm_term = Rimdian.getQueryParam(referrer, 'p');
          } else if (referrer.search('https?://(.*)ask.com') === 0) {
            utm_term = Rimdian.getQueryParam(referrer, 'q');
          } else if (referrer.search('https?://(.*)search.aol.com') === 0) {
            utm_term = Rimdian.getQueryParam(referrer, 'q');
          } else if (referrer.search('https?://(.*)duckduckgo.com') === 0) {
            utm_term = Rimdian.getQueryParam(referrer, 'q');
          }
        }
      } // extract gclid+fbclid+MSCLKID from url into utm_id + utm_id_from


      var ids = ['gclid', 'fbclid', 'msclkid'];
      ids.forEach(function (param) {
        var value = Rimdian.getQueryParam(document.URL, param) || Rimdian.getHashParam(window.location.hash, param);

        if (value) {
          utm_id = value;
          utm_id_from = param;
        }
      }); // google SEO & google Ads might both use the same source/medium
      // we detect Ads with gclid, and change its medium to
      // eventually trigger a new session

      if (utm_id_from === 'gclid' && utm_medium === 'referral') {
        utm_medium = 'ads';
      }

      Rimdian.log('info', 'RMD utm_source is:', utm_source);
      Rimdian.log('info', 'RMD utm_medium is:', utm_medium);
      Rimdian.log('info', 'RMD utm_campaign is:', utm_campaign);
      Rimdian.log('info', 'RMD utm_content is:', utm_content);
      Rimdian.log('info', 'RMD utm_term is:', utm_term);
      Rimdian.log('info', 'RMD utm_id is:', utm_id);
      Rimdian.log('info', 'RMD utm_id_from is:', utm_id_from); // read session cookie

      var sessionCookie = Rimdian.getCookie(Rimdian.config.namespace + 'session'); // 1. no existing session -> create new session

      if (!sessionCookie || sessionCookie === '') {
        Rimdian.log('info', 'RMD session cookie not found');

        Rimdian._startNewSession({
          utm_source: utm_source,
          utm_medium: utm_medium,
          utm_campaign: utm_campaign,
          utm_content: utm_content,
          utm_term: utm_term,
          utm_id: utm_id,
          utm_id_from: utm_id_from
        });

        return;
      } // check if this origin should be ignored


      var ignoredOrigin;

      if (utm_source && utm_source !== '' && Rimdian.config.ignored_origins.length > 0) {
        // find a matching origin
        ignoredOrigin = Rimdian.config.ignored_origins.find(function (origin) {
          // source medium matches
          if (origin.utm_source === utm_source && origin.utm_medium === utm_medium) {
            // if origin requires a campaign, check if it matches
            if (origin.utm_campaign && origin.utm_campaign !== '') {
              if (utm_campaign && origin.utm_campaign === utm_campaign) {
                return true;
              } // origin is not matching, continue


              return false;
            } // if origin does not require a campaign, its a match


            return true;
          }

          return false;
        });
      } // process existing session


      var existingSession = JSON.parse(sessionCookie);
      Rimdian.log('info', 'RMD existing session is:', existingSession); // check if session origin has changed from previous page

      var isEqual = true;
      if (utm_source && utm_source !== '' && existingSession.utm_source !== utm_source) isEqual = false;
      if (utm_medium && utm_medium !== '' && existingSession.utm_medium !== utm_medium) isEqual = false;
      if (utm_campaign && utm_campaign !== '' && existingSession.utm_campaign !== utm_campaign) isEqual = false;
      if (utm_content && utm_content !== '' && existingSession.utm_content !== utm_content) isEqual = false;
      if (utm_term && utm_term !== '' && existingSession.utm_term !== utm_term) isEqual = false;
      if (utm_id && utm_id !== '' && existingSession.utm_id !== utm_id) isEqual = false; // 2. if this origin is ignored, or same origin, or empty origin, resume session

      if (ignoredOrigin || isEqual || !utm_source || utm_source === '') {
        Rimdian.log('info', 'RMD resume session (ignored:' + (ignoredOrigin ? 'yes' : 'no') + ', isEqual:' + isEqual + ', utm_source:' + utm_source + ')');
        Rimdian.currentSession = existingSession;
        Rimdian.setCookie(Rimdian.config.namespace + 'session', JSON.stringify(Rimdian.currentSession), Rimdian.config.session_timeout);
        return;
      } // 3. origin has changed, start new session


      Rimdian._startNewSession({
        utm_source: utm_source,
        utm_medium: utm_medium,
        utm_campaign: utm_campaign,
        utm_content: utm_content,
        utm_term: utm_term,
        utm_id: utm_id,
        utm_id_from: utm_id_from
      });
    },
    _onPagePassive: function _onPagePassive() {
      Rimdian.log('info', 'page is passive state');
      Rimdian.itemsQueue.addPageviewDuration();
      Rimdian.dispatch(true); // use beacon as the window might be closing
    },
    _onPageActive: function _onPageActive() {
      Rimdian.log('info', 'page is active state'); // abort if we are not tracking the current pageview

      if (!Rimdian.currentPageview) {
        return;
      }

      Rimdian.currentPageviewVisibleSince = new Date(); // reset the timer
    },
    getTimezone: function getTimezone() {
      var _a;

      var DateTimeFormat = (_a = window.Intl) === null || _a === void 0 ? void 0 : _a.DateTimeFormat;

      if (DateTimeFormat) {
        var timezone = new DateTimeFormat().resolvedOptions().timeZone;

        if (timezone) {
          return timezone;
        }
      }

      return undefined;
    },
    getQueryParam: function getQueryParam(url, name) {
      try {
        var urlObject = new URL(url);
        var params = new URLSearchParams(urlObject.search);
        return params.get(name) || undefined;
      } catch (e) {
        return undefined;
      }
    },
    getHashParam: function getHashParam(hash, name) {
      var matches = hash.match(new RegExp(name + '=([^&]*)'));
      return matches ? matches[1] : undefined;
    },
    updateURLParam: function updateURLParam(url, name, value) {
      var urlObject = new URL(url);
      var params = new URLSearchParams(urlObject.search);
      params.set(name, value);
      urlObject.search = params.toString();
      return urlObject.toString();
    },
    hasAdBlocker: function hasAdBlocker() {
      var ads = document.createElement('div');
      ads.innerHTML = '&nbsp;';
      ads.className = 'adsbox';
      var blocked = false;

      try {
        // body may not exist, that's why we need try/catch
        document.body.appendChild(ads);
        blocked = document.getElementsByClassName('adsbox')[0].offsetHeight === 0;
        document.body.removeChild(ads);
      } catch (_e) {
        blocked = false;
      }

      return blocked;
    },
    isBrowserLegit: function isBrowserLegit() {
      // detect IE 9
      var ua = navigator.userAgent.toLowerCase();

      if (ua.indexOf('msie') !== -1) {
        if (parseInt(ua.split('msie')[1], 10) <= 9) {
          return false;
        }
      } // detect known bot


      if (/(google web preview|baiduspider|yandexbot|bingbot|googlebot|yahoo! slurp|nuhk|yammybot|openbot|slurp|msnBot|ask jeeves\/teoma|ia_archiver)/i.test(navigator.userAgent)) {
        return false;
      } // detect headless chrome


      if (navigator.webdriver) {
        return false;
      }

      return true;
    },
    uuidv4: function uuidv4() {
      return 'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'.replace(/[xy]/g, function (c) {
        var r = Math.random() * 16 | 0,
            v = c == 'x' ? r : r & 0x3 | 0x8;
        return v.toString(16);
      });
    },
    md5: function md5(str) {
      return _md(str);
    },
    getReferrer: function getReferrer() {
      var referrer = undefined;

      try {
        referrer = window.top.document.referrer !== '' ? window.top.document.referrer : undefined;
      } catch (e) {
        if (window.parent) {
          try {
            referrer = window.parent.document.referrer !== '' ? window.parent.document.referrer : undefined;
          } catch (_e) {
            referrer = undefined;
          }
        }
      }

      if (!referrer) {
        referrer = document.referrer !== '' ? document.referrer : undefined;
      }

      return referrer;
    },
    _startNewSession: function _startNewSession(params) {
      // default to direct / none on empty origin
      if (!params.utm_source || params.utm_source === '') {
        params.utm_source = 'direct';
      }

      if (!params.utm_medium || params.utm_medium === '') {
        params.utm_medium = 'none';
      }

      Rimdian.currentSession = {
        external_id: Rimdian.uuidv4(),
        created_at: new Date().toISOString(),
        device_external_id: Rimdian.currentDevice.external_id,
        landing_page: window.location.href,
        referrer: Rimdian.getReferrer(),
        timezone: Rimdian.getTimezone(),
        utm_source: params.utm_source,
        utm_medium: params.utm_medium,
        utm_campaign: params.utm_campaign,
        utm_content: params.utm_content,
        utm_term: params.utm_term,
        utm_id: params.utm_id,
        utm_id_from: params.utm_id_from,
        duration: 0,
        pageviews_count: 0,
        interactions_count: 0
      };
      Rimdian.log('info', 'RMD new session is:', Rimdian.currentSession); // persist session to cookie

      Rimdian.setCookie(Rimdian.config.namespace + 'session', JSON.stringify(Rimdian.currentSession), Rimdian.config.session_timeout);
    },
    getCookie: function getCookie(name) {
      return decodeURIComponent(document.cookie.replace(new RegExp('(?:(?:^|.*;)\\s*' + encodeURIComponent(name).replace(/[-.+*]/g, '\\$&') + '\\s*\\=\\s*([^;]*).*$)|^.*$'), '$1')) || null;
    },
    // cookies are secured and cross-domain by default
    setCookie: function setCookie(name, value, seconds) {
      // cross_domain
      var matches = window.location.hostname.match(/[a-z0-9][a-z0-9\-]+\.[a-z\.]{2,6}$/i);
      var domain = matches ? matches[0] : '';
      var xdomain = domain ? '; domain=.' + domain : '';
      var now = new Date();
      now.setTime(now.getTime() + seconds * 1000);
      var expires = '; expires=' + now.toUTCString();
      var cookie_value = name + '=' + encodeURIComponent(value) + expires + '; path=/' + xdomain + '; secure';
      document.cookie = cookie_value;
      return;
    },
    deleteCookie: function deleteCookie(name) {
      Rimdian.setCookie(name, '', -1);
    },
    _localStorage: {
      get: function get(key) {
        return localStorage.getItem(Rimdian.config.namespace + key);
      },
      set: function set(key, value) {
        try {
          localStorage.setItem(Rimdian.config.namespace + key, value);
        } catch (e) {
          Rimdian.log('error', 'localStorage error:', e);
        }
      },
      remove: function remove(key) {
        localStorage.removeItem(Rimdian.config.namespace + key);
      }
    },
    // inject the device + user ids on the fly
    _decorateURL: function _decorateURL(e) {
      var target = e.target;
      target.href = Rimdian.updateURLParam(target.href, URLParams.device_external_id, Rimdian.currentDevice.external_id);
      target.href = Rimdian.updateURLParam(target.href, URLParams.user_external_id, Rimdian.currentUser.external_id);
      target.href = Rimdian.updateURLParam(target.href, URLParams.user_is_authenticated, Rimdian.currentUser.is_authenticated.toString());

      if (Rimdian.currentUser.hmac) {
        target.href = Rimdian.updateURLParam(target.href, URLParams.user_external_id_hmac, Rimdian.currentUser.hmac);
      }
    },
    // the cart hash is a combination of public_url + products id + items variant id + items quantity
    _cartHash: function _cartHash(data) {
      var cartHash = data.public_url ? data.public_url : '';

      if (data.items && data.items.length > 0) {
        data.items.forEach(function (item) {
          cartHash = cartHash + item.product_external_id + (item.variant_external_id ? item.variant_external_id : '') + (item.quantity || '0');
        });
      }

      return _md(cartHash);
    },
    _wipeAll: function _wipeAll() {
      // create an alert and clear cookies and localtorage on confirmation
      if (window.confirm('Do you know what you are doing?')) {
        // clear cookies
        Rimdian.deleteCookie(Rimdian.config.namespace + 'device');
        Rimdian.deleteCookie(Rimdian.config.namespace + 'user');
        Rimdian.deleteCookie(Rimdian.config.namespace + 'session'); // clear localstorage

        Rimdian._localStorage.remove('items');

        Rimdian._localStorage.remove('dispatchQueue'); // reinitialize


        Rimdian.currentUser = undefined;
        Rimdian.currentDevice = undefined;
        Rimdian.currentSession = undefined;
        Rimdian.currentCart = undefined;
        Rimdian.currentPageview = undefined;
        Rimdian.isReady = false;

        Rimdian._onReady(Rimdian.config);
      }
    }
  };

  return Rimdian;

})();
//# sourceMappingURL=data:application/json;charset=utf-8;base64,eyJ2ZXJzaW9uIjozLCJmaWxlIjoic2RrLmpzIiwic291cmNlcyI6WyIuLi9ub2RlX21vZHVsZXMvQGJhYmVsL3J1bnRpbWUvaGVscGVycy9lc20vdHlwZW9mLmpzIiwiLi4vbm9kZV9tb2R1bGVzL3RzbGliL3RzbGliLmVzNi5qcyIsIi4uL25vZGVfbW9kdWxlcy9ibHVlaW1wLW1kNS9qcy9tZDUuanMiLCIuLi9ub2RlX21vZHVsZXMvcGFnZS1saWZlY3ljbGUvZGlzdC9saWZlY3ljbGUubWpzIiwiLi4vLi4vc3JjL3Nkay50cyJdLCJzb3VyY2VzQ29udGVudCI6WyJleHBvcnQgZGVmYXVsdCBmdW5jdGlvbiBfdHlwZW9mKG9iaikge1xuICBcIkBiYWJlbC9oZWxwZXJzIC0gdHlwZW9mXCI7XG5cbiAgcmV0dXJuIF90eXBlb2YgPSBcImZ1bmN0aW9uXCIgPT0gdHlwZW9mIFN5bWJvbCAmJiBcInN5bWJvbFwiID09IHR5cGVvZiBTeW1ib2wuaXRlcmF0b3IgPyBmdW5jdGlvbiAob2JqKSB7XG4gICAgcmV0dXJuIHR5cGVvZiBvYmo7XG4gIH0gOiBmdW5jdGlvbiAob2JqKSB7XG4gICAgcmV0dXJuIG9iaiAmJiBcImZ1bmN0aW9uXCIgPT0gdHlwZW9mIFN5bWJvbCAmJiBvYmouY29uc3RydWN0b3IgPT09IFN5bWJvbCAmJiBvYmogIT09IFN5bWJvbC5wcm90b3R5cGUgPyBcInN5bWJvbFwiIDogdHlwZW9mIG9iajtcbiAgfSwgX3R5cGVvZihvYmopO1xufSIsIi8qKioqKioqKioqKioqKioqKioqKioqKioqKioqKioqKioqKioqKioqKioqKioqKioqKioqKioqKioqKioqKioqKioqKioqKioqKioqKipcclxuQ29weXJpZ2h0IChjKSBNaWNyb3NvZnQgQ29ycG9yYXRpb24uXHJcblxyXG5QZXJtaXNzaW9uIHRvIHVzZSwgY29weSwgbW9kaWZ5LCBhbmQvb3IgZGlzdHJpYnV0ZSB0aGlzIHNvZnR3YXJlIGZvciBhbnlcclxucHVycG9zZSB3aXRoIG9yIHdpdGhvdXQgZmVlIGlzIGhlcmVieSBncmFudGVkLlxyXG5cclxuVEhFIFNPRlRXQVJFIElTIFBST1ZJREVEIFwiQVMgSVNcIiBBTkQgVEhFIEFVVEhPUiBESVNDTEFJTVMgQUxMIFdBUlJBTlRJRVMgV0lUSFxyXG5SRUdBUkQgVE8gVEhJUyBTT0ZUV0FSRSBJTkNMVURJTkcgQUxMIElNUExJRUQgV0FSUkFOVElFUyBPRiBNRVJDSEFOVEFCSUxJVFlcclxuQU5EIEZJVE5FU1MuIElOIE5PIEVWRU5UIFNIQUxMIFRIRSBBVVRIT1IgQkUgTElBQkxFIEZPUiBBTlkgU1BFQ0lBTCwgRElSRUNULFxyXG5JTkRJUkVDVCwgT1IgQ09OU0VRVUVOVElBTCBEQU1BR0VTIE9SIEFOWSBEQU1BR0VTIFdIQVRTT0VWRVIgUkVTVUxUSU5HIEZST01cclxuTE9TUyBPRiBVU0UsIERBVEEgT1IgUFJPRklUUywgV0hFVEhFUiBJTiBBTiBBQ1RJT04gT0YgQ09OVFJBQ1QsIE5FR0xJR0VOQ0UgT1JcclxuT1RIRVIgVE9SVElPVVMgQUNUSU9OLCBBUklTSU5HIE9VVCBPRiBPUiBJTiBDT05ORUNUSU9OIFdJVEggVEhFIFVTRSBPUlxyXG5QRVJGT1JNQU5DRSBPRiBUSElTIFNPRlRXQVJFLlxyXG4qKioqKioqKioqKioqKioqKioqKioqKioqKioqKioqKioqKioqKioqKioqKioqKioqKioqKioqKioqKioqKioqKioqKioqKioqKioqKiAqL1xyXG4vKiBnbG9iYWwgUmVmbGVjdCwgUHJvbWlzZSAqL1xyXG5cclxudmFyIGV4dGVuZFN0YXRpY3MgPSBmdW5jdGlvbihkLCBiKSB7XHJcbiAgICBleHRlbmRTdGF0aWNzID0gT2JqZWN0LnNldFByb3RvdHlwZU9mIHx8XHJcbiAgICAgICAgKHsgX19wcm90b19fOiBbXSB9IGluc3RhbmNlb2YgQXJyYXkgJiYgZnVuY3Rpb24gKGQsIGIpIHsgZC5fX3Byb3RvX18gPSBiOyB9KSB8fFxyXG4gICAgICAgIGZ1bmN0aW9uIChkLCBiKSB7IGZvciAodmFyIHAgaW4gYikgaWYgKE9iamVjdC5wcm90b3R5cGUuaGFzT3duUHJvcGVydHkuY2FsbChiLCBwKSkgZFtwXSA9IGJbcF07IH07XHJcbiAgICByZXR1cm4gZXh0ZW5kU3RhdGljcyhkLCBiKTtcclxufTtcclxuXHJcbmV4cG9ydCBmdW5jdGlvbiBfX2V4dGVuZHMoZCwgYikge1xyXG4gICAgaWYgKHR5cGVvZiBiICE9PSBcImZ1bmN0aW9uXCIgJiYgYiAhPT0gbnVsbClcclxuICAgICAgICB0aHJvdyBuZXcgVHlwZUVycm9yKFwiQ2xhc3MgZXh0ZW5kcyB2YWx1ZSBcIiArIFN0cmluZyhiKSArIFwiIGlzIG5vdCBhIGNvbnN0cnVjdG9yIG9yIG51bGxcIik7XHJcbiAgICBleHRlbmRTdGF0aWNzKGQsIGIpO1xyXG4gICAgZnVuY3Rpb24gX18oKSB7IHRoaXMuY29uc3RydWN0b3IgPSBkOyB9XHJcbiAgICBkLnByb3RvdHlwZSA9IGIgPT09IG51bGwgPyBPYmplY3QuY3JlYXRlKGIpIDogKF9fLnByb3RvdHlwZSA9IGIucHJvdG90eXBlLCBuZXcgX18oKSk7XHJcbn1cclxuXHJcbmV4cG9ydCB2YXIgX19hc3NpZ24gPSBmdW5jdGlvbigpIHtcclxuICAgIF9fYXNzaWduID0gT2JqZWN0LmFzc2lnbiB8fCBmdW5jdGlvbiBfX2Fzc2lnbih0KSB7XHJcbiAgICAgICAgZm9yICh2YXIgcywgaSA9IDEsIG4gPSBhcmd1bWVudHMubGVuZ3RoOyBpIDwgbjsgaSsrKSB7XHJcbiAgICAgICAgICAgIHMgPSBhcmd1bWVudHNbaV07XHJcbiAgICAgICAgICAgIGZvciAodmFyIHAgaW4gcykgaWYgKE9iamVjdC5wcm90b3R5cGUuaGFzT3duUHJvcGVydHkuY2FsbChzLCBwKSkgdFtwXSA9IHNbcF07XHJcbiAgICAgICAgfVxyXG4gICAgICAgIHJldHVybiB0O1xyXG4gICAgfVxyXG4gICAgcmV0dXJuIF9fYXNzaWduLmFwcGx5KHRoaXMsIGFyZ3VtZW50cyk7XHJcbn1cclxuXHJcbmV4cG9ydCBmdW5jdGlvbiBfX3Jlc3QocywgZSkge1xyXG4gICAgdmFyIHQgPSB7fTtcclxuICAgIGZvciAodmFyIHAgaW4gcykgaWYgKE9iamVjdC5wcm90b3R5cGUuaGFzT3duUHJvcGVydHkuY2FsbChzLCBwKSAmJiBlLmluZGV4T2YocCkgPCAwKVxyXG4gICAgICAgIHRbcF0gPSBzW3BdO1xyXG4gICAgaWYgKHMgIT0gbnVsbCAmJiB0eXBlb2YgT2JqZWN0LmdldE93blByb3BlcnR5U3ltYm9scyA9PT0gXCJmdW5jdGlvblwiKVxyXG4gICAgICAgIGZvciAodmFyIGkgPSAwLCBwID0gT2JqZWN0LmdldE93blByb3BlcnR5U3ltYm9scyhzKTsgaSA8IHAubGVuZ3RoOyBpKyspIHtcclxuICAgICAgICAgICAgaWYgKGUuaW5kZXhPZihwW2ldKSA8IDAgJiYgT2JqZWN0LnByb3RvdHlwZS5wcm9wZXJ0eUlzRW51bWVyYWJsZS5jYWxsKHMsIHBbaV0pKVxyXG4gICAgICAgICAgICAgICAgdFtwW2ldXSA9IHNbcFtpXV07XHJcbiAgICAgICAgfVxyXG4gICAgcmV0dXJuIHQ7XHJcbn1cclxuXHJcbmV4cG9ydCBmdW5jdGlvbiBfX2RlY29yYXRlKGRlY29yYXRvcnMsIHRhcmdldCwga2V5LCBkZXNjKSB7XHJcbiAgICB2YXIgYyA9IGFyZ3VtZW50cy5sZW5ndGgsIHIgPSBjIDwgMyA/IHRhcmdldCA6IGRlc2MgPT09IG51bGwgPyBkZXNjID0gT2JqZWN0LmdldE93blByb3BlcnR5RGVzY3JpcHRvcih0YXJnZXQsIGtleSkgOiBkZXNjLCBkO1xyXG4gICAgaWYgKHR5cGVvZiBSZWZsZWN0ID09PSBcIm9iamVjdFwiICYmIHR5cGVvZiBSZWZsZWN0LmRlY29yYXRlID09PSBcImZ1bmN0aW9uXCIpIHIgPSBSZWZsZWN0LmRlY29yYXRlKGRlY29yYXRvcnMsIHRhcmdldCwga2V5LCBkZXNjKTtcclxuICAgIGVsc2UgZm9yICh2YXIgaSA9IGRlY29yYXRvcnMubGVuZ3RoIC0gMTsgaSA+PSAwOyBpLS0pIGlmIChkID0gZGVjb3JhdG9yc1tpXSkgciA9IChjIDwgMyA/IGQocikgOiBjID4gMyA/IGQodGFyZ2V0LCBrZXksIHIpIDogZCh0YXJnZXQsIGtleSkpIHx8IHI7XHJcbiAgICByZXR1cm4gYyA+IDMgJiYgciAmJiBPYmplY3QuZGVmaW5lUHJvcGVydHkodGFyZ2V0LCBrZXksIHIpLCByO1xyXG59XHJcblxyXG5leHBvcnQgZnVuY3Rpb24gX19wYXJhbShwYXJhbUluZGV4LCBkZWNvcmF0b3IpIHtcclxuICAgIHJldHVybiBmdW5jdGlvbiAodGFyZ2V0LCBrZXkpIHsgZGVjb3JhdG9yKHRhcmdldCwga2V5LCBwYXJhbUluZGV4KTsgfVxyXG59XHJcblxyXG5leHBvcnQgZnVuY3Rpb24gX19tZXRhZGF0YShtZXRhZGF0YUtleSwgbWV0YWRhdGFWYWx1ZSkge1xyXG4gICAgaWYgKHR5cGVvZiBSZWZsZWN0ID09PSBcIm9iamVjdFwiICYmIHR5cGVvZiBSZWZsZWN0Lm1ldGFkYXRhID09PSBcImZ1bmN0aW9uXCIpIHJldHVybiBSZWZsZWN0Lm1ldGFkYXRhKG1ldGFkYXRhS2V5LCBtZXRhZGF0YVZhbHVlKTtcclxufVxyXG5cclxuZXhwb3J0IGZ1bmN0aW9uIF9fYXdhaXRlcih0aGlzQXJnLCBfYXJndW1lbnRzLCBQLCBnZW5lcmF0b3IpIHtcclxuICAgIGZ1bmN0aW9uIGFkb3B0KHZhbHVlKSB7IHJldHVybiB2YWx1ZSBpbnN0YW5jZW9mIFAgPyB2YWx1ZSA6IG5ldyBQKGZ1bmN0aW9uIChyZXNvbHZlKSB7IHJlc29sdmUodmFsdWUpOyB9KTsgfVxyXG4gICAgcmV0dXJuIG5ldyAoUCB8fCAoUCA9IFByb21pc2UpKShmdW5jdGlvbiAocmVzb2x2ZSwgcmVqZWN0KSB7XHJcbiAgICAgICAgZnVuY3Rpb24gZnVsZmlsbGVkKHZhbHVlKSB7IHRyeSB7IHN0ZXAoZ2VuZXJhdG9yLm5leHQodmFsdWUpKTsgfSBjYXRjaCAoZSkgeyByZWplY3QoZSk7IH0gfVxyXG4gICAgICAgIGZ1bmN0aW9uIHJlamVjdGVkKHZhbHVlKSB7IHRyeSB7IHN0ZXAoZ2VuZXJhdG9yW1widGhyb3dcIl0odmFsdWUpKTsgfSBjYXRjaCAoZSkgeyByZWplY3QoZSk7IH0gfVxyXG4gICAgICAgIGZ1bmN0aW9uIHN0ZXAocmVzdWx0KSB7IHJlc3VsdC5kb25lID8gcmVzb2x2ZShyZXN1bHQudmFsdWUpIDogYWRvcHQocmVzdWx0LnZhbHVlKS50aGVuKGZ1bGZpbGxlZCwgcmVqZWN0ZWQpOyB9XHJcbiAgICAgICAgc3RlcCgoZ2VuZXJhdG9yID0gZ2VuZXJhdG9yLmFwcGx5KHRoaXNBcmcsIF9hcmd1bWVudHMgfHwgW10pKS5uZXh0KCkpO1xyXG4gICAgfSk7XHJcbn1cclxuXHJcbmV4cG9ydCBmdW5jdGlvbiBfX2dlbmVyYXRvcih0aGlzQXJnLCBib2R5KSB7XHJcbiAgICB2YXIgXyA9IHsgbGFiZWw6IDAsIHNlbnQ6IGZ1bmN0aW9uKCkgeyBpZiAodFswXSAmIDEpIHRocm93IHRbMV07IHJldHVybiB0WzFdOyB9LCB0cnlzOiBbXSwgb3BzOiBbXSB9LCBmLCB5LCB0LCBnO1xyXG4gICAgcmV0dXJuIGcgPSB7IG5leHQ6IHZlcmIoMCksIFwidGhyb3dcIjogdmVyYigxKSwgXCJyZXR1cm5cIjogdmVyYigyKSB9LCB0eXBlb2YgU3ltYm9sID09PSBcImZ1bmN0aW9uXCIgJiYgKGdbU3ltYm9sLml0ZXJhdG9yXSA9IGZ1bmN0aW9uKCkgeyByZXR1cm4gdGhpczsgfSksIGc7XHJcbiAgICBmdW5jdGlvbiB2ZXJiKG4pIHsgcmV0dXJuIGZ1bmN0aW9uICh2KSB7IHJldHVybiBzdGVwKFtuLCB2XSk7IH07IH1cclxuICAgIGZ1bmN0aW9uIHN0ZXAob3ApIHtcclxuICAgICAgICBpZiAoZikgdGhyb3cgbmV3IFR5cGVFcnJvcihcIkdlbmVyYXRvciBpcyBhbHJlYWR5IGV4ZWN1dGluZy5cIik7XHJcbiAgICAgICAgd2hpbGUgKF8pIHRyeSB7XHJcbiAgICAgICAgICAgIGlmIChmID0gMSwgeSAmJiAodCA9IG9wWzBdICYgMiA/IHlbXCJyZXR1cm5cIl0gOiBvcFswXSA/IHlbXCJ0aHJvd1wiXSB8fCAoKHQgPSB5W1wicmV0dXJuXCJdKSAmJiB0LmNhbGwoeSksIDApIDogeS5uZXh0KSAmJiAhKHQgPSB0LmNhbGwoeSwgb3BbMV0pKS5kb25lKSByZXR1cm4gdDtcclxuICAgICAgICAgICAgaWYgKHkgPSAwLCB0KSBvcCA9IFtvcFswXSAmIDIsIHQudmFsdWVdO1xyXG4gICAgICAgICAgICBzd2l0Y2ggKG9wWzBdKSB7XHJcbiAgICAgICAgICAgICAgICBjYXNlIDA6IGNhc2UgMTogdCA9IG9wOyBicmVhaztcclxuICAgICAgICAgICAgICAgIGNhc2UgNDogXy5sYWJlbCsrOyByZXR1cm4geyB2YWx1ZTogb3BbMV0sIGRvbmU6IGZhbHNlIH07XHJcbiAgICAgICAgICAgICAgICBjYXNlIDU6IF8ubGFiZWwrKzsgeSA9IG9wWzFdOyBvcCA9IFswXTsgY29udGludWU7XHJcbiAgICAgICAgICAgICAgICBjYXNlIDc6IG9wID0gXy5vcHMucG9wKCk7IF8udHJ5cy5wb3AoKTsgY29udGludWU7XHJcbiAgICAgICAgICAgICAgICBkZWZhdWx0OlxyXG4gICAgICAgICAgICAgICAgICAgIGlmICghKHQgPSBfLnRyeXMsIHQgPSB0Lmxlbmd0aCA+IDAgJiYgdFt0Lmxlbmd0aCAtIDFdKSAmJiAob3BbMF0gPT09IDYgfHwgb3BbMF0gPT09IDIpKSB7IF8gPSAwOyBjb250aW51ZTsgfVxyXG4gICAgICAgICAgICAgICAgICAgIGlmIChvcFswXSA9PT0gMyAmJiAoIXQgfHwgKG9wWzFdID4gdFswXSAmJiBvcFsxXSA8IHRbM10pKSkgeyBfLmxhYmVsID0gb3BbMV07IGJyZWFrOyB9XHJcbiAgICAgICAgICAgICAgICAgICAgaWYgKG9wWzBdID09PSA2ICYmIF8ubGFiZWwgPCB0WzFdKSB7IF8ubGFiZWwgPSB0WzFdOyB0ID0gb3A7IGJyZWFrOyB9XHJcbiAgICAgICAgICAgICAgICAgICAgaWYgKHQgJiYgXy5sYWJlbCA8IHRbMl0pIHsgXy5sYWJlbCA9IHRbMl07IF8ub3BzLnB1c2gob3ApOyBicmVhazsgfVxyXG4gICAgICAgICAgICAgICAgICAgIGlmICh0WzJdKSBfLm9wcy5wb3AoKTtcclxuICAgICAgICAgICAgICAgICAgICBfLnRyeXMucG9wKCk7IGNvbnRpbnVlO1xyXG4gICAgICAgICAgICB9XHJcbiAgICAgICAgICAgIG9wID0gYm9keS5jYWxsKHRoaXNBcmcsIF8pO1xyXG4gICAgICAgIH0gY2F0Y2ggKGUpIHsgb3AgPSBbNiwgZV07IHkgPSAwOyB9IGZpbmFsbHkgeyBmID0gdCA9IDA7IH1cclxuICAgICAgICBpZiAob3BbMF0gJiA1KSB0aHJvdyBvcFsxXTsgcmV0dXJuIHsgdmFsdWU6IG9wWzBdID8gb3BbMV0gOiB2b2lkIDAsIGRvbmU6IHRydWUgfTtcclxuICAgIH1cclxufVxyXG5cclxuZXhwb3J0IHZhciBfX2NyZWF0ZUJpbmRpbmcgPSBPYmplY3QuY3JlYXRlID8gKGZ1bmN0aW9uKG8sIG0sIGssIGsyKSB7XHJcbiAgICBpZiAoazIgPT09IHVuZGVmaW5lZCkgazIgPSBrO1xyXG4gICAgdmFyIGRlc2MgPSBPYmplY3QuZ2V0T3duUHJvcGVydHlEZXNjcmlwdG9yKG0sIGspO1xyXG4gICAgaWYgKCFkZXNjIHx8IChcImdldFwiIGluIGRlc2MgPyAhbS5fX2VzTW9kdWxlIDogZGVzYy53cml0YWJsZSB8fCBkZXNjLmNvbmZpZ3VyYWJsZSkpIHtcclxuICAgICAgICBkZXNjID0geyBlbnVtZXJhYmxlOiB0cnVlLCBnZXQ6IGZ1bmN0aW9uKCkgeyByZXR1cm4gbVtrXTsgfSB9O1xyXG4gICAgfVxyXG4gICAgT2JqZWN0LmRlZmluZVByb3BlcnR5KG8sIGsyLCBkZXNjKTtcclxufSkgOiAoZnVuY3Rpb24obywgbSwgaywgazIpIHtcclxuICAgIGlmIChrMiA9PT0gdW5kZWZpbmVkKSBrMiA9IGs7XHJcbiAgICBvW2syXSA9IG1ba107XHJcbn0pO1xyXG5cclxuZXhwb3J0IGZ1bmN0aW9uIF9fZXhwb3J0U3RhcihtLCBvKSB7XHJcbiAgICBmb3IgKHZhciBwIGluIG0pIGlmIChwICE9PSBcImRlZmF1bHRcIiAmJiAhT2JqZWN0LnByb3RvdHlwZS5oYXNPd25Qcm9wZXJ0eS5jYWxsKG8sIHApKSBfX2NyZWF0ZUJpbmRpbmcobywgbSwgcCk7XHJcbn1cclxuXHJcbmV4cG9ydCBmdW5jdGlvbiBfX3ZhbHVlcyhvKSB7XHJcbiAgICB2YXIgcyA9IHR5cGVvZiBTeW1ib2wgPT09IFwiZnVuY3Rpb25cIiAmJiBTeW1ib2wuaXRlcmF0b3IsIG0gPSBzICYmIG9bc10sIGkgPSAwO1xyXG4gICAgaWYgKG0pIHJldHVybiBtLmNhbGwobyk7XHJcbiAgICBpZiAobyAmJiB0eXBlb2Ygby5sZW5ndGggPT09IFwibnVtYmVyXCIpIHJldHVybiB7XHJcbiAgICAgICAgbmV4dDogZnVuY3Rpb24gKCkge1xyXG4gICAgICAgICAgICBpZiAobyAmJiBpID49IG8ubGVuZ3RoKSBvID0gdm9pZCAwO1xyXG4gICAgICAgICAgICByZXR1cm4geyB2YWx1ZTogbyAmJiBvW2krK10sIGRvbmU6ICFvIH07XHJcbiAgICAgICAgfVxyXG4gICAgfTtcclxuICAgIHRocm93IG5ldyBUeXBlRXJyb3IocyA/IFwiT2JqZWN0IGlzIG5vdCBpdGVyYWJsZS5cIiA6IFwiU3ltYm9sLml0ZXJhdG9yIGlzIG5vdCBkZWZpbmVkLlwiKTtcclxufVxyXG5cclxuZXhwb3J0IGZ1bmN0aW9uIF9fcmVhZChvLCBuKSB7XHJcbiAgICB2YXIgbSA9IHR5cGVvZiBTeW1ib2wgPT09IFwiZnVuY3Rpb25cIiAmJiBvW1N5bWJvbC5pdGVyYXRvcl07XHJcbiAgICBpZiAoIW0pIHJldHVybiBvO1xyXG4gICAgdmFyIGkgPSBtLmNhbGwobyksIHIsIGFyID0gW10sIGU7XHJcbiAgICB0cnkge1xyXG4gICAgICAgIHdoaWxlICgobiA9PT0gdm9pZCAwIHx8IG4tLSA+IDApICYmICEociA9IGkubmV4dCgpKS5kb25lKSBhci5wdXNoKHIudmFsdWUpO1xyXG4gICAgfVxyXG4gICAgY2F0Y2ggKGVycm9yKSB7IGUgPSB7IGVycm9yOiBlcnJvciB9OyB9XHJcbiAgICBmaW5hbGx5IHtcclxuICAgICAgICB0cnkge1xyXG4gICAgICAgICAgICBpZiAociAmJiAhci5kb25lICYmIChtID0gaVtcInJldHVyblwiXSkpIG0uY2FsbChpKTtcclxuICAgICAgICB9XHJcbiAgICAgICAgZmluYWxseSB7IGlmIChlKSB0aHJvdyBlLmVycm9yOyB9XHJcbiAgICB9XHJcbiAgICByZXR1cm4gYXI7XHJcbn1cclxuXHJcbi8qKiBAZGVwcmVjYXRlZCAqL1xyXG5leHBvcnQgZnVuY3Rpb24gX19zcHJlYWQoKSB7XHJcbiAgICBmb3IgKHZhciBhciA9IFtdLCBpID0gMDsgaSA8IGFyZ3VtZW50cy5sZW5ndGg7IGkrKylcclxuICAgICAgICBhciA9IGFyLmNvbmNhdChfX3JlYWQoYXJndW1lbnRzW2ldKSk7XHJcbiAgICByZXR1cm4gYXI7XHJcbn1cclxuXHJcbi8qKiBAZGVwcmVjYXRlZCAqL1xyXG5leHBvcnQgZnVuY3Rpb24gX19zcHJlYWRBcnJheXMoKSB7XHJcbiAgICBmb3IgKHZhciBzID0gMCwgaSA9IDAsIGlsID0gYXJndW1lbnRzLmxlbmd0aDsgaSA8IGlsOyBpKyspIHMgKz0gYXJndW1lbnRzW2ldLmxlbmd0aDtcclxuICAgIGZvciAodmFyIHIgPSBBcnJheShzKSwgayA9IDAsIGkgPSAwOyBpIDwgaWw7IGkrKylcclxuICAgICAgICBmb3IgKHZhciBhID0gYXJndW1lbnRzW2ldLCBqID0gMCwgamwgPSBhLmxlbmd0aDsgaiA8IGpsOyBqKyssIGsrKylcclxuICAgICAgICAgICAgcltrXSA9IGFbal07XHJcbiAgICByZXR1cm4gcjtcclxufVxyXG5cclxuZXhwb3J0IGZ1bmN0aW9uIF9fc3ByZWFkQXJyYXkodG8sIGZyb20sIHBhY2spIHtcclxuICAgIGlmIChwYWNrIHx8IGFyZ3VtZW50cy5sZW5ndGggPT09IDIpIGZvciAodmFyIGkgPSAwLCBsID0gZnJvbS5sZW5ndGgsIGFyOyBpIDwgbDsgaSsrKSB7XHJcbiAgICAgICAgaWYgKGFyIHx8ICEoaSBpbiBmcm9tKSkge1xyXG4gICAgICAgICAgICBpZiAoIWFyKSBhciA9IEFycmF5LnByb3RvdHlwZS5zbGljZS5jYWxsKGZyb20sIDAsIGkpO1xyXG4gICAgICAgICAgICBhcltpXSA9IGZyb21baV07XHJcbiAgICAgICAgfVxyXG4gICAgfVxyXG4gICAgcmV0dXJuIHRvLmNvbmNhdChhciB8fCBBcnJheS5wcm90b3R5cGUuc2xpY2UuY2FsbChmcm9tKSk7XHJcbn1cclxuXHJcbmV4cG9ydCBmdW5jdGlvbiBfX2F3YWl0KHYpIHtcclxuICAgIHJldHVybiB0aGlzIGluc3RhbmNlb2YgX19hd2FpdCA/ICh0aGlzLnYgPSB2LCB0aGlzKSA6IG5ldyBfX2F3YWl0KHYpO1xyXG59XHJcblxyXG5leHBvcnQgZnVuY3Rpb24gX19hc3luY0dlbmVyYXRvcih0aGlzQXJnLCBfYXJndW1lbnRzLCBnZW5lcmF0b3IpIHtcclxuICAgIGlmICghU3ltYm9sLmFzeW5jSXRlcmF0b3IpIHRocm93IG5ldyBUeXBlRXJyb3IoXCJTeW1ib2wuYXN5bmNJdGVyYXRvciBpcyBub3QgZGVmaW5lZC5cIik7XHJcbiAgICB2YXIgZyA9IGdlbmVyYXRvci5hcHBseSh0aGlzQXJnLCBfYXJndW1lbnRzIHx8IFtdKSwgaSwgcSA9IFtdO1xyXG4gICAgcmV0dXJuIGkgPSB7fSwgdmVyYihcIm5leHRcIiksIHZlcmIoXCJ0aHJvd1wiKSwgdmVyYihcInJldHVyblwiKSwgaVtTeW1ib2wuYXN5bmNJdGVyYXRvcl0gPSBmdW5jdGlvbiAoKSB7IHJldHVybiB0aGlzOyB9LCBpO1xyXG4gICAgZnVuY3Rpb24gdmVyYihuKSB7IGlmIChnW25dKSBpW25dID0gZnVuY3Rpb24gKHYpIHsgcmV0dXJuIG5ldyBQcm9taXNlKGZ1bmN0aW9uIChhLCBiKSB7IHEucHVzaChbbiwgdiwgYSwgYl0pID4gMSB8fCByZXN1bWUobiwgdik7IH0pOyB9OyB9XHJcbiAgICBmdW5jdGlvbiByZXN1bWUobiwgdikgeyB0cnkgeyBzdGVwKGdbbl0odikpOyB9IGNhdGNoIChlKSB7IHNldHRsZShxWzBdWzNdLCBlKTsgfSB9XHJcbiAgICBmdW5jdGlvbiBzdGVwKHIpIHsgci52YWx1ZSBpbnN0YW5jZW9mIF9fYXdhaXQgPyBQcm9taXNlLnJlc29sdmUoci52YWx1ZS52KS50aGVuKGZ1bGZpbGwsIHJlamVjdCkgOiBzZXR0bGUocVswXVsyXSwgcik7IH1cclxuICAgIGZ1bmN0aW9uIGZ1bGZpbGwodmFsdWUpIHsgcmVzdW1lKFwibmV4dFwiLCB2YWx1ZSk7IH1cclxuICAgIGZ1bmN0aW9uIHJlamVjdCh2YWx1ZSkgeyByZXN1bWUoXCJ0aHJvd1wiLCB2YWx1ZSk7IH1cclxuICAgIGZ1bmN0aW9uIHNldHRsZShmLCB2KSB7IGlmIChmKHYpLCBxLnNoaWZ0KCksIHEubGVuZ3RoKSByZXN1bWUocVswXVswXSwgcVswXVsxXSk7IH1cclxufVxyXG5cclxuZXhwb3J0IGZ1bmN0aW9uIF9fYXN5bmNEZWxlZ2F0b3Iobykge1xyXG4gICAgdmFyIGksIHA7XHJcbiAgICByZXR1cm4gaSA9IHt9LCB2ZXJiKFwibmV4dFwiKSwgdmVyYihcInRocm93XCIsIGZ1bmN0aW9uIChlKSB7IHRocm93IGU7IH0pLCB2ZXJiKFwicmV0dXJuXCIpLCBpW1N5bWJvbC5pdGVyYXRvcl0gPSBmdW5jdGlvbiAoKSB7IHJldHVybiB0aGlzOyB9LCBpO1xyXG4gICAgZnVuY3Rpb24gdmVyYihuLCBmKSB7IGlbbl0gPSBvW25dID8gZnVuY3Rpb24gKHYpIHsgcmV0dXJuIChwID0gIXApID8geyB2YWx1ZTogX19hd2FpdChvW25dKHYpKSwgZG9uZTogbiA9PT0gXCJyZXR1cm5cIiB9IDogZiA/IGYodikgOiB2OyB9IDogZjsgfVxyXG59XHJcblxyXG5leHBvcnQgZnVuY3Rpb24gX19hc3luY1ZhbHVlcyhvKSB7XHJcbiAgICBpZiAoIVN5bWJvbC5hc3luY0l0ZXJhdG9yKSB0aHJvdyBuZXcgVHlwZUVycm9yKFwiU3ltYm9sLmFzeW5jSXRlcmF0b3IgaXMgbm90IGRlZmluZWQuXCIpO1xyXG4gICAgdmFyIG0gPSBvW1N5bWJvbC5hc3luY0l0ZXJhdG9yXSwgaTtcclxuICAgIHJldHVybiBtID8gbS5jYWxsKG8pIDogKG8gPSB0eXBlb2YgX192YWx1ZXMgPT09IFwiZnVuY3Rpb25cIiA/IF9fdmFsdWVzKG8pIDogb1tTeW1ib2wuaXRlcmF0b3JdKCksIGkgPSB7fSwgdmVyYihcIm5leHRcIiksIHZlcmIoXCJ0aHJvd1wiKSwgdmVyYihcInJldHVyblwiKSwgaVtTeW1ib2wuYXN5bmNJdGVyYXRvcl0gPSBmdW5jdGlvbiAoKSB7IHJldHVybiB0aGlzOyB9LCBpKTtcclxuICAgIGZ1bmN0aW9uIHZlcmIobikgeyBpW25dID0gb1tuXSAmJiBmdW5jdGlvbiAodikgeyByZXR1cm4gbmV3IFByb21pc2UoZnVuY3Rpb24gKHJlc29sdmUsIHJlamVjdCkgeyB2ID0gb1tuXSh2KSwgc2V0dGxlKHJlc29sdmUsIHJlamVjdCwgdi5kb25lLCB2LnZhbHVlKTsgfSk7IH07IH1cclxuICAgIGZ1bmN0aW9uIHNldHRsZShyZXNvbHZlLCByZWplY3QsIGQsIHYpIHsgUHJvbWlzZS5yZXNvbHZlKHYpLnRoZW4oZnVuY3Rpb24odikgeyByZXNvbHZlKHsgdmFsdWU6IHYsIGRvbmU6IGQgfSk7IH0sIHJlamVjdCk7IH1cclxufVxyXG5cclxuZXhwb3J0IGZ1bmN0aW9uIF9fbWFrZVRlbXBsYXRlT2JqZWN0KGNvb2tlZCwgcmF3KSB7XHJcbiAgICBpZiAoT2JqZWN0LmRlZmluZVByb3BlcnR5KSB7IE9iamVjdC5kZWZpbmVQcm9wZXJ0eShjb29rZWQsIFwicmF3XCIsIHsgdmFsdWU6IHJhdyB9KTsgfSBlbHNlIHsgY29va2VkLnJhdyA9IHJhdzsgfVxyXG4gICAgcmV0dXJuIGNvb2tlZDtcclxufTtcclxuXHJcbnZhciBfX3NldE1vZHVsZURlZmF1bHQgPSBPYmplY3QuY3JlYXRlID8gKGZ1bmN0aW9uKG8sIHYpIHtcclxuICAgIE9iamVjdC5kZWZpbmVQcm9wZXJ0eShvLCBcImRlZmF1bHRcIiwgeyBlbnVtZXJhYmxlOiB0cnVlLCB2YWx1ZTogdiB9KTtcclxufSkgOiBmdW5jdGlvbihvLCB2KSB7XHJcbiAgICBvW1wiZGVmYXVsdFwiXSA9IHY7XHJcbn07XHJcblxyXG5leHBvcnQgZnVuY3Rpb24gX19pbXBvcnRTdGFyKG1vZCkge1xyXG4gICAgaWYgKG1vZCAmJiBtb2QuX19lc01vZHVsZSkgcmV0dXJuIG1vZDtcclxuICAgIHZhciByZXN1bHQgPSB7fTtcclxuICAgIGlmIChtb2QgIT0gbnVsbCkgZm9yICh2YXIgayBpbiBtb2QpIGlmIChrICE9PSBcImRlZmF1bHRcIiAmJiBPYmplY3QucHJvdG90eXBlLmhhc093blByb3BlcnR5LmNhbGwobW9kLCBrKSkgX19jcmVhdGVCaW5kaW5nKHJlc3VsdCwgbW9kLCBrKTtcclxuICAgIF9fc2V0TW9kdWxlRGVmYXVsdChyZXN1bHQsIG1vZCk7XHJcbiAgICByZXR1cm4gcmVzdWx0O1xyXG59XHJcblxyXG5leHBvcnQgZnVuY3Rpb24gX19pbXBvcnREZWZhdWx0KG1vZCkge1xyXG4gICAgcmV0dXJuIChtb2QgJiYgbW9kLl9fZXNNb2R1bGUpID8gbW9kIDogeyBkZWZhdWx0OiBtb2QgfTtcclxufVxyXG5cclxuZXhwb3J0IGZ1bmN0aW9uIF9fY2xhc3NQcml2YXRlRmllbGRHZXQocmVjZWl2ZXIsIHN0YXRlLCBraW5kLCBmKSB7XHJcbiAgICBpZiAoa2luZCA9PT0gXCJhXCIgJiYgIWYpIHRocm93IG5ldyBUeXBlRXJyb3IoXCJQcml2YXRlIGFjY2Vzc29yIHdhcyBkZWZpbmVkIHdpdGhvdXQgYSBnZXR0ZXJcIik7XHJcbiAgICBpZiAodHlwZW9mIHN0YXRlID09PSBcImZ1bmN0aW9uXCIgPyByZWNlaXZlciAhPT0gc3RhdGUgfHwgIWYgOiAhc3RhdGUuaGFzKHJlY2VpdmVyKSkgdGhyb3cgbmV3IFR5cGVFcnJvcihcIkNhbm5vdCByZWFkIHByaXZhdGUgbWVtYmVyIGZyb20gYW4gb2JqZWN0IHdob3NlIGNsYXNzIGRpZCBub3QgZGVjbGFyZSBpdFwiKTtcclxuICAgIHJldHVybiBraW5kID09PSBcIm1cIiA/IGYgOiBraW5kID09PSBcImFcIiA/IGYuY2FsbChyZWNlaXZlcikgOiBmID8gZi52YWx1ZSA6IHN0YXRlLmdldChyZWNlaXZlcik7XHJcbn1cclxuXHJcbmV4cG9ydCBmdW5jdGlvbiBfX2NsYXNzUHJpdmF0ZUZpZWxkU2V0KHJlY2VpdmVyLCBzdGF0ZSwgdmFsdWUsIGtpbmQsIGYpIHtcclxuICAgIGlmIChraW5kID09PSBcIm1cIikgdGhyb3cgbmV3IFR5cGVFcnJvcihcIlByaXZhdGUgbWV0aG9kIGlzIG5vdCB3cml0YWJsZVwiKTtcclxuICAgIGlmIChraW5kID09PSBcImFcIiAmJiAhZikgdGhyb3cgbmV3IFR5cGVFcnJvcihcIlByaXZhdGUgYWNjZXNzb3Igd2FzIGRlZmluZWQgd2l0aG91dCBhIHNldHRlclwiKTtcclxuICAgIGlmICh0eXBlb2Ygc3RhdGUgPT09IFwiZnVuY3Rpb25cIiA/IHJlY2VpdmVyICE9PSBzdGF0ZSB8fCAhZiA6ICFzdGF0ZS5oYXMocmVjZWl2ZXIpKSB0aHJvdyBuZXcgVHlwZUVycm9yKFwiQ2Fubm90IHdyaXRlIHByaXZhdGUgbWVtYmVyIHRvIGFuIG9iamVjdCB3aG9zZSBjbGFzcyBkaWQgbm90IGRlY2xhcmUgaXRcIik7XHJcbiAgICByZXR1cm4gKGtpbmQgPT09IFwiYVwiID8gZi5jYWxsKHJlY2VpdmVyLCB2YWx1ZSkgOiBmID8gZi52YWx1ZSA9IHZhbHVlIDogc3RhdGUuc2V0KHJlY2VpdmVyLCB2YWx1ZSkpLCB2YWx1ZTtcclxufVxyXG5cclxuZXhwb3J0IGZ1bmN0aW9uIF9fY2xhc3NQcml2YXRlRmllbGRJbihzdGF0ZSwgcmVjZWl2ZXIpIHtcclxuICAgIGlmIChyZWNlaXZlciA9PT0gbnVsbCB8fCAodHlwZW9mIHJlY2VpdmVyICE9PSBcIm9iamVjdFwiICYmIHR5cGVvZiByZWNlaXZlciAhPT0gXCJmdW5jdGlvblwiKSkgdGhyb3cgbmV3IFR5cGVFcnJvcihcIkNhbm5vdCB1c2UgJ2luJyBvcGVyYXRvciBvbiBub24tb2JqZWN0XCIpO1xyXG4gICAgcmV0dXJuIHR5cGVvZiBzdGF0ZSA9PT0gXCJmdW5jdGlvblwiID8gcmVjZWl2ZXIgPT09IHN0YXRlIDogc3RhdGUuaGFzKHJlY2VpdmVyKTtcclxufVxyXG4iLCIvKlxuICogSmF2YVNjcmlwdCBNRDVcbiAqIGh0dHBzOi8vZ2l0aHViLmNvbS9ibHVlaW1wL0phdmFTY3JpcHQtTUQ1XG4gKlxuICogQ29weXJpZ2h0IDIwMTEsIFNlYmFzdGlhbiBUc2NoYW5cbiAqIGh0dHBzOi8vYmx1ZWltcC5uZXRcbiAqXG4gKiBMaWNlbnNlZCB1bmRlciB0aGUgTUlUIGxpY2Vuc2U6XG4gKiBodHRwczovL29wZW5zb3VyY2Uub3JnL2xpY2Vuc2VzL01JVFxuICpcbiAqIEJhc2VkIG9uXG4gKiBBIEphdmFTY3JpcHQgaW1wbGVtZW50YXRpb24gb2YgdGhlIFJTQSBEYXRhIFNlY3VyaXR5LCBJbmMuIE1ENSBNZXNzYWdlXG4gKiBEaWdlc3QgQWxnb3JpdGhtLCBhcyBkZWZpbmVkIGluIFJGQyAxMzIxLlxuICogVmVyc2lvbiAyLjIgQ29weXJpZ2h0IChDKSBQYXVsIEpvaG5zdG9uIDE5OTkgLSAyMDA5XG4gKiBPdGhlciBjb250cmlidXRvcnM6IEdyZWcgSG9sdCwgQW5kcmV3IEtlcGVydCwgWWRuYXIsIExvc3RpbmV0XG4gKiBEaXN0cmlidXRlZCB1bmRlciB0aGUgQlNEIExpY2Vuc2VcbiAqIFNlZSBodHRwOi8vcGFqaG9tZS5vcmcudWsvY3J5cHQvbWQ1IGZvciBtb3JlIGluZm8uXG4gKi9cblxuLyogZ2xvYmFsIGRlZmluZSAqL1xuXG4vKiBlc2xpbnQtZGlzYWJsZSBzdHJpY3QgKi9cblxuOyhmdW5jdGlvbiAoJCkge1xuICAndXNlIHN0cmljdCdcblxuICAvKipcbiAgICogQWRkIGludGVnZXJzLCB3cmFwcGluZyBhdCAyXjMyLlxuICAgKiBUaGlzIHVzZXMgMTYtYml0IG9wZXJhdGlvbnMgaW50ZXJuYWxseSB0byB3b3JrIGFyb3VuZCBidWdzIGluIGludGVycHJldGVycy5cbiAgICpcbiAgICogQHBhcmFtIHtudW1iZXJ9IHggRmlyc3QgaW50ZWdlclxuICAgKiBAcGFyYW0ge251bWJlcn0geSBTZWNvbmQgaW50ZWdlclxuICAgKiBAcmV0dXJucyB7bnVtYmVyfSBTdW1cbiAgICovXG4gIGZ1bmN0aW9uIHNhZmVBZGQoeCwgeSkge1xuICAgIHZhciBsc3cgPSAoeCAmIDB4ZmZmZikgKyAoeSAmIDB4ZmZmZilcbiAgICB2YXIgbXN3ID0gKHggPj4gMTYpICsgKHkgPj4gMTYpICsgKGxzdyA+PiAxNilcbiAgICByZXR1cm4gKG1zdyA8PCAxNikgfCAobHN3ICYgMHhmZmZmKVxuICB9XG5cbiAgLyoqXG4gICAqIEJpdHdpc2Ugcm90YXRlIGEgMzItYml0IG51bWJlciB0byB0aGUgbGVmdC5cbiAgICpcbiAgICogQHBhcmFtIHtudW1iZXJ9IG51bSAzMi1iaXQgbnVtYmVyXG4gICAqIEBwYXJhbSB7bnVtYmVyfSBjbnQgUm90YXRpb24gY291bnRcbiAgICogQHJldHVybnMge251bWJlcn0gUm90YXRlZCBudW1iZXJcbiAgICovXG4gIGZ1bmN0aW9uIGJpdFJvdGF0ZUxlZnQobnVtLCBjbnQpIHtcbiAgICByZXR1cm4gKG51bSA8PCBjbnQpIHwgKG51bSA+Pj4gKDMyIC0gY250KSlcbiAgfVxuXG4gIC8qKlxuICAgKiBCYXNpYyBvcGVyYXRpb24gdGhlIGFsZ29yaXRobSB1c2VzLlxuICAgKlxuICAgKiBAcGFyYW0ge251bWJlcn0gcSBxXG4gICAqIEBwYXJhbSB7bnVtYmVyfSBhIGFcbiAgICogQHBhcmFtIHtudW1iZXJ9IGIgYlxuICAgKiBAcGFyYW0ge251bWJlcn0geCB4XG4gICAqIEBwYXJhbSB7bnVtYmVyfSBzIHNcbiAgICogQHBhcmFtIHtudW1iZXJ9IHQgdFxuICAgKiBAcmV0dXJucyB7bnVtYmVyfSBSZXN1bHRcbiAgICovXG4gIGZ1bmN0aW9uIG1kNWNtbihxLCBhLCBiLCB4LCBzLCB0KSB7XG4gICAgcmV0dXJuIHNhZmVBZGQoYml0Um90YXRlTGVmdChzYWZlQWRkKHNhZmVBZGQoYSwgcSksIHNhZmVBZGQoeCwgdCkpLCBzKSwgYilcbiAgfVxuICAvKipcbiAgICogQmFzaWMgb3BlcmF0aW9uIHRoZSBhbGdvcml0aG0gdXNlcy5cbiAgICpcbiAgICogQHBhcmFtIHtudW1iZXJ9IGEgYVxuICAgKiBAcGFyYW0ge251bWJlcn0gYiBiXG4gICAqIEBwYXJhbSB7bnVtYmVyfSBjIGNcbiAgICogQHBhcmFtIHtudW1iZXJ9IGQgZFxuICAgKiBAcGFyYW0ge251bWJlcn0geCB4XG4gICAqIEBwYXJhbSB7bnVtYmVyfSBzIHNcbiAgICogQHBhcmFtIHtudW1iZXJ9IHQgdFxuICAgKiBAcmV0dXJucyB7bnVtYmVyfSBSZXN1bHRcbiAgICovXG4gIGZ1bmN0aW9uIG1kNWZmKGEsIGIsIGMsIGQsIHgsIHMsIHQpIHtcbiAgICByZXR1cm4gbWQ1Y21uKChiICYgYykgfCAofmIgJiBkKSwgYSwgYiwgeCwgcywgdClcbiAgfVxuICAvKipcbiAgICogQmFzaWMgb3BlcmF0aW9uIHRoZSBhbGdvcml0aG0gdXNlcy5cbiAgICpcbiAgICogQHBhcmFtIHtudW1iZXJ9IGEgYVxuICAgKiBAcGFyYW0ge251bWJlcn0gYiBiXG4gICAqIEBwYXJhbSB7bnVtYmVyfSBjIGNcbiAgICogQHBhcmFtIHtudW1iZXJ9IGQgZFxuICAgKiBAcGFyYW0ge251bWJlcn0geCB4XG4gICAqIEBwYXJhbSB7bnVtYmVyfSBzIHNcbiAgICogQHBhcmFtIHtudW1iZXJ9IHQgdFxuICAgKiBAcmV0dXJucyB7bnVtYmVyfSBSZXN1bHRcbiAgICovXG4gIGZ1bmN0aW9uIG1kNWdnKGEsIGIsIGMsIGQsIHgsIHMsIHQpIHtcbiAgICByZXR1cm4gbWQ1Y21uKChiICYgZCkgfCAoYyAmIH5kKSwgYSwgYiwgeCwgcywgdClcbiAgfVxuICAvKipcbiAgICogQmFzaWMgb3BlcmF0aW9uIHRoZSBhbGdvcml0aG0gdXNlcy5cbiAgICpcbiAgICogQHBhcmFtIHtudW1iZXJ9IGEgYVxuICAgKiBAcGFyYW0ge251bWJlcn0gYiBiXG4gICAqIEBwYXJhbSB7bnVtYmVyfSBjIGNcbiAgICogQHBhcmFtIHtudW1iZXJ9IGQgZFxuICAgKiBAcGFyYW0ge251bWJlcn0geCB4XG4gICAqIEBwYXJhbSB7bnVtYmVyfSBzIHNcbiAgICogQHBhcmFtIHtudW1iZXJ9IHQgdFxuICAgKiBAcmV0dXJucyB7bnVtYmVyfSBSZXN1bHRcbiAgICovXG4gIGZ1bmN0aW9uIG1kNWhoKGEsIGIsIGMsIGQsIHgsIHMsIHQpIHtcbiAgICByZXR1cm4gbWQ1Y21uKGIgXiBjIF4gZCwgYSwgYiwgeCwgcywgdClcbiAgfVxuICAvKipcbiAgICogQmFzaWMgb3BlcmF0aW9uIHRoZSBhbGdvcml0aG0gdXNlcy5cbiAgICpcbiAgICogQHBhcmFtIHtudW1iZXJ9IGEgYVxuICAgKiBAcGFyYW0ge251bWJlcn0gYiBiXG4gICAqIEBwYXJhbSB7bnVtYmVyfSBjIGNcbiAgICogQHBhcmFtIHtudW1iZXJ9IGQgZFxuICAgKiBAcGFyYW0ge251bWJlcn0geCB4XG4gICAqIEBwYXJhbSB7bnVtYmVyfSBzIHNcbiAgICogQHBhcmFtIHtudW1iZXJ9IHQgdFxuICAgKiBAcmV0dXJucyB7bnVtYmVyfSBSZXN1bHRcbiAgICovXG4gIGZ1bmN0aW9uIG1kNWlpKGEsIGIsIGMsIGQsIHgsIHMsIHQpIHtcbiAgICByZXR1cm4gbWQ1Y21uKGMgXiAoYiB8IH5kKSwgYSwgYiwgeCwgcywgdClcbiAgfVxuXG4gIC8qKlxuICAgKiBDYWxjdWxhdGUgdGhlIE1ENSBvZiBhbiBhcnJheSBvZiBsaXR0bGUtZW5kaWFuIHdvcmRzLCBhbmQgYSBiaXQgbGVuZ3RoLlxuICAgKlxuICAgKiBAcGFyYW0ge0FycmF5fSB4IEFycmF5IG9mIGxpdHRsZS1lbmRpYW4gd29yZHNcbiAgICogQHBhcmFtIHtudW1iZXJ9IGxlbiBCaXQgbGVuZ3RoXG4gICAqIEByZXR1cm5zIHtBcnJheTxudW1iZXI+fSBNRDUgQXJyYXlcbiAgICovXG4gIGZ1bmN0aW9uIGJpbmxNRDUoeCwgbGVuKSB7XG4gICAgLyogYXBwZW5kIHBhZGRpbmcgKi9cbiAgICB4W2xlbiA+PiA1XSB8PSAweDgwIDw8IGxlbiAlIDMyXG4gICAgeFsoKChsZW4gKyA2NCkgPj4+IDkpIDw8IDQpICsgMTRdID0gbGVuXG5cbiAgICB2YXIgaVxuICAgIHZhciBvbGRhXG4gICAgdmFyIG9sZGJcbiAgICB2YXIgb2xkY1xuICAgIHZhciBvbGRkXG4gICAgdmFyIGEgPSAxNzMyNTg0MTkzXG4gICAgdmFyIGIgPSAtMjcxNzMzODc5XG4gICAgdmFyIGMgPSAtMTczMjU4NDE5NFxuICAgIHZhciBkID0gMjcxNzMzODc4XG5cbiAgICBmb3IgKGkgPSAwOyBpIDwgeC5sZW5ndGg7IGkgKz0gMTYpIHtcbiAgICAgIG9sZGEgPSBhXG4gICAgICBvbGRiID0gYlxuICAgICAgb2xkYyA9IGNcbiAgICAgIG9sZGQgPSBkXG5cbiAgICAgIGEgPSBtZDVmZihhLCBiLCBjLCBkLCB4W2ldLCA3LCAtNjgwODc2OTM2KVxuICAgICAgZCA9IG1kNWZmKGQsIGEsIGIsIGMsIHhbaSArIDFdLCAxMiwgLTM4OTU2NDU4NilcbiAgICAgIGMgPSBtZDVmZihjLCBkLCBhLCBiLCB4W2kgKyAyXSwgMTcsIDYwNjEwNTgxOSlcbiAgICAgIGIgPSBtZDVmZihiLCBjLCBkLCBhLCB4W2kgKyAzXSwgMjIsIC0xMDQ0NTI1MzMwKVxuICAgICAgYSA9IG1kNWZmKGEsIGIsIGMsIGQsIHhbaSArIDRdLCA3LCAtMTc2NDE4ODk3KVxuICAgICAgZCA9IG1kNWZmKGQsIGEsIGIsIGMsIHhbaSArIDVdLCAxMiwgMTIwMDA4MDQyNilcbiAgICAgIGMgPSBtZDVmZihjLCBkLCBhLCBiLCB4W2kgKyA2XSwgMTcsIC0xNDczMjMxMzQxKVxuICAgICAgYiA9IG1kNWZmKGIsIGMsIGQsIGEsIHhbaSArIDddLCAyMiwgLTQ1NzA1OTgzKVxuICAgICAgYSA9IG1kNWZmKGEsIGIsIGMsIGQsIHhbaSArIDhdLCA3LCAxNzcwMDM1NDE2KVxuICAgICAgZCA9IG1kNWZmKGQsIGEsIGIsIGMsIHhbaSArIDldLCAxMiwgLTE5NTg0MTQ0MTcpXG4gICAgICBjID0gbWQ1ZmYoYywgZCwgYSwgYiwgeFtpICsgMTBdLCAxNywgLTQyMDYzKVxuICAgICAgYiA9IG1kNWZmKGIsIGMsIGQsIGEsIHhbaSArIDExXSwgMjIsIC0xOTkwNDA0MTYyKVxuICAgICAgYSA9IG1kNWZmKGEsIGIsIGMsIGQsIHhbaSArIDEyXSwgNywgMTgwNDYwMzY4MilcbiAgICAgIGQgPSBtZDVmZihkLCBhLCBiLCBjLCB4W2kgKyAxM10sIDEyLCAtNDAzNDExMDEpXG4gICAgICBjID0gbWQ1ZmYoYywgZCwgYSwgYiwgeFtpICsgMTRdLCAxNywgLTE1MDIwMDIyOTApXG4gICAgICBiID0gbWQ1ZmYoYiwgYywgZCwgYSwgeFtpICsgMTVdLCAyMiwgMTIzNjUzNTMyOSlcblxuICAgICAgYSA9IG1kNWdnKGEsIGIsIGMsIGQsIHhbaSArIDFdLCA1LCAtMTY1Nzk2NTEwKVxuICAgICAgZCA9IG1kNWdnKGQsIGEsIGIsIGMsIHhbaSArIDZdLCA5LCAtMTA2OTUwMTYzMilcbiAgICAgIGMgPSBtZDVnZyhjLCBkLCBhLCBiLCB4W2kgKyAxMV0sIDE0LCA2NDM3MTc3MTMpXG4gICAgICBiID0gbWQ1Z2coYiwgYywgZCwgYSwgeFtpXSwgMjAsIC0zNzM4OTczMDIpXG4gICAgICBhID0gbWQ1Z2coYSwgYiwgYywgZCwgeFtpICsgNV0sIDUsIC03MDE1NTg2OTEpXG4gICAgICBkID0gbWQ1Z2coZCwgYSwgYiwgYywgeFtpICsgMTBdLCA5LCAzODAxNjA4MylcbiAgICAgIGMgPSBtZDVnZyhjLCBkLCBhLCBiLCB4W2kgKyAxNV0sIDE0LCAtNjYwNDc4MzM1KVxuICAgICAgYiA9IG1kNWdnKGIsIGMsIGQsIGEsIHhbaSArIDRdLCAyMCwgLTQwNTUzNzg0OClcbiAgICAgIGEgPSBtZDVnZyhhLCBiLCBjLCBkLCB4W2kgKyA5XSwgNSwgNTY4NDQ2NDM4KVxuICAgICAgZCA9IG1kNWdnKGQsIGEsIGIsIGMsIHhbaSArIDE0XSwgOSwgLTEwMTk4MDM2OTApXG4gICAgICBjID0gbWQ1Z2coYywgZCwgYSwgYiwgeFtpICsgM10sIDE0LCAtMTg3MzYzOTYxKVxuICAgICAgYiA9IG1kNWdnKGIsIGMsIGQsIGEsIHhbaSArIDhdLCAyMCwgMTE2MzUzMTUwMSlcbiAgICAgIGEgPSBtZDVnZyhhLCBiLCBjLCBkLCB4W2kgKyAxM10sIDUsIC0xNDQ0NjgxNDY3KVxuICAgICAgZCA9IG1kNWdnKGQsIGEsIGIsIGMsIHhbaSArIDJdLCA5LCAtNTE0MDM3ODQpXG4gICAgICBjID0gbWQ1Z2coYywgZCwgYSwgYiwgeFtpICsgN10sIDE0LCAxNzM1MzI4NDczKVxuICAgICAgYiA9IG1kNWdnKGIsIGMsIGQsIGEsIHhbaSArIDEyXSwgMjAsIC0xOTI2NjA3NzM0KVxuXG4gICAgICBhID0gbWQ1aGgoYSwgYiwgYywgZCwgeFtpICsgNV0sIDQsIC0zNzg1NTgpXG4gICAgICBkID0gbWQ1aGgoZCwgYSwgYiwgYywgeFtpICsgOF0sIDExLCAtMjAyMjU3NDQ2MylcbiAgICAgIGMgPSBtZDVoaChjLCBkLCBhLCBiLCB4W2kgKyAxMV0sIDE2LCAxODM5MDMwNTYyKVxuICAgICAgYiA9IG1kNWhoKGIsIGMsIGQsIGEsIHhbaSArIDE0XSwgMjMsIC0zNTMwOTU1NilcbiAgICAgIGEgPSBtZDVoaChhLCBiLCBjLCBkLCB4W2kgKyAxXSwgNCwgLTE1MzA5OTIwNjApXG4gICAgICBkID0gbWQ1aGgoZCwgYSwgYiwgYywgeFtpICsgNF0sIDExLCAxMjcyODkzMzUzKVxuICAgICAgYyA9IG1kNWhoKGMsIGQsIGEsIGIsIHhbaSArIDddLCAxNiwgLTE1NTQ5NzYzMilcbiAgICAgIGIgPSBtZDVoaChiLCBjLCBkLCBhLCB4W2kgKyAxMF0sIDIzLCAtMTA5NDczMDY0MClcbiAgICAgIGEgPSBtZDVoaChhLCBiLCBjLCBkLCB4W2kgKyAxM10sIDQsIDY4MTI3OTE3NClcbiAgICAgIGQgPSBtZDVoaChkLCBhLCBiLCBjLCB4W2ldLCAxMSwgLTM1ODUzNzIyMilcbiAgICAgIGMgPSBtZDVoaChjLCBkLCBhLCBiLCB4W2kgKyAzXSwgMTYsIC03MjI1MjE5NzkpXG4gICAgICBiID0gbWQ1aGgoYiwgYywgZCwgYSwgeFtpICsgNl0sIDIzLCA3NjAyOTE4OSlcbiAgICAgIGEgPSBtZDVoaChhLCBiLCBjLCBkLCB4W2kgKyA5XSwgNCwgLTY0MDM2NDQ4NylcbiAgICAgIGQgPSBtZDVoaChkLCBhLCBiLCBjLCB4W2kgKyAxMl0sIDExLCAtNDIxODE1ODM1KVxuICAgICAgYyA9IG1kNWhoKGMsIGQsIGEsIGIsIHhbaSArIDE1XSwgMTYsIDUzMDc0MjUyMClcbiAgICAgIGIgPSBtZDVoaChiLCBjLCBkLCBhLCB4W2kgKyAyXSwgMjMsIC05OTUzMzg2NTEpXG5cbiAgICAgIGEgPSBtZDVpaShhLCBiLCBjLCBkLCB4W2ldLCA2LCAtMTk4NjMwODQ0KVxuICAgICAgZCA9IG1kNWlpKGQsIGEsIGIsIGMsIHhbaSArIDddLCAxMCwgMTEyNjg5MTQxNSlcbiAgICAgIGMgPSBtZDVpaShjLCBkLCBhLCBiLCB4W2kgKyAxNF0sIDE1LCAtMTQxNjM1NDkwNSlcbiAgICAgIGIgPSBtZDVpaShiLCBjLCBkLCBhLCB4W2kgKyA1XSwgMjEsIC01NzQzNDA1NSlcbiAgICAgIGEgPSBtZDVpaShhLCBiLCBjLCBkLCB4W2kgKyAxMl0sIDYsIDE3MDA0ODU1NzEpXG4gICAgICBkID0gbWQ1aWkoZCwgYSwgYiwgYywgeFtpICsgM10sIDEwLCAtMTg5NDk4NjYwNilcbiAgICAgIGMgPSBtZDVpaShjLCBkLCBhLCBiLCB4W2kgKyAxMF0sIDE1LCAtMTA1MTUyMylcbiAgICAgIGIgPSBtZDVpaShiLCBjLCBkLCBhLCB4W2kgKyAxXSwgMjEsIC0yMDU0OTIyNzk5KVxuICAgICAgYSA9IG1kNWlpKGEsIGIsIGMsIGQsIHhbaSArIDhdLCA2LCAxODczMzEzMzU5KVxuICAgICAgZCA9IG1kNWlpKGQsIGEsIGIsIGMsIHhbaSArIDE1XSwgMTAsIC0zMDYxMTc0NClcbiAgICAgIGMgPSBtZDVpaShjLCBkLCBhLCBiLCB4W2kgKyA2XSwgMTUsIC0xNTYwMTk4MzgwKVxuICAgICAgYiA9IG1kNWlpKGIsIGMsIGQsIGEsIHhbaSArIDEzXSwgMjEsIDEzMDkxNTE2NDkpXG4gICAgICBhID0gbWQ1aWkoYSwgYiwgYywgZCwgeFtpICsgNF0sIDYsIC0xNDU1MjMwNzApXG4gICAgICBkID0gbWQ1aWkoZCwgYSwgYiwgYywgeFtpICsgMTFdLCAxMCwgLTExMjAyMTAzNzkpXG4gICAgICBjID0gbWQ1aWkoYywgZCwgYSwgYiwgeFtpICsgMl0sIDE1LCA3MTg3ODcyNTkpXG4gICAgICBiID0gbWQ1aWkoYiwgYywgZCwgYSwgeFtpICsgOV0sIDIxLCAtMzQzNDg1NTUxKVxuXG4gICAgICBhID0gc2FmZUFkZChhLCBvbGRhKVxuICAgICAgYiA9IHNhZmVBZGQoYiwgb2xkYilcbiAgICAgIGMgPSBzYWZlQWRkKGMsIG9sZGMpXG4gICAgICBkID0gc2FmZUFkZChkLCBvbGRkKVxuICAgIH1cbiAgICByZXR1cm4gW2EsIGIsIGMsIGRdXG4gIH1cblxuICAvKipcbiAgICogQ29udmVydCBhbiBhcnJheSBvZiBsaXR0bGUtZW5kaWFuIHdvcmRzIHRvIGEgc3RyaW5nXG4gICAqXG4gICAqIEBwYXJhbSB7QXJyYXk8bnVtYmVyPn0gaW5wdXQgTUQ1IEFycmF5XG4gICAqIEByZXR1cm5zIHtzdHJpbmd9IE1ENSBzdHJpbmdcbiAgICovXG4gIGZ1bmN0aW9uIGJpbmwycnN0cihpbnB1dCkge1xuICAgIHZhciBpXG4gICAgdmFyIG91dHB1dCA9ICcnXG4gICAgdmFyIGxlbmd0aDMyID0gaW5wdXQubGVuZ3RoICogMzJcbiAgICBmb3IgKGkgPSAwOyBpIDwgbGVuZ3RoMzI7IGkgKz0gOCkge1xuICAgICAgb3V0cHV0ICs9IFN0cmluZy5mcm9tQ2hhckNvZGUoKGlucHV0W2kgPj4gNV0gPj4+IGkgJSAzMikgJiAweGZmKVxuICAgIH1cbiAgICByZXR1cm4gb3V0cHV0XG4gIH1cblxuICAvKipcbiAgICogQ29udmVydCBhIHJhdyBzdHJpbmcgdG8gYW4gYXJyYXkgb2YgbGl0dGxlLWVuZGlhbiB3b3Jkc1xuICAgKiBDaGFyYWN0ZXJzID4yNTUgaGF2ZSB0aGVpciBoaWdoLWJ5dGUgc2lsZW50bHkgaWdub3JlZC5cbiAgICpcbiAgICogQHBhcmFtIHtzdHJpbmd9IGlucHV0IFJhdyBpbnB1dCBzdHJpbmdcbiAgICogQHJldHVybnMge0FycmF5PG51bWJlcj59IEFycmF5IG9mIGxpdHRsZS1lbmRpYW4gd29yZHNcbiAgICovXG4gIGZ1bmN0aW9uIHJzdHIyYmlubChpbnB1dCkge1xuICAgIHZhciBpXG4gICAgdmFyIG91dHB1dCA9IFtdXG4gICAgb3V0cHV0WyhpbnB1dC5sZW5ndGggPj4gMikgLSAxXSA9IHVuZGVmaW5lZFxuICAgIGZvciAoaSA9IDA7IGkgPCBvdXRwdXQubGVuZ3RoOyBpICs9IDEpIHtcbiAgICAgIG91dHB1dFtpXSA9IDBcbiAgICB9XG4gICAgdmFyIGxlbmd0aDggPSBpbnB1dC5sZW5ndGggKiA4XG4gICAgZm9yIChpID0gMDsgaSA8IGxlbmd0aDg7IGkgKz0gOCkge1xuICAgICAgb3V0cHV0W2kgPj4gNV0gfD0gKGlucHV0LmNoYXJDb2RlQXQoaSAvIDgpICYgMHhmZikgPDwgaSAlIDMyXG4gICAgfVxuICAgIHJldHVybiBvdXRwdXRcbiAgfVxuXG4gIC8qKlxuICAgKiBDYWxjdWxhdGUgdGhlIE1ENSBvZiBhIHJhdyBzdHJpbmdcbiAgICpcbiAgICogQHBhcmFtIHtzdHJpbmd9IHMgSW5wdXQgc3RyaW5nXG4gICAqIEByZXR1cm5zIHtzdHJpbmd9IFJhdyBNRDUgc3RyaW5nXG4gICAqL1xuICBmdW5jdGlvbiByc3RyTUQ1KHMpIHtcbiAgICByZXR1cm4gYmlubDJyc3RyKGJpbmxNRDUocnN0cjJiaW5sKHMpLCBzLmxlbmd0aCAqIDgpKVxuICB9XG5cbiAgLyoqXG4gICAqIENhbGN1bGF0ZXMgdGhlIEhNQUMtTUQ1IG9mIGEga2V5IGFuZCBzb21lIGRhdGEgKHJhdyBzdHJpbmdzKVxuICAgKlxuICAgKiBAcGFyYW0ge3N0cmluZ30ga2V5IEhNQUMga2V5XG4gICAqIEBwYXJhbSB7c3RyaW5nfSBkYXRhIFJhdyBpbnB1dCBzdHJpbmdcbiAgICogQHJldHVybnMge3N0cmluZ30gUmF3IE1ENSBzdHJpbmdcbiAgICovXG4gIGZ1bmN0aW9uIHJzdHJITUFDTUQ1KGtleSwgZGF0YSkge1xuICAgIHZhciBpXG4gICAgdmFyIGJrZXkgPSByc3RyMmJpbmwoa2V5KVxuICAgIHZhciBpcGFkID0gW11cbiAgICB2YXIgb3BhZCA9IFtdXG4gICAgdmFyIGhhc2hcbiAgICBpcGFkWzE1XSA9IG9wYWRbMTVdID0gdW5kZWZpbmVkXG4gICAgaWYgKGJrZXkubGVuZ3RoID4gMTYpIHtcbiAgICAgIGJrZXkgPSBiaW5sTUQ1KGJrZXksIGtleS5sZW5ndGggKiA4KVxuICAgIH1cbiAgICBmb3IgKGkgPSAwOyBpIDwgMTY7IGkgKz0gMSkge1xuICAgICAgaXBhZFtpXSA9IGJrZXlbaV0gXiAweDM2MzYzNjM2XG4gICAgICBvcGFkW2ldID0gYmtleVtpXSBeIDB4NWM1YzVjNWNcbiAgICB9XG4gICAgaGFzaCA9IGJpbmxNRDUoaXBhZC5jb25jYXQocnN0cjJiaW5sKGRhdGEpKSwgNTEyICsgZGF0YS5sZW5ndGggKiA4KVxuICAgIHJldHVybiBiaW5sMnJzdHIoYmlubE1ENShvcGFkLmNvbmNhdChoYXNoKSwgNTEyICsgMTI4KSlcbiAgfVxuXG4gIC8qKlxuICAgKiBDb252ZXJ0IGEgcmF3IHN0cmluZyB0byBhIGhleCBzdHJpbmdcbiAgICpcbiAgICogQHBhcmFtIHtzdHJpbmd9IGlucHV0IFJhdyBpbnB1dCBzdHJpbmdcbiAgICogQHJldHVybnMge3N0cmluZ30gSGV4IGVuY29kZWQgc3RyaW5nXG4gICAqL1xuICBmdW5jdGlvbiByc3RyMmhleChpbnB1dCkge1xuICAgIHZhciBoZXhUYWIgPSAnMDEyMzQ1Njc4OWFiY2RlZidcbiAgICB2YXIgb3V0cHV0ID0gJydcbiAgICB2YXIgeFxuICAgIHZhciBpXG4gICAgZm9yIChpID0gMDsgaSA8IGlucHV0Lmxlbmd0aDsgaSArPSAxKSB7XG4gICAgICB4ID0gaW5wdXQuY2hhckNvZGVBdChpKVxuICAgICAgb3V0cHV0ICs9IGhleFRhYi5jaGFyQXQoKHggPj4+IDQpICYgMHgwZikgKyBoZXhUYWIuY2hhckF0KHggJiAweDBmKVxuICAgIH1cbiAgICByZXR1cm4gb3V0cHV0XG4gIH1cblxuICAvKipcbiAgICogRW5jb2RlIGEgc3RyaW5nIGFzIFVURi04XG4gICAqXG4gICAqIEBwYXJhbSB7c3RyaW5nfSBpbnB1dCBJbnB1dCBzdHJpbmdcbiAgICogQHJldHVybnMge3N0cmluZ30gVVRGOCBzdHJpbmdcbiAgICovXG4gIGZ1bmN0aW9uIHN0cjJyc3RyVVRGOChpbnB1dCkge1xuICAgIHJldHVybiB1bmVzY2FwZShlbmNvZGVVUklDb21wb25lbnQoaW5wdXQpKVxuICB9XG5cbiAgLyoqXG4gICAqIEVuY29kZXMgaW5wdXQgc3RyaW5nIGFzIHJhdyBNRDUgc3RyaW5nXG4gICAqXG4gICAqIEBwYXJhbSB7c3RyaW5nfSBzIElucHV0IHN0cmluZ1xuICAgKiBAcmV0dXJucyB7c3RyaW5nfSBSYXcgTUQ1IHN0cmluZ1xuICAgKi9cbiAgZnVuY3Rpb24gcmF3TUQ1KHMpIHtcbiAgICByZXR1cm4gcnN0ck1ENShzdHIycnN0clVURjgocykpXG4gIH1cbiAgLyoqXG4gICAqIEVuY29kZXMgaW5wdXQgc3RyaW5nIGFzIEhleCBlbmNvZGVkIHN0cmluZ1xuICAgKlxuICAgKiBAcGFyYW0ge3N0cmluZ30gcyBJbnB1dCBzdHJpbmdcbiAgICogQHJldHVybnMge3N0cmluZ30gSGV4IGVuY29kZWQgc3RyaW5nXG4gICAqL1xuICBmdW5jdGlvbiBoZXhNRDUocykge1xuICAgIHJldHVybiByc3RyMmhleChyYXdNRDUocykpXG4gIH1cbiAgLyoqXG4gICAqIENhbGN1bGF0ZXMgdGhlIHJhdyBITUFDLU1ENSBmb3IgdGhlIGdpdmVuIGtleSBhbmQgZGF0YVxuICAgKlxuICAgKiBAcGFyYW0ge3N0cmluZ30gayBITUFDIGtleVxuICAgKiBAcGFyYW0ge3N0cmluZ30gZCBJbnB1dCBzdHJpbmdcbiAgICogQHJldHVybnMge3N0cmluZ30gUmF3IE1ENSBzdHJpbmdcbiAgICovXG4gIGZ1bmN0aW9uIHJhd0hNQUNNRDUoaywgZCkge1xuICAgIHJldHVybiByc3RySE1BQ01ENShzdHIycnN0clVURjgoayksIHN0cjJyc3RyVVRGOChkKSlcbiAgfVxuICAvKipcbiAgICogQ2FsY3VsYXRlcyB0aGUgSGV4IGVuY29kZWQgSE1BQy1NRDUgZm9yIHRoZSBnaXZlbiBrZXkgYW5kIGRhdGFcbiAgICpcbiAgICogQHBhcmFtIHtzdHJpbmd9IGsgSE1BQyBrZXlcbiAgICogQHBhcmFtIHtzdHJpbmd9IGQgSW5wdXQgc3RyaW5nXG4gICAqIEByZXR1cm5zIHtzdHJpbmd9IFJhdyBNRDUgc3RyaW5nXG4gICAqL1xuICBmdW5jdGlvbiBoZXhITUFDTUQ1KGssIGQpIHtcbiAgICByZXR1cm4gcnN0cjJoZXgocmF3SE1BQ01ENShrLCBkKSlcbiAgfVxuXG4gIC8qKlxuICAgKiBDYWxjdWxhdGVzIE1ENSB2YWx1ZSBmb3IgYSBnaXZlbiBzdHJpbmcuXG4gICAqIElmIGEga2V5IGlzIHByb3ZpZGVkLCBjYWxjdWxhdGVzIHRoZSBITUFDLU1ENSB2YWx1ZS5cbiAgICogUmV0dXJucyBhIEhleCBlbmNvZGVkIHN0cmluZyB1bmxlc3MgdGhlIHJhdyBhcmd1bWVudCBpcyBnaXZlbi5cbiAgICpcbiAgICogQHBhcmFtIHtzdHJpbmd9IHN0cmluZyBJbnB1dCBzdHJpbmdcbiAgICogQHBhcmFtIHtzdHJpbmd9IFtrZXldIEhNQUMga2V5XG4gICAqIEBwYXJhbSB7Ym9vbGVhbn0gW3Jhd10gUmF3IG91dHB1dCBzd2l0Y2hcbiAgICogQHJldHVybnMge3N0cmluZ30gTUQ1IG91dHB1dFxuICAgKi9cbiAgZnVuY3Rpb24gbWQ1KHN0cmluZywga2V5LCByYXcpIHtcbiAgICBpZiAoIWtleSkge1xuICAgICAgaWYgKCFyYXcpIHtcbiAgICAgICAgcmV0dXJuIGhleE1ENShzdHJpbmcpXG4gICAgICB9XG4gICAgICByZXR1cm4gcmF3TUQ1KHN0cmluZylcbiAgICB9XG4gICAgaWYgKCFyYXcpIHtcbiAgICAgIHJldHVybiBoZXhITUFDTUQ1KGtleSwgc3RyaW5nKVxuICAgIH1cbiAgICByZXR1cm4gcmF3SE1BQ01ENShrZXksIHN0cmluZylcbiAgfVxuXG4gIGlmICh0eXBlb2YgZGVmaW5lID09PSAnZnVuY3Rpb24nICYmIGRlZmluZS5hbWQpIHtcbiAgICBkZWZpbmUoZnVuY3Rpb24gKCkge1xuICAgICAgcmV0dXJuIG1kNVxuICAgIH0pXG4gIH0gZWxzZSBpZiAodHlwZW9mIG1vZHVsZSA9PT0gJ29iamVjdCcgJiYgbW9kdWxlLmV4cG9ydHMpIHtcbiAgICBtb2R1bGUuZXhwb3J0cyA9IG1kNVxuICB9IGVsc2Uge1xuICAgICQubWQ1ID0gbWQ1XG4gIH1cbn0pKHRoaXMpXG4iLCIvKiFcbiBDb3B5cmlnaHQgMjAxOCBHb29nbGUgSW5jLiBBbGwgUmlnaHRzIFJlc2VydmVkLlxuIExpY2Vuc2VkIHVuZGVyIHRoZSBBcGFjaGUgTGljZW5zZSwgVmVyc2lvbiAyLjAgKHRoZSBcIkxpY2Vuc2VcIik7XG4geW91IG1heSBub3QgdXNlIHRoaXMgZmlsZSBleGNlcHQgaW4gY29tcGxpYW5jZSB3aXRoIHRoZSBMaWNlbnNlLlxuIFlvdSBtYXkgb2J0YWluIGEgY29weSBvZiB0aGUgTGljZW5zZSBhdFxuXG4gICAgIGh0dHA6Ly93d3cuYXBhY2hlLm9yZy9saWNlbnNlcy9MSUNFTlNFLTIuMFxuXG4gVW5sZXNzIHJlcXVpcmVkIGJ5IGFwcGxpY2FibGUgbGF3IG9yIGFncmVlZCB0byBpbiB3cml0aW5nLCBzb2Z0d2FyZVxuIGRpc3RyaWJ1dGVkIHVuZGVyIHRoZSBMaWNlbnNlIGlzIGRpc3RyaWJ1dGVkIG9uIGFuIFwiQVMgSVNcIiBCQVNJUyxcbiBXSVRIT1VUIFdBUlJBTlRJRVMgT1IgQ09ORElUSU9OUyBPRiBBTlkgS0lORCwgZWl0aGVyIGV4cHJlc3Mgb3IgaW1wbGllZC5cbiBTZWUgdGhlIExpY2Vuc2UgZm9yIHRoZSBzcGVjaWZpYyBsYW5ndWFnZSBnb3Zlcm5pbmcgcGVybWlzc2lvbnMgYW5kXG4gbGltaXRhdGlvbnMgdW5kZXIgdGhlIExpY2Vuc2UuXG4qL1xuLyohIGxpZmVjeWNsZS5tanMgdjAuMS4xICovXG5sZXQgZTt0cnl7bmV3IEV2ZW50VGFyZ2V0LGU9ITB9Y2F0Y2godCl7ZT0hMX1jbGFzcyB0e2NvbnN0cnVjdG9yKCl7dGhpcy5lPXt9fWFkZEV2ZW50TGlzdGVuZXIoZSx0LHM9ITEpe3RoaXMudChlKS5wdXNoKHQpfXJlbW92ZUV2ZW50TGlzdGVuZXIoZSx0LHM9ITEpe2NvbnN0IGk9dGhpcy50KGUpLGE9aS5pbmRleE9mKHQpO2E+LTEmJmkuc3BsaWNlKGEsMSl9ZGlzcGF0Y2hFdmVudChlKXtyZXR1cm4gZS50YXJnZXQ9dGhpcyxPYmplY3QuZnJlZXplKGUpLHRoaXMudChlLnR5cGUpLmZvckVhY2godD0+dChlKSksITB9dChlKXtyZXR1cm4gdGhpcy5lW2VdPXRoaXMuZVtlXXx8W119fXZhciBzPWU/RXZlbnRUYXJnZXQ6dDtjbGFzcyBpe2NvbnN0cnVjdG9yKGUpe3RoaXMudHlwZT1lfX12YXIgYT1lP0V2ZW50Omk7Y2xhc3MgbiBleHRlbmRzIGF7Y29uc3RydWN0b3IoZSx0KXtzdXBlcihlKSx0aGlzLm5ld1N0YXRlPXQubmV3U3RhdGUsdGhpcy5vbGRTdGF0ZT10Lm9sZFN0YXRlLHRoaXMub3JpZ2luYWxFdmVudD10Lm9yaWdpbmFsRXZlbnR9fWNvbnN0IHI9XCJhY3RpdmVcIixoPVwicGFzc2l2ZVwiLGM9XCJoaWRkZW5cIixvPVwiZnJvemVuXCIsZD1cInRlcm1pbmF0ZWRcIix1PVwib2JqZWN0XCI9PXR5cGVvZiBzYWZhcmkmJnNhZmFyaS5wdXNoTm90aWZpY2F0aW9uLHY9XCJvbnBhZ2VzaG93XCJpbiBzZWxmLGw9W1wiZm9jdXNcIixcImJsdXJcIixcInZpc2liaWxpdHljaGFuZ2VcIixcImZyZWV6ZVwiLFwicmVzdW1lXCIsXCJwYWdlc2hvd1wiLHY/XCJwYWdlaGlkZVwiOlwidW5sb2FkXCJdLGc9ZT0+KGUucHJldmVudERlZmF1bHQoKSxlLnJldHVyblZhbHVlPVwiQXJlIHlvdSBzdXJlP1wiKSxmPWU9PmUucmVkdWNlKChlLHQscyk9PihlW3RdPXMsZSkse30pLGI9W1tyLGgsYyxkXSxbcixoLGMsb10sW2MsaCxyXSxbbyxjXSxbbyxyXSxbbyxoXV0ubWFwKGYpLHA9KGUsdCk9Pntmb3IobGV0IHMsaT0wO3M9YltpXTsrK2kpe2NvbnN0IGk9c1tlXSxhPXNbdF07aWYoaT49MCYmYT49MCYmYT5pKXJldHVybiBPYmplY3Qua2V5cyhzKS5zbGljZShpLGErMSl9cmV0dXJuW119LEU9KCk9PmRvY3VtZW50LnZpc2liaWxpdHlTdGF0ZT09PWM/Yzpkb2N1bWVudC5oYXNGb2N1cygpP3I6aDtjbGFzcyBtIGV4dGVuZHMgc3tjb25zdHJ1Y3Rvcigpe3N1cGVyKCk7Y29uc3QgZT1FKCk7dGhpcy5zPWUsdGhpcy5pPVtdLHRoaXMuYT10aGlzLmEuYmluZCh0aGlzKSxsLmZvckVhY2goZT0+YWRkRXZlbnRMaXN0ZW5lcihlLHRoaXMuYSwhMCkpLHUmJmFkZEV2ZW50TGlzdGVuZXIoXCJiZWZvcmV1bmxvYWRcIixlPT57dGhpcy5uPXNldFRpbWVvdXQoKCk9PntlLmRlZmF1bHRQcmV2ZW50ZWR8fGUucmV0dXJuVmFsdWUubGVuZ3RoPjB8fHRoaXMucihlLGMpfSwwKX0pfWdldCBzdGF0ZSgpe3JldHVybiB0aGlzLnN9Z2V0IHBhZ2VXYXNEaXNjYXJkZWQoKXtyZXR1cm4gZG9jdW1lbnQud2FzRGlzY2FyZGVkfHwhMX1hZGRVbnNhdmVkQ2hhbmdlcyhlKXshdGhpcy5pLmluZGV4T2YoZSk+LTEmJigwPT09dGhpcy5pLmxlbmd0aCYmYWRkRXZlbnRMaXN0ZW5lcihcImJlZm9yZXVubG9hZFwiLGcpLHRoaXMuaS5wdXNoKGUpKX1yZW1vdmVVbnNhdmVkQ2hhbmdlcyhlKXtjb25zdCB0PXRoaXMuaS5pbmRleE9mKGUpO3Q+LTEmJih0aGlzLmkuc3BsaWNlKHQsMSksMD09PXRoaXMuaS5sZW5ndGgmJnJlbW92ZUV2ZW50TGlzdGVuZXIoXCJiZWZvcmV1bmxvYWRcIixnKSl9cihlLHQpe2lmKHQhPT10aGlzLnMpe2NvbnN0IHM9dGhpcy5zLGk9cChzLHQpO2ZvcihsZXQgdD0wO3Q8aS5sZW5ndGgtMTsrK3Qpe2NvbnN0IHM9aVt0XSxhPWlbdCsxXTt0aGlzLnM9YSx0aGlzLmRpc3BhdGNoRXZlbnQobmV3IG4oXCJzdGF0ZWNoYW5nZVwiLHtvbGRTdGF0ZTpzLG5ld1N0YXRlOmEsb3JpZ2luYWxFdmVudDplfSkpfX19YShlKXtzd2l0Y2godSYmY2xlYXJUaW1lb3V0KHRoaXMubiksZS50eXBlKXtjYXNlXCJwYWdlc2hvd1wiOmNhc2VcInJlc3VtZVwiOnRoaXMucihlLEUoKSk7YnJlYWs7Y2FzZVwiZm9jdXNcIjp0aGlzLnIoZSxyKTticmVhaztjYXNlXCJibHVyXCI6dGhpcy5zPT09ciYmdGhpcy5yKGUsRSgpKTticmVhaztjYXNlXCJwYWdlaGlkZVwiOmNhc2VcInVubG9hZFwiOnRoaXMucihlLGUucGVyc2lzdGVkP286ZCk7YnJlYWs7Y2FzZVwidmlzaWJpbGl0eWNoYW5nZVwiOnRoaXMucyE9PW8mJnRoaXMucyE9PWQmJnRoaXMucihlLEUoKSk7YnJlYWs7Y2FzZVwiZnJlZXplXCI6dGhpcy5yKGUsbyl9fX12YXIgdz1uZXcgbTtleHBvcnQgZGVmYXVsdCB3O1xuLy8jIHNvdXJjZU1hcHBpbmdVUkw9bGlmZWN5Y2xlLm1qcy5tYXBcbiIsbnVsbF0sIm5hbWVzIjpbInRoaXMiLCJVUkxQYXJhbXMiLCJkZXZpY2VfZXh0ZXJuYWxfaWQiLCJ1c2VyX2V4dGVybmFsX2lkIiwidXNlcl9pc19hdXRoZW50aWNhdGVkIiwidXNlcl9leHRlcm5hbF9pZF9obWFjIiwiT25lWWVhckluU2Vjb25kcyIsIlBhZ2VTdGF0ZXMiLCJhY3RpdmUiLCJwYXNzaXZlIiwiaGlkZGVuIiwiZnJvemVuIiwiUmltZGlhbiIsImNvbmZpZyIsIndvcmtzcGFjZV9pZCIsImhvc3QiLCJzZXNzaW9uX3RpbWVvdXQiLCJuYW1lc3BhY2UiLCJjcm9zc19kb21haW5zIiwiaWdub3JlZF9vcmlnaW5zIiwidmVyc2lvbiIsImxvZ19sZXZlbCIsIm1heF9yZXRyeSIsImZyb21fY20iLCJpc1JlYWR5IiwiZGlzcGF0Y2hDb25zZW50IiwiY3VycmVudFVzZXIiLCJ1bmRlZmluZWQiLCJjdXJyZW50RGV2aWNlIiwiY3VycmVudFNlc3Npb24iLCJjdXJyZW50UGFnZXZpZXciLCJjdXJyZW50UGFnZXZpZXdWaXNpYmxlU2luY2UiLCJjdXJyZW50UGFnZXZpZXdEdXJhdGlvbiIsImN1cnJlbnRDYXJ0IiwiaXRlbXNRdWV1ZSIsIml0ZW1zIiwiYWRkIiwia2luZCIsImRhdGEiLCJzZXNzaW9uQ29va2llIiwiZ2V0Q29va2llIiwiaW5jbHVkZXMiLCJub25faW50ZXJhY3RpdmUiLCJfc3RhcnROZXdTZXNzaW9uIiwiaXRlbSIsInVzZXIiLCJfX2Fzc2lnbiIsInNlc3Npb24iLCJwdXNoIiwiX2xvY2FsU3RvcmFnZSIsInNldCIsIkpTT04iLCJzdHJpbmdpZnkiLCJhZGRQYWdldmlld0R1cmF0aW9uIiwiaW5jcmVtZW50IiwiTWF0aCIsInJvdW5kIiwiRGF0ZSIsImdldFRpbWUiLCJsb2ciLCJwYWdldmlldyIsIm5vdyIsInRvSVNPU3RyaW5nIiwiZHVyYXRpb24iLCJ1cGRhdGVkX2F0Iiwic2V0U2Vzc2lvbkNvbnRleHQiLCJkZXZpY2UiLCJkaXNwYXRjaFF1ZXVlIiwiaXNEaXNwYXRjaGluZyIsIm9uUmVhZHlRdWV1ZSIsImxldmVsIiwiYXJncyIsIl9pIiwiYXJndW1lbnRzIiwibGVuZ3RoIiwiY29uc29sZSIsIndhcm4iLCJhcHBseSIsImluZm8iLCJkZWJ1ZyIsInRyYWNlIiwiZXJyb3IiLCJpbml0IiwiY2ZnIiwiZG9jdW1lbnQiLCJyZWFkeVN0YXRlIiwiX29uUmVhZHkiLCJvbnJlYWR5c3RhdGVjaGFuZ2UiLCJsb2dMZXZlbCIsInNldERpc3BhdGNoQ29uc2VudCIsImNvbnNlbnQiLCJnZXRDdXJyZW50VXNlciIsImNhbGxiYWNrIiwiX2V4ZWNXaGVuUmVhZHkiLCJvblJlYWR5IiwiZm4iLCJ0cmFja1BhZ2V2aWV3IiwicGFyc2UiLCJleHRlcm5hbF9pZCIsInV1aWR2NCIsImNyZWF0ZWRfYXQiLCJyZWZlcnJlciIsImdldFJlZmVycmVyIiwidGl0bGUiLCJwYWdlX2lkIiwid2luZG93IiwibG9jYXRpb24iLCJocmVmIiwicmVwbGFjZSIsInByb2R1Y3RfcHJpY2UiLCJpc1BhZ2VWaXNpYmxlIiwibGFzdF9pbnRlcmFjdGlvbl9hdCIsInNldFVzZXJDb250ZXh0IiwicGFnZXZpZXdzX2NvdW50IiwiaW50ZXJhY3Rpb25zX2NvdW50IiwidHJhY2tDdXN0b21FdmVudCIsImxhYmVsIiwiY3VzdG9tRXZlbnQiLCJzdHJpbmdfdmFsdWUiLCJ0cmFja0NhcnQiLCJjYXJ0Iiwic2Vzc2lvbl9leHRlcm5hbF9pZCIsImZvckVhY2giLCJjYXJ0X2V4dGVybmFsX2lkIiwicHJpY2UiLCJoYXNoIiwiX2NhcnRIYXNoIiwidHJhY2tPcmRlciIsIm9yZGVyIiwib3JkZXJfZXh0ZXJuYWxfaWQiLCJzZXREZXZpY2VDb250ZXh0IiwibmV3RGV2aWNlIiwic2V0Q29va2llIiwibmV3U2Vzc2lvbiIsInVzZXJfY2VudHJpY19jb25zZW50IiwiY29uc2VudF9hbGwiLCJpc19hdXRoZW50aWNhdGVkIiwicmVtb3ZlIiwiX2NyZWF0ZURldmljZSIsImRlbGV0ZUNvb2tpZSIsIl9oYW5kbGVTZXNzaW9uIiwiT2JqZWN0Iiwia2V5cyIsImtleSIsImZyb21fdXNlcl9leHRlcm5hbF9pZCIsInRvX3VzZXJfZXh0ZXJuYWxfaWQiLCJ0b191c2VyX2lzX2F1dGhlbnRpY2F0ZWQiLCJ0b191c2VyX2NyZWF0ZWRfYXQiLCJuZXdVc2VyIiwic2F2ZVVzZXJQcm9maWxlIiwibGlmZWN5Y2xlIiwic3RhdGUiLCJkaXNwYXRjaCIsInVzZUJlYWNvbiIsInNldFRpbWVvdXQiLCJkZXZpY2VDdHgiLCJ1c2VyX2FnZW50IiwibmF2aWdhdG9yIiwidXNlckFnZW50IiwibGFuZ3VhZ2UiLCJhZF9ibG9ja2VyIiwiaGFzQWRCbG9ja2VyIiwicmVzb2x1dGlvbiIsInNjcmVlbiIsIndpZHRoIiwiaGVpZ2h0IiwiYmF0Y2hlcyIsIml0ZW1zQmF0Y2giLCJzaGlmdCIsImJhdGNoIiwiaWQiLCJjb250ZXh0IiwiX2luaXREaXNwYXRjaExvb3AiLCJzb3J0IiwiYSIsImIiLCJjdXJyZW50QmF0Y2giLCJfcG9zdFBheWxvYWQiLCJkYXRhSW1wb3J0IiwicmV0cnlDb3VudCIsIl9wb3N0Iiwic3VjY2VzcyIsImpzb25FcnJvciIsImNvZGUiLCJlcnIiLCJkZWxheSIsInRvV2FpdCIsImkiLCJyZW1haW5pbmdCYXRjaGVzXzEiLCJkaSIsImRhdGFfc2VudF9hdCIsInNlbmRCZWFjb24iLCJxdWV1ZWQiLCJCbG9iIiwidHlwZSIsInhociIsIlhNTEh0dHBSZXF1ZXN0Iiwib25sb2FkIiwiYm9keSIsInJlc3BvbnNlIiwic3RhdHVzIiwib25lcnJvciIsIm9udGltZW91dCIsIm9wZW4iLCJzZXRSZXF1ZXN0SGVhZGVyIiwid2l0aENyZWRlbnRpYWxzIiwic2VuZCIsIl9oYW5kbGVVc2VyIiwiZm9yYmlkZGVuUGF0dGVybnMiLCJ1c2VyQ29va2llIiwicHJldmlvdXNVc2VyIiwic29tZSIsInBhdHRlcm4iLCJpbmRleE9mIiwiaG1hYyIsInVzZXJJZCIsImdldFF1ZXJ5UGFyYW0iLCJVUkwiLCJldmVyeSIsImlzQXV0aGVudGljYXRlZCIsIl9jcmVhdGVVc2VyIiwidXNlckV4dGVybmFsSWQiLCJjcmVhdGVkQXQiLCJfZW5yaWNoVXNlckNvbnRleHQiLCJvdGhlclVzZXJJZHMiLCJ2YWx1ZSIsIngiLCJfaGFuZGxlRGV2aWNlIiwiZGV2aWNlSWQiLCJkZXZpY2VDb29raWUiLCJsZWdhY3lDbGllbnRJRCIsImxlZ2FjeUNsaWVudFRpbWVzdGFtcCIsInBhcnNlSW50IiwiX2FkZEV2ZW50TGlzdGVuZXIiLCJlbGVtZW50IiwiZXZlbnRUeXBlIiwiZXZlbnRIYW5kbGVyIiwidXNlQ2FwdHVyZSIsImFkZEV2ZW50TGlzdGVuZXIiLCJhdHRhY2hFdmVudCIsImlzQnJvd3NlckxlZ2l0IiwibG9jYWxTdG9yYWdlIiwiZ2V0SXRlbSIsInNldEludGVydmFsIiwiY29va2llU2Vzc2lvbiIsImxpbmtzIiwiZWx0IiwiZCIsIl9kZWNvcmF0ZVVSTCIsImV2ZW50Iiwib2xkU3RhdGUiLCJuZXdTdGF0ZSIsIl9vblBhZ2VQYXNzaXZlIiwiX29uUGFnZUFjdGl2ZSIsInV0bV9zb3VyY2UiLCJnZXRIYXNoUGFyYW0iLCJ1dG1fbWVkaXVtIiwidXRtX2NhbXBhaWduIiwidXRtX2NvbnRlbnQiLCJ1dG1fdGVybSIsInV0bV9pZCIsInV0bV9pZF9mcm9tIiwicmVmZXJyZXJVUkwiLCJjcmVhdGVFbGVtZW50IiwiZnJvbUFub3RoZXJEb21haW4iLCJob3N0bmFtZSIsImlzQ3Jvc3NEb21haW5fMSIsImRvbSIsInNlYXJjaCIsImlkcyIsInBhcmFtIiwiaWdub3JlZE9yaWdpbiIsImZpbmQiLCJvcmlnaW4iLCJleGlzdGluZ1Nlc3Npb24iLCJpc0VxdWFsIiwiZ2V0VGltZXpvbmUiLCJEYXRlVGltZUZvcm1hdCIsIl9hIiwiSW50bCIsInRpbWV6b25lIiwicmVzb2x2ZWRPcHRpb25zIiwidGltZVpvbmUiLCJ1cmwiLCJuYW1lIiwidXJsT2JqZWN0IiwicGFyYW1zIiwiVVJMU2VhcmNoUGFyYW1zIiwiZ2V0IiwiZSIsIm1hdGNoZXMiLCJtYXRjaCIsIlJlZ0V4cCIsInVwZGF0ZVVSTFBhcmFtIiwidG9TdHJpbmciLCJhZHMiLCJpbm5lckhUTUwiLCJjbGFzc05hbWUiLCJibG9ja2VkIiwiYXBwZW5kQ2hpbGQiLCJnZXRFbGVtZW50c0J5Q2xhc3NOYW1lIiwib2Zmc2V0SGVpZ2h0IiwicmVtb3ZlQ2hpbGQiLCJfZSIsInVhIiwidG9Mb3dlckNhc2UiLCJzcGxpdCIsInRlc3QiLCJ3ZWJkcml2ZXIiLCJjIiwiciIsInJhbmRvbSIsInYiLCJtZDUiLCJzdHIiLCJ0b3AiLCJwYXJlbnQiLCJsYW5kaW5nX3BhZ2UiLCJkZWNvZGVVUklDb21wb25lbnQiLCJjb29raWUiLCJlbmNvZGVVUklDb21wb25lbnQiLCJzZWNvbmRzIiwiZG9tYWluIiwieGRvbWFpbiIsInNldFRpbWUiLCJleHBpcmVzIiwidG9VVENTdHJpbmciLCJjb29raWVfdmFsdWUiLCJzZXRJdGVtIiwicmVtb3ZlSXRlbSIsInRhcmdldCIsImNhcnRIYXNoIiwicHVibGljX3VybCIsInByb2R1Y3RfZXh0ZXJuYWxfaWQiLCJ2YXJpYW50X2V4dGVybmFsX2lkIiwicXVhbnRpdHkiLCJfd2lwZUFsbCIsImNvbmZpcm0iXSwibWFwcGluZ3MiOiI7O0VBQWUsU0FBUyxPQUFPLENBQUMsR0FBRyxFQUFFO0VBQ3JDLEVBQUUseUJBQXlCLENBQUM7QUFDNUI7RUFDQSxFQUFFLE9BQU8sT0FBTyxHQUFHLFVBQVUsSUFBSSxPQUFPLE1BQU0sSUFBSSxRQUFRLElBQUksT0FBTyxNQUFNLENBQUMsUUFBUSxHQUFHLFVBQVUsR0FBRyxFQUFFO0VBQ3RHLElBQUksT0FBTyxPQUFPLEdBQUcsQ0FBQztFQUN0QixHQUFHLEdBQUcsVUFBVSxHQUFHLEVBQUU7RUFDckIsSUFBSSxPQUFPLEdBQUcsSUFBSSxVQUFVLElBQUksT0FBTyxNQUFNLElBQUksR0FBRyxDQUFDLFdBQVcsS0FBSyxNQUFNLElBQUksR0FBRyxLQUFLLE1BQU0sQ0FBQyxTQUFTLEdBQUcsUUFBUSxHQUFHLE9BQU8sR0FBRyxDQUFDO0VBQ2hJLEdBQUcsRUFBRSxPQUFPLENBQUMsR0FBRyxDQUFDLENBQUM7RUFDbEI7O0VDUkE7RUFDQTtBQUNBO0VBQ0E7RUFDQTtBQUNBO0VBQ0E7RUFDQTtFQUNBO0VBQ0E7RUFDQTtFQUNBO0VBQ0E7RUFDQTtBQWlCQTtFQUNPLElBQUksUUFBUSxHQUFHLFdBQVc7RUFDakMsSUFBSSxRQUFRLEdBQUcsTUFBTSxDQUFDLE1BQU0sSUFBSSxTQUFTLFFBQVEsQ0FBQyxDQUFDLEVBQUU7RUFDckQsUUFBUSxLQUFLLElBQUksQ0FBQyxFQUFFLENBQUMsR0FBRyxDQUFDLEVBQUUsQ0FBQyxHQUFHLFNBQVMsQ0FBQyxNQUFNLEVBQUUsQ0FBQyxHQUFHLENBQUMsRUFBRSxDQUFDLEVBQUUsRUFBRTtFQUM3RCxZQUFZLENBQUMsR0FBRyxTQUFTLENBQUMsQ0FBQyxDQUFDLENBQUM7RUFDN0IsWUFBWSxLQUFLLElBQUksQ0FBQyxJQUFJLENBQUMsRUFBRSxJQUFJLE1BQU0sQ0FBQyxTQUFTLENBQUMsY0FBYyxDQUFDLElBQUksQ0FBQyxDQUFDLEVBQUUsQ0FBQyxDQUFDLEVBQUUsQ0FBQyxDQUFDLENBQUMsQ0FBQyxHQUFHLENBQUMsQ0FBQyxDQUFDLENBQUMsQ0FBQztFQUN6RixTQUFTO0VBQ1QsUUFBUSxPQUFPLENBQUMsQ0FBQztFQUNqQixNQUFLO0VBQ0wsSUFBSSxPQUFPLFFBQVEsQ0FBQyxLQUFLLENBQUMsSUFBSSxFQUFFLFNBQVMsQ0FBQyxDQUFDO0VBQzNDOzs7Ozs7Ozs7Ozs7Ozs7Ozs7Ozs7Ozs7OztFQ2pCQyxDQUFDLFVBQVUsQ0FBQyxFQUFFO0FBRWY7RUFDQTtFQUNBO0VBQ0E7RUFDQTtFQUNBO0VBQ0E7RUFDQTtFQUNBO0VBQ0EsR0FBRSxTQUFTLE9BQU8sQ0FBQyxDQUFDLEVBQUUsQ0FBQyxFQUFFO09BQ3JCLElBQUksR0FBRyxHQUFHLENBQUMsQ0FBQyxHQUFHLE1BQU0sS0FBSyxDQUFDLEdBQUcsTUFBTSxFQUFDO0VBQ3pDLEtBQUksSUFBSSxHQUFHLEdBQUcsQ0FBQyxDQUFDLElBQUksRUFBRSxLQUFLLENBQUMsSUFBSSxFQUFFLENBQUMsSUFBSSxHQUFHLElBQUksRUFBRSxFQUFDO09BQzdDLE9BQU8sQ0FBQyxHQUFHLElBQUksRUFBRSxLQUFLLEdBQUcsR0FBRyxNQUFNLENBQUM7TUFDcEM7QUFDSDtFQUNBO0VBQ0E7RUFDQTtFQUNBO0VBQ0E7RUFDQTtFQUNBO0VBQ0EsR0FBRSxTQUFTLGFBQWEsQ0FBQyxHQUFHLEVBQUUsR0FBRyxFQUFFO0VBQ25DLEtBQUksT0FBTyxDQUFDLEdBQUcsSUFBSSxHQUFHLEtBQUssR0FBRyxNQUFNLEVBQUUsR0FBRyxHQUFHLENBQUMsQ0FBQztNQUMzQztBQUNIO0VBQ0E7RUFDQTtFQUNBO0VBQ0E7RUFDQTtFQUNBO0VBQ0E7RUFDQTtFQUNBO0VBQ0E7RUFDQTtFQUNBLEdBQUUsU0FBUyxNQUFNLENBQUMsQ0FBQyxFQUFFLENBQUMsRUFBRSxDQUFDLEVBQUUsQ0FBQyxFQUFFLENBQUMsRUFBRSxDQUFDLEVBQUU7T0FDaEMsT0FBTyxPQUFPLENBQUMsYUFBYSxDQUFDLE9BQU8sQ0FBQyxPQUFPLENBQUMsQ0FBQyxFQUFFLENBQUMsQ0FBQyxFQUFFLE9BQU8sQ0FBQyxDQUFDLEVBQUUsQ0FBQyxDQUFDLENBQUMsRUFBRSxDQUFDLENBQUMsRUFBRSxDQUFDLENBQUM7TUFDM0U7RUFDSDtFQUNBO0VBQ0E7RUFDQTtFQUNBO0VBQ0E7RUFDQTtFQUNBO0VBQ0E7RUFDQTtFQUNBO0VBQ0E7RUFDQSxHQUFFLFNBQVMsS0FBSyxDQUFDLENBQUMsRUFBRSxDQUFDLEVBQUUsQ0FBQyxFQUFFLENBQUMsRUFBRSxDQUFDLEVBQUUsQ0FBQyxFQUFFLENBQUMsRUFBRTtPQUNsQyxPQUFPLE1BQU0sQ0FBQyxDQUFDLENBQUMsR0FBRyxDQUFDLEtBQUssQ0FBQyxDQUFDLEdBQUcsQ0FBQyxDQUFDLEVBQUUsQ0FBQyxFQUFFLENBQUMsRUFBRSxDQUFDLEVBQUUsQ0FBQyxFQUFFLENBQUMsQ0FBQztNQUNqRDtFQUNIO0VBQ0E7RUFDQTtFQUNBO0VBQ0E7RUFDQTtFQUNBO0VBQ0E7RUFDQTtFQUNBO0VBQ0E7RUFDQTtFQUNBLEdBQUUsU0FBUyxLQUFLLENBQUMsQ0FBQyxFQUFFLENBQUMsRUFBRSxDQUFDLEVBQUUsQ0FBQyxFQUFFLENBQUMsRUFBRSxDQUFDLEVBQUUsQ0FBQyxFQUFFO09BQ2xDLE9BQU8sTUFBTSxDQUFDLENBQUMsQ0FBQyxHQUFHLENBQUMsS0FBSyxDQUFDLEdBQUcsQ0FBQyxDQUFDLENBQUMsRUFBRSxDQUFDLEVBQUUsQ0FBQyxFQUFFLENBQUMsRUFBRSxDQUFDLEVBQUUsQ0FBQyxDQUFDO01BQ2pEO0VBQ0g7RUFDQTtFQUNBO0VBQ0E7RUFDQTtFQUNBO0VBQ0E7RUFDQTtFQUNBO0VBQ0E7RUFDQTtFQUNBO0VBQ0EsR0FBRSxTQUFTLEtBQUssQ0FBQyxDQUFDLEVBQUUsQ0FBQyxFQUFFLENBQUMsRUFBRSxDQUFDLEVBQUUsQ0FBQyxFQUFFLENBQUMsRUFBRSxDQUFDLEVBQUU7RUFDdEMsS0FBSSxPQUFPLE1BQU0sQ0FBQyxDQUFDLEdBQUcsQ0FBQyxHQUFHLENBQUMsRUFBRSxDQUFDLEVBQUUsQ0FBQyxFQUFFLENBQUMsRUFBRSxDQUFDLEVBQUUsQ0FBQyxDQUFDO01BQ3hDO0VBQ0g7RUFDQTtFQUNBO0VBQ0E7RUFDQTtFQUNBO0VBQ0E7RUFDQTtFQUNBO0VBQ0E7RUFDQTtFQUNBO0VBQ0EsR0FBRSxTQUFTLEtBQUssQ0FBQyxDQUFDLEVBQUUsQ0FBQyxFQUFFLENBQUMsRUFBRSxDQUFDLEVBQUUsQ0FBQyxFQUFFLENBQUMsRUFBRSxDQUFDLEVBQUU7T0FDbEMsT0FBTyxNQUFNLENBQUMsQ0FBQyxJQUFJLENBQUMsR0FBRyxDQUFDLENBQUMsQ0FBQyxFQUFFLENBQUMsRUFBRSxDQUFDLEVBQUUsQ0FBQyxFQUFFLENBQUMsRUFBRSxDQUFDLENBQUM7TUFDM0M7QUFDSDtFQUNBO0VBQ0E7RUFDQTtFQUNBO0VBQ0E7RUFDQTtFQUNBO0VBQ0EsR0FBRSxTQUFTLE9BQU8sQ0FBQyxDQUFDLEVBQUUsR0FBRyxFQUFFO0VBQzNCO09BQ0ksQ0FBQyxDQUFDLEdBQUcsSUFBSSxDQUFDLENBQUMsSUFBSSxJQUFJLElBQUksR0FBRyxHQUFHLEdBQUU7RUFDbkMsS0FBSSxDQUFDLENBQUMsQ0FBQyxDQUFDLENBQUMsR0FBRyxHQUFHLEVBQUUsTUFBTSxDQUFDLEtBQUssQ0FBQyxJQUFJLEVBQUUsQ0FBQyxHQUFHLElBQUc7QUFDM0M7RUFDQSxLQUFJLElBQUksRUFBQztFQUNULEtBQUksSUFBSSxLQUFJO0VBQ1osS0FBSSxJQUFJLEtBQUk7RUFDWixLQUFJLElBQUksS0FBSTtFQUNaLEtBQUksSUFBSSxLQUFJO09BQ1IsSUFBSSxDQUFDLEdBQUcsV0FBVTtFQUN0QixLQUFJLElBQUksQ0FBQyxHQUFHLENBQUMsVUFBUztFQUN0QixLQUFJLElBQUksQ0FBQyxHQUFHLENBQUMsV0FBVTtPQUNuQixJQUFJLENBQUMsR0FBRyxVQUFTO0FBQ3JCO0VBQ0EsS0FBSSxLQUFLLENBQUMsR0FBRyxDQUFDLEVBQUUsQ0FBQyxHQUFHLENBQUMsQ0FBQyxNQUFNLEVBQUUsQ0FBQyxJQUFJLEVBQUUsRUFBRTtTQUNqQyxJQUFJLEdBQUcsRUFBQztTQUNSLElBQUksR0FBRyxFQUFDO1NBQ1IsSUFBSSxHQUFHLEVBQUM7U0FDUixJQUFJLEdBQUcsRUFBQztBQUNkO1NBQ00sQ0FBQyxHQUFHLEtBQUssQ0FBQyxDQUFDLEVBQUUsQ0FBQyxFQUFFLENBQUMsRUFBRSxDQUFDLEVBQUUsQ0FBQyxDQUFDLENBQUMsQ0FBQyxFQUFFLENBQUMsRUFBRSxDQUFDLFNBQVMsRUFBQztTQUMxQyxDQUFDLEdBQUcsS0FBSyxDQUFDLENBQUMsRUFBRSxDQUFDLEVBQUUsQ0FBQyxFQUFFLENBQUMsRUFBRSxDQUFDLENBQUMsQ0FBQyxHQUFHLENBQUMsQ0FBQyxFQUFFLEVBQUUsRUFBRSxDQUFDLFNBQVMsRUFBQztTQUMvQyxDQUFDLEdBQUcsS0FBSyxDQUFDLENBQUMsRUFBRSxDQUFDLEVBQUUsQ0FBQyxFQUFFLENBQUMsRUFBRSxDQUFDLENBQUMsQ0FBQyxHQUFHLENBQUMsQ0FBQyxFQUFFLEVBQUUsRUFBRSxTQUFTLEVBQUM7U0FDOUMsQ0FBQyxHQUFHLEtBQUssQ0FBQyxDQUFDLEVBQUUsQ0FBQyxFQUFFLENBQUMsRUFBRSxDQUFDLEVBQUUsQ0FBQyxDQUFDLENBQUMsR0FBRyxDQUFDLENBQUMsRUFBRSxFQUFFLEVBQUUsQ0FBQyxVQUFVLEVBQUM7U0FDaEQsQ0FBQyxHQUFHLEtBQUssQ0FBQyxDQUFDLEVBQUUsQ0FBQyxFQUFFLENBQUMsRUFBRSxDQUFDLEVBQUUsQ0FBQyxDQUFDLENBQUMsR0FBRyxDQUFDLENBQUMsRUFBRSxDQUFDLEVBQUUsQ0FBQyxTQUFTLEVBQUM7U0FDOUMsQ0FBQyxHQUFHLEtBQUssQ0FBQyxDQUFDLEVBQUUsQ0FBQyxFQUFFLENBQUMsRUFBRSxDQUFDLEVBQUUsQ0FBQyxDQUFDLENBQUMsR0FBRyxDQUFDLENBQUMsRUFBRSxFQUFFLEVBQUUsVUFBVSxFQUFDO1NBQy9DLENBQUMsR0FBRyxLQUFLLENBQUMsQ0FBQyxFQUFFLENBQUMsRUFBRSxDQUFDLEVBQUUsQ0FBQyxFQUFFLENBQUMsQ0FBQyxDQUFDLEdBQUcsQ0FBQyxDQUFDLEVBQUUsRUFBRSxFQUFFLENBQUMsVUFBVSxFQUFDO1NBQ2hELENBQUMsR0FBRyxLQUFLLENBQUMsQ0FBQyxFQUFFLENBQUMsRUFBRSxDQUFDLEVBQUUsQ0FBQyxFQUFFLENBQUMsQ0FBQyxDQUFDLEdBQUcsQ0FBQyxDQUFDLEVBQUUsRUFBRSxFQUFFLENBQUMsUUFBUSxFQUFDO1NBQzlDLENBQUMsR0FBRyxLQUFLLENBQUMsQ0FBQyxFQUFFLENBQUMsRUFBRSxDQUFDLEVBQUUsQ0FBQyxFQUFFLENBQUMsQ0FBQyxDQUFDLEdBQUcsQ0FBQyxDQUFDLEVBQUUsQ0FBQyxFQUFFLFVBQVUsRUFBQztTQUM5QyxDQUFDLEdBQUcsS0FBSyxDQUFDLENBQUMsRUFBRSxDQUFDLEVBQUUsQ0FBQyxFQUFFLENBQUMsRUFBRSxDQUFDLENBQUMsQ0FBQyxHQUFHLENBQUMsQ0FBQyxFQUFFLEVBQUUsRUFBRSxDQUFDLFVBQVUsRUFBQztTQUNoRCxDQUFDLEdBQUcsS0FBSyxDQUFDLENBQUMsRUFBRSxDQUFDLEVBQUUsQ0FBQyxFQUFFLENBQUMsRUFBRSxDQUFDLENBQUMsQ0FBQyxHQUFHLEVBQUUsQ0FBQyxFQUFFLEVBQUUsRUFBRSxDQUFDLEtBQUssRUFBQztTQUM1QyxDQUFDLEdBQUcsS0FBSyxDQUFDLENBQUMsRUFBRSxDQUFDLEVBQUUsQ0FBQyxFQUFFLENBQUMsRUFBRSxDQUFDLENBQUMsQ0FBQyxHQUFHLEVBQUUsQ0FBQyxFQUFFLEVBQUUsRUFBRSxDQUFDLFVBQVUsRUFBQztTQUNqRCxDQUFDLEdBQUcsS0FBSyxDQUFDLENBQUMsRUFBRSxDQUFDLEVBQUUsQ0FBQyxFQUFFLENBQUMsRUFBRSxDQUFDLENBQUMsQ0FBQyxHQUFHLEVBQUUsQ0FBQyxFQUFFLENBQUMsRUFBRSxVQUFVLEVBQUM7U0FDL0MsQ0FBQyxHQUFHLEtBQUssQ0FBQyxDQUFDLEVBQUUsQ0FBQyxFQUFFLENBQUMsRUFBRSxDQUFDLEVBQUUsQ0FBQyxDQUFDLENBQUMsR0FBRyxFQUFFLENBQUMsRUFBRSxFQUFFLEVBQUUsQ0FBQyxRQUFRLEVBQUM7U0FDL0MsQ0FBQyxHQUFHLEtBQUssQ0FBQyxDQUFDLEVBQUUsQ0FBQyxFQUFFLENBQUMsRUFBRSxDQUFDLEVBQUUsQ0FBQyxDQUFDLENBQUMsR0FBRyxFQUFFLENBQUMsRUFBRSxFQUFFLEVBQUUsQ0FBQyxVQUFVLEVBQUM7U0FDakQsQ0FBQyxHQUFHLEtBQUssQ0FBQyxDQUFDLEVBQUUsQ0FBQyxFQUFFLENBQUMsRUFBRSxDQUFDLEVBQUUsQ0FBQyxDQUFDLENBQUMsR0FBRyxFQUFFLENBQUMsRUFBRSxFQUFFLEVBQUUsVUFBVSxFQUFDO0FBQ3REO1NBQ00sQ0FBQyxHQUFHLEtBQUssQ0FBQyxDQUFDLEVBQUUsQ0FBQyxFQUFFLENBQUMsRUFBRSxDQUFDLEVBQUUsQ0FBQyxDQUFDLENBQUMsR0FBRyxDQUFDLENBQUMsRUFBRSxDQUFDLEVBQUUsQ0FBQyxTQUFTLEVBQUM7U0FDOUMsQ0FBQyxHQUFHLEtBQUssQ0FBQyxDQUFDLEVBQUUsQ0FBQyxFQUFFLENBQUMsRUFBRSxDQUFDLEVBQUUsQ0FBQyxDQUFDLENBQUMsR0FBRyxDQUFDLENBQUMsRUFBRSxDQUFDLEVBQUUsQ0FBQyxVQUFVLEVBQUM7U0FDL0MsQ0FBQyxHQUFHLEtBQUssQ0FBQyxDQUFDLEVBQUUsQ0FBQyxFQUFFLENBQUMsRUFBRSxDQUFDLEVBQUUsQ0FBQyxDQUFDLENBQUMsR0FBRyxFQUFFLENBQUMsRUFBRSxFQUFFLEVBQUUsU0FBUyxFQUFDO1NBQy9DLENBQUMsR0FBRyxLQUFLLENBQUMsQ0FBQyxFQUFFLENBQUMsRUFBRSxDQUFDLEVBQUUsQ0FBQyxFQUFFLENBQUMsQ0FBQyxDQUFDLENBQUMsRUFBRSxFQUFFLEVBQUUsQ0FBQyxTQUFTLEVBQUM7U0FDM0MsQ0FBQyxHQUFHLEtBQUssQ0FBQyxDQUFDLEVBQUUsQ0FBQyxFQUFFLENBQUMsRUFBRSxDQUFDLEVBQUUsQ0FBQyxDQUFDLENBQUMsR0FBRyxDQUFDLENBQUMsRUFBRSxDQUFDLEVBQUUsQ0FBQyxTQUFTLEVBQUM7U0FDOUMsQ0FBQyxHQUFHLEtBQUssQ0FBQyxDQUFDLEVBQUUsQ0FBQyxFQUFFLENBQUMsRUFBRSxDQUFDLEVBQUUsQ0FBQyxDQUFDLENBQUMsR0FBRyxFQUFFLENBQUMsRUFBRSxDQUFDLEVBQUUsUUFBUSxFQUFDO1NBQzdDLENBQUMsR0FBRyxLQUFLLENBQUMsQ0FBQyxFQUFFLENBQUMsRUFBRSxDQUFDLEVBQUUsQ0FBQyxFQUFFLENBQUMsQ0FBQyxDQUFDLEdBQUcsRUFBRSxDQUFDLEVBQUUsRUFBRSxFQUFFLENBQUMsU0FBUyxFQUFDO1NBQ2hELENBQUMsR0FBRyxLQUFLLENBQUMsQ0FBQyxFQUFFLENBQUMsRUFBRSxDQUFDLEVBQUUsQ0FBQyxFQUFFLENBQUMsQ0FBQyxDQUFDLEdBQUcsQ0FBQyxDQUFDLEVBQUUsRUFBRSxFQUFFLENBQUMsU0FBUyxFQUFDO1NBQy9DLENBQUMsR0FBRyxLQUFLLENBQUMsQ0FBQyxFQUFFLENBQUMsRUFBRSxDQUFDLEVBQUUsQ0FBQyxFQUFFLENBQUMsQ0FBQyxDQUFDLEdBQUcsQ0FBQyxDQUFDLEVBQUUsQ0FBQyxFQUFFLFNBQVMsRUFBQztTQUM3QyxDQUFDLEdBQUcsS0FBSyxDQUFDLENBQUMsRUFBRSxDQUFDLEVBQUUsQ0FBQyxFQUFFLENBQUMsRUFBRSxDQUFDLENBQUMsQ0FBQyxHQUFHLEVBQUUsQ0FBQyxFQUFFLENBQUMsRUFBRSxDQUFDLFVBQVUsRUFBQztTQUNoRCxDQUFDLEdBQUcsS0FBSyxDQUFDLENBQUMsRUFBRSxDQUFDLEVBQUUsQ0FBQyxFQUFFLENBQUMsRUFBRSxDQUFDLENBQUMsQ0FBQyxHQUFHLENBQUMsQ0FBQyxFQUFFLEVBQUUsRUFBRSxDQUFDLFNBQVMsRUFBQztTQUMvQyxDQUFDLEdBQUcsS0FBSyxDQUFDLENBQUMsRUFBRSxDQUFDLEVBQUUsQ0FBQyxFQUFFLENBQUMsRUFBRSxDQUFDLENBQUMsQ0FBQyxHQUFHLENBQUMsQ0FBQyxFQUFFLEVBQUUsRUFBRSxVQUFVLEVBQUM7U0FDL0MsQ0FBQyxHQUFHLEtBQUssQ0FBQyxDQUFDLEVBQUUsQ0FBQyxFQUFFLENBQUMsRUFBRSxDQUFDLEVBQUUsQ0FBQyxDQUFDLENBQUMsR0FBRyxFQUFFLENBQUMsRUFBRSxDQUFDLEVBQUUsQ0FBQyxVQUFVLEVBQUM7U0FDaEQsQ0FBQyxHQUFHLEtBQUssQ0FBQyxDQUFDLEVBQUUsQ0FBQyxFQUFFLENBQUMsRUFBRSxDQUFDLEVBQUUsQ0FBQyxDQUFDLENBQUMsR0FBRyxDQUFDLENBQUMsRUFBRSxDQUFDLEVBQUUsQ0FBQyxRQUFRLEVBQUM7U0FDN0MsQ0FBQyxHQUFHLEtBQUssQ0FBQyxDQUFDLEVBQUUsQ0FBQyxFQUFFLENBQUMsRUFBRSxDQUFDLEVBQUUsQ0FBQyxDQUFDLENBQUMsR0FBRyxDQUFDLENBQUMsRUFBRSxFQUFFLEVBQUUsVUFBVSxFQUFDO1NBQy9DLENBQUMsR0FBRyxLQUFLLENBQUMsQ0FBQyxFQUFFLENBQUMsRUFBRSxDQUFDLEVBQUUsQ0FBQyxFQUFFLENBQUMsQ0FBQyxDQUFDLEdBQUcsRUFBRSxDQUFDLEVBQUUsRUFBRSxFQUFFLENBQUMsVUFBVSxFQUFDO0FBQ3ZEO1NBQ00sQ0FBQyxHQUFHLEtBQUssQ0FBQyxDQUFDLEVBQUUsQ0FBQyxFQUFFLENBQUMsRUFBRSxDQUFDLEVBQUUsQ0FBQyxDQUFDLENBQUMsR0FBRyxDQUFDLENBQUMsRUFBRSxDQUFDLEVBQUUsQ0FBQyxNQUFNLEVBQUM7U0FDM0MsQ0FBQyxHQUFHLEtBQUssQ0FBQyxDQUFDLEVBQUUsQ0FBQyxFQUFFLENBQUMsRUFBRSxDQUFDLEVBQUUsQ0FBQyxDQUFDLENBQUMsR0FBRyxDQUFDLENBQUMsRUFBRSxFQUFFLEVBQUUsQ0FBQyxVQUFVLEVBQUM7U0FDaEQsQ0FBQyxHQUFHLEtBQUssQ0FBQyxDQUFDLEVBQUUsQ0FBQyxFQUFFLENBQUMsRUFBRSxDQUFDLEVBQUUsQ0FBQyxDQUFDLENBQUMsR0FBRyxFQUFFLENBQUMsRUFBRSxFQUFFLEVBQUUsVUFBVSxFQUFDO1NBQ2hELENBQUMsR0FBRyxLQUFLLENBQUMsQ0FBQyxFQUFFLENBQUMsRUFBRSxDQUFDLEVBQUUsQ0FBQyxFQUFFLENBQUMsQ0FBQyxDQUFDLEdBQUcsRUFBRSxDQUFDLEVBQUUsRUFBRSxFQUFFLENBQUMsUUFBUSxFQUFDO1NBQy9DLENBQUMsR0FBRyxLQUFLLENBQUMsQ0FBQyxFQUFFLENBQUMsRUFBRSxDQUFDLEVBQUUsQ0FBQyxFQUFFLENBQUMsQ0FBQyxDQUFDLEdBQUcsQ0FBQyxDQUFDLEVBQUUsQ0FBQyxFQUFFLENBQUMsVUFBVSxFQUFDO1NBQy9DLENBQUMsR0FBRyxLQUFLLENBQUMsQ0FBQyxFQUFFLENBQUMsRUFBRSxDQUFDLEVBQUUsQ0FBQyxFQUFFLENBQUMsQ0FBQyxDQUFDLEdBQUcsQ0FBQyxDQUFDLEVBQUUsRUFBRSxFQUFFLFVBQVUsRUFBQztTQUMvQyxDQUFDLEdBQUcsS0FBSyxDQUFDLENBQUMsRUFBRSxDQUFDLEVBQUUsQ0FBQyxFQUFFLENBQUMsRUFBRSxDQUFDLENBQUMsQ0FBQyxHQUFHLENBQUMsQ0FBQyxFQUFFLEVBQUUsRUFBRSxDQUFDLFNBQVMsRUFBQztTQUMvQyxDQUFDLEdBQUcsS0FBSyxDQUFDLENBQUMsRUFBRSxDQUFDLEVBQUUsQ0FBQyxFQUFFLENBQUMsRUFBRSxDQUFDLENBQUMsQ0FBQyxHQUFHLEVBQUUsQ0FBQyxFQUFFLEVBQUUsRUFBRSxDQUFDLFVBQVUsRUFBQztTQUNqRCxDQUFDLEdBQUcsS0FBSyxDQUFDLENBQUMsRUFBRSxDQUFDLEVBQUUsQ0FBQyxFQUFFLENBQUMsRUFBRSxDQUFDLENBQUMsQ0FBQyxHQUFHLEVBQUUsQ0FBQyxFQUFFLENBQUMsRUFBRSxTQUFTLEVBQUM7U0FDOUMsQ0FBQyxHQUFHLEtBQUssQ0FBQyxDQUFDLEVBQUUsQ0FBQyxFQUFFLENBQUMsRUFBRSxDQUFDLEVBQUUsQ0FBQyxDQUFDLENBQUMsQ0FBQyxFQUFFLEVBQUUsRUFBRSxDQUFDLFNBQVMsRUFBQztTQUMzQyxDQUFDLEdBQUcsS0FBSyxDQUFDLENBQUMsRUFBRSxDQUFDLEVBQUUsQ0FBQyxFQUFFLENBQUMsRUFBRSxDQUFDLENBQUMsQ0FBQyxHQUFHLENBQUMsQ0FBQyxFQUFFLEVBQUUsRUFBRSxDQUFDLFNBQVMsRUFBQztTQUMvQyxDQUFDLEdBQUcsS0FBSyxDQUFDLENBQUMsRUFBRSxDQUFDLEVBQUUsQ0FBQyxFQUFFLENBQUMsRUFBRSxDQUFDLENBQUMsQ0FBQyxHQUFHLENBQUMsQ0FBQyxFQUFFLEVBQUUsRUFBRSxRQUFRLEVBQUM7U0FDN0MsQ0FBQyxHQUFHLEtBQUssQ0FBQyxDQUFDLEVBQUUsQ0FBQyxFQUFFLENBQUMsRUFBRSxDQUFDLEVBQUUsQ0FBQyxDQUFDLENBQUMsR0FBRyxDQUFDLENBQUMsRUFBRSxDQUFDLEVBQUUsQ0FBQyxTQUFTLEVBQUM7U0FDOUMsQ0FBQyxHQUFHLEtBQUssQ0FBQyxDQUFDLEVBQUUsQ0FBQyxFQUFFLENBQUMsRUFBRSxDQUFDLEVBQUUsQ0FBQyxDQUFDLENBQUMsR0FBRyxFQUFFLENBQUMsRUFBRSxFQUFFLEVBQUUsQ0FBQyxTQUFTLEVBQUM7U0FDaEQsQ0FBQyxHQUFHLEtBQUssQ0FBQyxDQUFDLEVBQUUsQ0FBQyxFQUFFLENBQUMsRUFBRSxDQUFDLEVBQUUsQ0FBQyxDQUFDLENBQUMsR0FBRyxFQUFFLENBQUMsRUFBRSxFQUFFLEVBQUUsU0FBUyxFQUFDO1NBQy9DLENBQUMsR0FBRyxLQUFLLENBQUMsQ0FBQyxFQUFFLENBQUMsRUFBRSxDQUFDLEVBQUUsQ0FBQyxFQUFFLENBQUMsQ0FBQyxDQUFDLEdBQUcsQ0FBQyxDQUFDLEVBQUUsRUFBRSxFQUFFLENBQUMsU0FBUyxFQUFDO0FBQ3JEO1NBQ00sQ0FBQyxHQUFHLEtBQUssQ0FBQyxDQUFDLEVBQUUsQ0FBQyxFQUFFLENBQUMsRUFBRSxDQUFDLEVBQUUsQ0FBQyxDQUFDLENBQUMsQ0FBQyxFQUFFLENBQUMsRUFBRSxDQUFDLFNBQVMsRUFBQztTQUMxQyxDQUFDLEdBQUcsS0FBSyxDQUFDLENBQUMsRUFBRSxDQUFDLEVBQUUsQ0FBQyxFQUFFLENBQUMsRUFBRSxDQUFDLENBQUMsQ0FBQyxHQUFHLENBQUMsQ0FBQyxFQUFFLEVBQUUsRUFBRSxVQUFVLEVBQUM7U0FDL0MsQ0FBQyxHQUFHLEtBQUssQ0FBQyxDQUFDLEVBQUUsQ0FBQyxFQUFFLENBQUMsRUFBRSxDQUFDLEVBQUUsQ0FBQyxDQUFDLENBQUMsR0FBRyxFQUFFLENBQUMsRUFBRSxFQUFFLEVBQUUsQ0FBQyxVQUFVLEVBQUM7U0FDakQsQ0FBQyxHQUFHLEtBQUssQ0FBQyxDQUFDLEVBQUUsQ0FBQyxFQUFFLENBQUMsRUFBRSxDQUFDLEVBQUUsQ0FBQyxDQUFDLENBQUMsR0FBRyxDQUFDLENBQUMsRUFBRSxFQUFFLEVBQUUsQ0FBQyxRQUFRLEVBQUM7U0FDOUMsQ0FBQyxHQUFHLEtBQUssQ0FBQyxDQUFDLEVBQUUsQ0FBQyxFQUFFLENBQUMsRUFBRSxDQUFDLEVBQUUsQ0FBQyxDQUFDLENBQUMsR0FBRyxFQUFFLENBQUMsRUFBRSxDQUFDLEVBQUUsVUFBVSxFQUFDO1NBQy9DLENBQUMsR0FBRyxLQUFLLENBQUMsQ0FBQyxFQUFFLENBQUMsRUFBRSxDQUFDLEVBQUUsQ0FBQyxFQUFFLENBQUMsQ0FBQyxDQUFDLEdBQUcsQ0FBQyxDQUFDLEVBQUUsRUFBRSxFQUFFLENBQUMsVUFBVSxFQUFDO1NBQ2hELENBQUMsR0FBRyxLQUFLLENBQUMsQ0FBQyxFQUFFLENBQUMsRUFBRSxDQUFDLEVBQUUsQ0FBQyxFQUFFLENBQUMsQ0FBQyxDQUFDLEdBQUcsRUFBRSxDQUFDLEVBQUUsRUFBRSxFQUFFLENBQUMsT0FBTyxFQUFDO1NBQzlDLENBQUMsR0FBRyxLQUFLLENBQUMsQ0FBQyxFQUFFLENBQUMsRUFBRSxDQUFDLEVBQUUsQ0FBQyxFQUFFLENBQUMsQ0FBQyxDQUFDLEdBQUcsQ0FBQyxDQUFDLEVBQUUsRUFBRSxFQUFFLENBQUMsVUFBVSxFQUFDO1NBQ2hELENBQUMsR0FBRyxLQUFLLENBQUMsQ0FBQyxFQUFFLENBQUMsRUFBRSxDQUFDLEVBQUUsQ0FBQyxFQUFFLENBQUMsQ0FBQyxDQUFDLEdBQUcsQ0FBQyxDQUFDLEVBQUUsQ0FBQyxFQUFFLFVBQVUsRUFBQztTQUM5QyxDQUFDLEdBQUcsS0FBSyxDQUFDLENBQUMsRUFBRSxDQUFDLEVBQUUsQ0FBQyxFQUFFLENBQUMsRUFBRSxDQUFDLENBQUMsQ0FBQyxHQUFHLEVBQUUsQ0FBQyxFQUFFLEVBQUUsRUFBRSxDQUFDLFFBQVEsRUFBQztTQUMvQyxDQUFDLEdBQUcsS0FBSyxDQUFDLENBQUMsRUFBRSxDQUFDLEVBQUUsQ0FBQyxFQUFFLENBQUMsRUFBRSxDQUFDLENBQUMsQ0FBQyxHQUFHLENBQUMsQ0FBQyxFQUFFLEVBQUUsRUFBRSxDQUFDLFVBQVUsRUFBQztTQUNoRCxDQUFDLEdBQUcsS0FBSyxDQUFDLENBQUMsRUFBRSxDQUFDLEVBQUUsQ0FBQyxFQUFFLENBQUMsRUFBRSxDQUFDLENBQUMsQ0FBQyxHQUFHLEVBQUUsQ0FBQyxFQUFFLEVBQUUsRUFBRSxVQUFVLEVBQUM7U0FDaEQsQ0FBQyxHQUFHLEtBQUssQ0FBQyxDQUFDLEVBQUUsQ0FBQyxFQUFFLENBQUMsRUFBRSxDQUFDLEVBQUUsQ0FBQyxDQUFDLENBQUMsR0FBRyxDQUFDLENBQUMsRUFBRSxDQUFDLEVBQUUsQ0FBQyxTQUFTLEVBQUM7U0FDOUMsQ0FBQyxHQUFHLEtBQUssQ0FBQyxDQUFDLEVBQUUsQ0FBQyxFQUFFLENBQUMsRUFBRSxDQUFDLEVBQUUsQ0FBQyxDQUFDLENBQUMsR0FBRyxFQUFFLENBQUMsRUFBRSxFQUFFLEVBQUUsQ0FBQyxVQUFVLEVBQUM7U0FDakQsQ0FBQyxHQUFHLEtBQUssQ0FBQyxDQUFDLEVBQUUsQ0FBQyxFQUFFLENBQUMsRUFBRSxDQUFDLEVBQUUsQ0FBQyxDQUFDLENBQUMsR0FBRyxDQUFDLENBQUMsRUFBRSxFQUFFLEVBQUUsU0FBUyxFQUFDO1NBQzlDLENBQUMsR0FBRyxLQUFLLENBQUMsQ0FBQyxFQUFFLENBQUMsRUFBRSxDQUFDLEVBQUUsQ0FBQyxFQUFFLENBQUMsQ0FBQyxDQUFDLEdBQUcsQ0FBQyxDQUFDLEVBQUUsRUFBRSxFQUFFLENBQUMsU0FBUyxFQUFDO0FBQ3JEO0VBQ0EsT0FBTSxDQUFDLEdBQUcsT0FBTyxDQUFDLENBQUMsRUFBRSxJQUFJLEVBQUM7RUFDMUIsT0FBTSxDQUFDLEdBQUcsT0FBTyxDQUFDLENBQUMsRUFBRSxJQUFJLEVBQUM7RUFDMUIsT0FBTSxDQUFDLEdBQUcsT0FBTyxDQUFDLENBQUMsRUFBRSxJQUFJLEVBQUM7RUFDMUIsT0FBTSxDQUFDLEdBQUcsT0FBTyxDQUFDLENBQUMsRUFBRSxJQUFJLEVBQUM7UUFDckI7T0FDRCxPQUFPLENBQUMsQ0FBQyxFQUFFLENBQUMsRUFBRSxDQUFDLEVBQUUsQ0FBQyxDQUFDO01BQ3BCO0FBQ0g7RUFDQTtFQUNBO0VBQ0E7RUFDQTtFQUNBO0VBQ0E7RUFDQSxHQUFFLFNBQVMsU0FBUyxDQUFDLEtBQUssRUFBRTtFQUM1QixLQUFJLElBQUksRUFBQztPQUNMLElBQUksTUFBTSxHQUFHLEdBQUU7RUFDbkIsS0FBSSxJQUFJLFFBQVEsR0FBRyxLQUFLLENBQUMsTUFBTSxHQUFHLEdBQUU7RUFDcEMsS0FBSSxLQUFLLENBQUMsR0FBRyxDQUFDLEVBQUUsQ0FBQyxHQUFHLFFBQVEsRUFBRSxDQUFDLElBQUksQ0FBQyxFQUFFO0VBQ3RDLE9BQU0sTUFBTSxJQUFJLE1BQU0sQ0FBQyxZQUFZLENBQUMsQ0FBQyxLQUFLLENBQUMsQ0FBQyxJQUFJLENBQUMsQ0FBQyxLQUFLLENBQUMsR0FBRyxFQUFFLElBQUksSUFBSSxFQUFDO1FBQ2pFO0VBQ0wsS0FBSSxPQUFPLE1BQU07TUFDZDtBQUNIO0VBQ0E7RUFDQTtFQUNBO0VBQ0E7RUFDQTtFQUNBO0VBQ0E7RUFDQSxHQUFFLFNBQVMsU0FBUyxDQUFDLEtBQUssRUFBRTtFQUM1QixLQUFJLElBQUksRUFBQztPQUNMLElBQUksTUFBTSxHQUFHLEdBQUU7RUFDbkIsS0FBSSxNQUFNLENBQUMsQ0FBQyxLQUFLLENBQUMsTUFBTSxJQUFJLENBQUMsSUFBSSxDQUFDLENBQUMsR0FBRyxVQUFTO0VBQy9DLEtBQUksS0FBSyxDQUFDLEdBQUcsQ0FBQyxFQUFFLENBQUMsR0FBRyxNQUFNLENBQUMsTUFBTSxFQUFFLENBQUMsSUFBSSxDQUFDLEVBQUU7RUFDM0MsT0FBTSxNQUFNLENBQUMsQ0FBQyxDQUFDLEdBQUcsRUFBQztRQUNkO0VBQ0wsS0FBSSxJQUFJLE9BQU8sR0FBRyxLQUFLLENBQUMsTUFBTSxHQUFHLEVBQUM7RUFDbEMsS0FBSSxLQUFLLENBQUMsR0FBRyxDQUFDLEVBQUUsQ0FBQyxHQUFHLE9BQU8sRUFBRSxDQUFDLElBQUksQ0FBQyxFQUFFO1NBQy9CLE1BQU0sQ0FBQyxDQUFDLElBQUksQ0FBQyxDQUFDLElBQUksQ0FBQyxLQUFLLENBQUMsVUFBVSxDQUFDLENBQUMsR0FBRyxDQUFDLENBQUMsR0FBRyxJQUFJLEtBQUssQ0FBQyxHQUFHLEdBQUU7UUFDN0Q7RUFDTCxLQUFJLE9BQU8sTUFBTTtNQUNkO0FBQ0g7RUFDQTtFQUNBO0VBQ0E7RUFDQTtFQUNBO0VBQ0E7RUFDQSxHQUFFLFNBQVMsT0FBTyxDQUFDLENBQUMsRUFBRTtFQUN0QixLQUFJLE9BQU8sU0FBUyxDQUFDLE9BQU8sQ0FBQyxTQUFTLENBQUMsQ0FBQyxDQUFDLEVBQUUsQ0FBQyxDQUFDLE1BQU0sR0FBRyxDQUFDLENBQUMsQ0FBQztNQUN0RDtBQUNIO0VBQ0E7RUFDQTtFQUNBO0VBQ0E7RUFDQTtFQUNBO0VBQ0E7RUFDQSxHQUFFLFNBQVMsV0FBVyxDQUFDLEdBQUcsRUFBRSxJQUFJLEVBQUU7RUFDbEMsS0FBSSxJQUFJLEVBQUM7RUFDVCxLQUFJLElBQUksSUFBSSxHQUFHLFNBQVMsQ0FBQyxHQUFHLEVBQUM7T0FDekIsSUFBSSxJQUFJLEdBQUcsR0FBRTtPQUNiLElBQUksSUFBSSxHQUFHLEdBQUU7RUFDakIsS0FBSSxJQUFJLEtBQUk7T0FDUixJQUFJLENBQUMsRUFBRSxDQUFDLEdBQUcsSUFBSSxDQUFDLEVBQUUsQ0FBQyxHQUFHLFVBQVM7RUFDbkMsS0FBSSxJQUFJLElBQUksQ0FBQyxNQUFNLEdBQUcs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QUMsQ0FBQyxNQUFNLENBQUMsQ0FBQyxDQUFDLENBQUMsRUFBQyxDQUFDLGFBQWEsQ0FBQyxDQUFDLENBQUMsQ0FBQyxPQUFPLENBQUMsQ0FBQyxNQUFNLENBQUMsSUFBSSxDQUFDLE1BQU0sQ0FBQyxNQUFNLENBQUMsQ0FBQyxDQUFDLENBQUMsSUFBSSxDQUFDLENBQUMsQ0FBQyxDQUFDLENBQUMsSUFBSSxDQUFDLENBQUMsT0FBTyxDQUFDLENBQUMsRUFBRSxDQUFDLENBQUMsQ0FBQyxDQUFDLENBQUMsQ0FBQyxDQUFDLENBQUMsQ0FBQyxDQUFDLENBQUMsQ0FBQyxDQUFDLENBQUMsT0FBTyxJQUFJLENBQUMsQ0FBQyxDQUFDLENBQUMsQ0FBQyxDQUFDLElBQUksQ0FBQyxDQUFDLENBQUMsQ0FBQyxDQUFDLEVBQUUsRUFBRSxDQUFDLENBQUMsSUFBSSxDQUFDLENBQUMsQ0FBQyxDQUFDLFdBQVcsQ0FBQyxDQUFDLENBQUMsTUFBTSxDQUFDLENBQUMsV0FBVyxDQUFDLENBQUMsQ0FBQyxDQUFDLElBQUksQ0FBQyxJQUFJLENBQUMsRUFBQyxDQUFDLENBQUMsSUFBSSxDQUFDLENBQUMsQ0FBQyxDQUFDLEtBQUssQ0FBQyxDQUFDLENBQUMsTUFBTSxDQUFDLFNBQVMsQ0FBQyxDQUFDLFdBQVcsQ0FBQyxDQUFDLENBQUMsQ0FBQyxDQUFDLENBQUMsS0FBSyxDQUFDLENBQUMsQ0FBQyxDQUFDLElBQUksQ0FBQyxRQUFRLENBQUMsQ0FBQyxDQUFDLFFBQVEsQ0FBQyxJQUFJLENBQUMsUUFBUSxDQUFDLENBQUMsQ0FBQyxRQUFRLENBQUMsSUFBSSxDQUFDLGFBQWEsQ0FBQyxDQUFDLENBQUMsY0FBYSxDQUFDLENBQUMsTUFBTSxDQUFDLENBQUMsUUFBUSxDQUFDLENBQUMsQ0FBQyxTQUFTLENBQUMsQ0FBQyxDQUFDLFFBQVEsQ0FBQyxDQUFDLENBQUMsUUFBUSxDQUFDLENBQUMsQ0FBQyxZQUFZLENBQUMsQ0FBQyxDQUFDLFFBQVEsRUFBRSxPQUFPLE1BQU0sRUFBRSxNQUFNLENBQUMsZ0JBQWdCLENBQUMsQ0FBQyxDQUFDLFlBQVksR0FBRyxJQUFJLENBQUMsQ0FBQyxDQUFDLENBQUMsT0FBTyxDQUFDLE1BQU0sQ0FBQyxrQkFBa0IsQ0FBQyxRQUFRLENBQUMsUUFBUSxDQUFDLFVBQVUsQ0FBQyxDQUFDLENBQUMsVUFBVSxDQUFDLFFBQVEsQ0FBQyxDQUFDLENBQUMsQ0FBQyxDQUFDLEdBQUcsQ0FBQyxDQUFDLGNBQWMsRUFBRSxDQUFDLENBQUMsQ0FBQyxXQUFXLENBQUMsZUFBZSxDQUFDLENBQUMsQ0FBQyxDQUFDLENBQUMsRUFBRSxDQUFDLENBQUMsTUFBTSxDQUFDLENBQUMsQ0FBQyxDQUFDLENBQUMsQ0FBQyxDQUFDLElBQUksQ0FBQyxDQUFDLENBQUMsQ0FBQyxDQUFDLENBQUMsQ0FBQyxDQUFDLENBQUMsQ0FBQyxFQUFFLENBQUMsQ0FBQyxDQUFDLENBQUMsQ0FBQyxDQUFDLENBQUMsQ0FBQyxDQUFDLENBQUMsQ0FBQyxDQUFDLENBQUMsQ0FBQyxDQUFDLENBQUMsQ0FBQyxDQUFDLENBQUMsQ0FBQyxDQUFDLENBQUMsQ0FBQyxDQUFDLENBQUMsQ0FBQyxDQUFDLENBQUMsQ0FBQyxDQUFDLENBQUMsQ0FBQyxDQUFDLENBQUMsQ0FBQyxDQUFDLENBQUMsQ0FBQyxDQUFDLENBQUMsQ0FBQyxDQUFDLENBQUMsQ0FBQyxDQUFDLENBQUMsQ0FBQyxDQUFDLENBQUMsQ0FBQyxDQUFDLENBQUMsR0FBRyxDQUFDLENBQUMsQ0FBQyxDQUFDLENBQUMsQ0FBQyxDQUFDLENBQUMsQ0FBQyxDQUFDLEdBQUcsQ0FBQyxJQUFJLElBQUksQ0FBQyxDQUFDLENBQUMsQ0FBQyxDQUFDLENBQUMsQ0FBQyxDQUFDLENBQUMsQ0FBQyxDQUFDLENBQUMsQ0FBQyxFQUFFLENBQUMsQ0FBQyxDQUFDLE1BQU0sQ0FBQyxDQUFDLENBQUMsQ0FBQyxDQUFDLENBQUMsQ0FBQyxDQUFDLENBQUMsQ0FBQyxDQUFDLENBQUMsQ0FBQyxDQUFDLEdBQUcsQ0FBQyxFQUFFLENBQUMsRUFBRSxDQUFDLEVBQUUsQ0FBQyxFQUFFLENBQUMsQ0FBQyxDQUFDLENBQUMsT0FBTyxNQUFNLENBQUMsSUFBSSxDQUFDLENBQUMsQ0FBQyxDQUFDLEtBQUssQ0FBQyxDQUFDLENBQUMsQ0FBQyxDQUFDLENBQUMsQ0FBQyxDQUFDLE9BQU0sRUFBRSxDQUFDLENBQUMsQ0FBQyxDQUFDLElBQUksUUFBUSxDQUFDLGVBQWUsR0FBRyxDQUFDLENBQUMsQ0FBQyxDQUFDLFFBQVEsQ0FBQyxRQUFRLEVBQUUsQ0FBQyxDQUFDLENBQUMsQ0FBQyxDQUFDLE1BQU0sQ0FBQyxTQUFTLENBQUMsQ0FBQyxXQUFXLEVBQUUsQ0FBQyxLQUFLLEVBQUUsQ0FBQyxNQUFNLENBQUMsQ0FBQyxDQUFDLEVBQUUsQ0FBQyxJQUFJLENBQUMsQ0FBQyxDQUFDLENBQUMsQ0FBQyxJQUFJLENBQUMsQ0FBQyxDQUFDLEVBQUUsQ0FBQyxJQUFJLENBQUMsQ0FBQyxDQUFDLElBQUksQ0FBQyxDQUFDLENBQUMsSUFBSSxDQUFDLElBQUksQ0FBQyxDQUFDLENBQUMsQ0FBQyxPQUFPLENBQUMsQ0FBQyxFQUFFLGdCQUFnQixDQUFDLENBQUMsQ0FBQyxJQUFJLENBQUMsQ0FBQyxDQUFDLENBQUMsQ0FBQyxDQUFDLENBQUMsQ0FBQyxDQUFDLEVBQUUsZ0JBQWdCLENBQUMsY0FBYyxDQUFDLENBQUMsRUFBRSxDQUFDLElBQUksQ0FBQyxDQUFDLENBQUMsVUFBVSxDQUFDLElBQUksQ0FBQyxDQUFDLENBQUMsZ0JBQWdCLEVBQUUsQ0FBQyxDQUFDLFdBQVcsQ0FBQyxNQUFNLENBQUMsQ0FBQyxFQUFFLElBQUksQ0FBQyxDQUFDLENBQUMsQ0FBQyxDQUFDLENBQUMsRUFBQyxDQUFDLENBQUMsQ0FBQyxFQUFDLENBQUMsRUFBQyxDQUFDLElBQUksS0FBSyxFQUFFLENBQUMsT0FBTyxJQUFJLENBQUMsQ0FBQyxDQUFDLElBQUksZ0JBQWdCLEVBQUUsQ0FBQyxPQUFPLFFBQVEsQ0FBQyxZQUFZLEVBQUUsQ0FBQyxDQUFDLENBQUMsaUJBQWlCLENBQUMsQ0FBQyxDQUFDLENBQUMsQ0FBQyxJQUFJLENBQUMsQ0FBQyxDQUFDLE9BQU8sQ0FBQyxDQUFDLENBQUMsQ0FBQyxDQUFDLENBQUMsR0FBRyxDQUFDLEdBQUcsSUFBSSxDQUFDLENBQUMsQ0FBQyxNQUFNLEVBQUUsZ0JBQWdCLENBQUMsY0FBYyxDQUFDLENBQUMsQ0FBQyxDQUFDLElBQUksQ0FBQyxDQUFDLENBQUMsSUFBSSxDQUFDLENBQUMsQ0FBQyxFQUFDLENBQUMsb0JBQW9CLENBQUMsQ0FBQyxDQUFDLENBQUMsTUFBTSxDQUFDLENBQUMsSUFBSSxDQUFDLENBQUMsQ0FBQyxPQUFPLENBQUMsQ0FBQyxDQUFDLENBQUMsQ0FBQyxDQUFDLENBQUMsQ0FBQyxHQUFHLElBQUksQ0FBQyxDQUFDLENBQUMsTUFBTSxDQUFDLENBQUMsQ0FBQyxDQUFDLENBQUMsQ0FBQyxDQUFDLEdBQUcsSUFBSSxDQUFDLENBQUMsQ0FBQyxNQUFNLEVBQUUsbUJBQW1CLENBQUMsY0FBYyxDQUFDLENBQUMsQ0FBQyxFQUFDLENBQUMsQ0FBQyxDQUFDLENBQUMsQ0FBQyxDQUFDLENBQUMsQ0FBQyxHQUFHLENBQUMsR0FBRyxJQUFJLENBQUMsQ0FBQyxDQUFDLENBQUMsTUFBTSxDQUFDLENBQUMsSUFBSSxDQUFDLENBQUMsQ0FBQyxDQUFDLENBQUMsQ0FBQyxDQUFDLENBQUMsQ0FBQyxDQUFDLENBQUMsQ0FBQyxJQUFJLElBQUksQ0FBQyxDQUFDLENBQUMsQ0FBQyxDQUFDLENBQUMsQ0FBQyxDQUFDLE1BQU0sQ0FBQyxDQUFDLENBQUMsRUFBRSxDQUFDLENBQUMsQ0FBQyxNQUFNLENBQUMsQ0FBQyxDQUFDLENBQUMsQ0FBQyxDQUFDLENBQUMsQ0FBQyxDQUFDLENBQUMsQ0FBQyxDQUFDLENBQUMsQ0FBQyxDQUFDLENBQUMsSUFBSSxDQUFDLENBQUMsQ0FBQyxDQUFDLENBQUMsSUFBSSxDQUFDLGFBQWEsQ0FBQyxJQUFJLENBQUMsQ0FBQyxhQUFhLENBQUMsQ0FBQyxRQUFRLENBQUMsQ0FBQyxDQUFDLFFBQVEsQ0FBQyxDQUFDLENBQUMsYUFBYSxDQUFDLENBQUMsQ0FBQyxDQUFDLEVBQUMsQ0FBQyxDQUFDLENBQUMsQ0FBQyxDQUFDLENBQUMsQ0FBQyxDQUFDLE9BQU8sQ0FBQyxFQUFFLFlBQVksQ0FBQyxJQUFJLENBQUMsQ0FBQyxDQUFDLENBQUMsQ0FBQyxDQUFDLElBQUksRUFBRSxJQUFJLFVBQVUsQ0FBQyxJQUFJLFFBQVEsQ0FBQyxJQUFJLENBQUMsQ0FBQyxDQUFDLENBQUMsQ0FBQyxDQUFDLEVBQUUsQ0FBQyxDQUFDLE1BQU0sSUFBSSxPQUFPLENBQUMsSUFBSSxDQUFDLENBQUMsQ0FBQyxDQUFDLENBQUMsQ0FBQyxDQUFDLENBQUMsTUFBTSxJQUFJLE1BQU0sQ0FBQyxJQUFJLENBQUMsQ0FBQyxHQUFHLENBQUMsRUFBRSxJQUFJLENBQUMsQ0FBQyxDQUFDLENBQUMsQ0FBQyxDQUFDLEVBQUUsQ0FBQyxDQUFDLE1BQU0sSUFBSSxVQUFVLENBQUMsSUFBSSxRQUFRLENBQUMsSUFBSSxDQUFDLENBQUMsQ0FBQyxDQUFDLENBQUMsQ0FBQyxDQUFDLFNBQVMsQ0FBQyxDQUFDLENBQUMsQ0FBQyxDQUFDLENBQUMsTUFBTSxJQUFJLGtCQUFrQixDQUFDLElBQUksQ0FBQyxDQUFDLEdBQUcsQ0FBQyxFQUFFLElBQUksQ0FBQyxDQUFDLEdBQUcsQ0FBQyxFQUFFLElBQUksQ0FBQyxDQUFDLENBQUMsQ0FBQyxDQUFDLENBQUMsRUFBRSxDQUFDLENBQUMsTUFBTSxJQUFJLFFBQVEsQ0FBQyxJQUFJLENBQUMsQ0FBQyxDQUFDLENBQUMsQ0FBQyxDQUFDLEVBQUMsQ0FBQyxDQUFDLENBQUMsSUFBSSxDQUFDLENBQUMsSUFBSSxDQUFDOztFQ1RucUU7O0VBQ0EsSUFBTUMsU0FBUyxHQUFHO0VBQ2hCQyxFQUFBQSxrQkFBa0IsRUFBRSxNQURKO0VBRWhCQyxFQUFBQSxnQkFBZ0IsRUFBRSxNQUZGO0VBR2hCQyxFQUFBQSxxQkFBcUIsRUFBRSxPQUhQO0VBSWhCQyxFQUFBQSxxQkFBcUIsRUFBRSxPQUFBO0VBSlAsQ0FBbEIsQ0FBQTtFQU9BLElBQU1DLGdCQUFnQixHQUFHLFFBQXpCLENBQUE7RUFFQSxJQUFNQyxVQUFVLEdBQUc7RUFDakJDLEVBQUFBLE1BQU0sRUFBRSxRQURTO0VBRWpCQyxFQUFBQSxPQUFPLEVBQUUsU0FGUTtFQUdqQkMsRUFBQUEsTUFBTSxFQUFFLFFBSFM7RUFJakJDLEVBQUFBLE1BQU0sRUFBRSxRQUFBO0VBSlMsQ0FBbkIsQ0FBQTtBQWllQSxNQUFNQyxPQUFPLEdBQWE7RUFDeEJDLEVBQUFBLE1BQU0sRUFBRTtFQUNOQyxJQUFBQSxZQUFZLEVBQUUsRUFEUjtFQUVOQyxJQUFBQSxJQUFJLEVBQUUsa0NBRkE7TUFHTkMsZUFBZSxFQUFFLEtBQUssRUFIaEI7RUFJTkMsSUFBQUEsU0FBUyxFQUFFLE9BSkw7RUFLTkMsSUFBQUEsYUFBYSxFQUFFLEVBTFQ7RUFNTkMsSUFBQUEsZUFBZSxFQUFFLEVBTlg7RUFPTkMsSUFBQUEsT0FBTyxFQUFFLE9BUEg7RUFRTkMsSUFBQUEsU0FBUyxFQUFFLE9BUkw7RUFTTkMsSUFBQUEsU0FBUyxFQUFFLEVBVEw7RUFVTkMsSUFBQUEsT0FBTyxFQUFFLEtBQUE7S0FYYTtFQWN4QkMsRUFBQUEsT0FBTyxFQUFFLEtBZGU7RUFnQnhCQyxFQUFBQSxlQUFlLEVBQUUsS0FoQk87RUFpQnhCQyxFQUFBQSxXQUFXLEVBQUVDLFNBakJXO0VBa0J4QkMsRUFBQUEsYUFBYSxFQUFFRCxTQWxCUztFQW1CeEJFLEVBQUFBLGNBQWMsRUFBRUYsU0FuQlE7RUFvQnhCRyxFQUFBQSxlQUFlLEVBQUVILFNBcEJPO0VBcUJ4QkksRUFBQUEsMkJBQTJCLEVBQUVKLFNBckJMO0VBc0J4QkssRUFBQUEsdUJBQXVCLEVBQUUsQ0F0QkQ7RUF1QnhCQyxFQUFBQSxXQUFXLEVBQUVOLFNBdkJXO0VBeUJ4Qk8sRUFBQUEsVUFBVSxFQUFFO0VBQ1ZDLElBQUFBLEtBQUssRUFBRSxFQURHO0VBRVZDLElBQUFBLEdBQUcsRUFBRSxTQUFBLEdBQUEsQ0FBQ0MsSUFBRCxFQUFpQkMsSUFBakIsRUFBK0I7RUFDbEM7RUFDQSxNQUFBLElBQU1DLGFBQWEsR0FBRzNCLE9BQU8sQ0FBQzRCLFNBQVIsQ0FBa0I1QixPQUFPLENBQUNDLE1BQVIsQ0FBZUksU0FBZixHQUEyQixTQUE3QyxDQUF0QixDQUFBOztRQUNBLElBQ0UsQ0FBQ3NCLGFBQUQsS0FDQyxDQUFDLFVBQUQsRUFBYSxNQUFiLEVBQXFCLE9BQXJCLENBQUEsQ0FBOEJFLFFBQTlCLENBQXVDSixJQUF2QyxDQUNFQSxJQUFBQSxJQUFJLEtBQUssY0FBVCxJQUE0QkMsSUFBcUIsQ0FBQ0ksZUFBdEIsS0FBMEMsSUFGekUsQ0FERixFQUlFO1VBQ0E5QixPQUFPLENBQUMrQixnQkFBUixDQUF5QixFQUF6QixDQUFBLENBQUE7RUFDRCxPQUFBOztFQUVELE1BQUEsSUFBTUMsSUFBSSxHQUFVO0VBQUVQLFFBQUFBLElBQUksRUFBRUEsSUFBQUE7U0FBNUIsQ0FBQTtFQUNBTyxNQUFBQSxJQUFJLENBQUNQLElBQUQsQ0FBSixHQUFhQyxJQUFiLENBWmtDOztFQWVsQyxNQUFBLElBQUksQ0FBQyxVQUFELEVBQWEsY0FBYixFQUE2QixNQUE3QixFQUFxQyxPQUFyQyxDQUE4Q0csQ0FBQUEsUUFBOUMsQ0FBdURKLElBQXZELENBQUosRUFBa0U7VUFDaEVPLElBQUksQ0FBQ0MsSUFBTCxHQUFTQyxRQUFBLENBQUEsRUFBQSxFQUFRbEMsT0FBTyxDQUFDYyxXQUFoQixDQUFULENBQUE7VUFDQWtCLElBQUksQ0FBQ0csT0FBTCxHQUFZRCxRQUFBLENBQUEsRUFBQSxFQUFRbEMsT0FBTyxDQUFDaUIsY0FBaEIsQ0FBWixDQUFBO0VBQ0QsT0FBQTs7UUFFRGpCLE9BQU8sQ0FBQ3NCLFVBQVIsQ0FBbUJDLEtBQW5CLENBQXlCYSxJQUF6QixDQUE4QkosSUFBOUIsQ0FBQSxDQXBCa0M7O0VBc0JsQ2hDLE1BQUFBLE9BQU8sQ0FBQ3FDLGFBQVIsQ0FBc0JDLEdBQXRCLENBQTBCLE9BQTFCLEVBQW1DQyxJQUFJLENBQUNDLFNBQUwsQ0FBZXhDLE9BQU8sQ0FBQ3NCLFVBQVIsQ0FBbUJDLEtBQWxDLENBQW5DLENBQUEsQ0FBQTtPQXhCUTtFQTBCVjtFQUNBO0VBQ0E7RUFDQWtCLElBQUFBLG1CQUFtQixFQUFFLFNBQUEsbUJBQUEsR0FBQTtFQUNuQjtRQUNBLElBQUksQ0FBQ3pDLE9BQU8sQ0FBQ2tCLGVBQVQsSUFBNEIsQ0FBQ2xCLE9BQU8sQ0FBQ21CLDJCQUF6QyxFQUFzRTtFQUNwRSxRQUFBLE9BQUE7RUFDRCxPQUprQjs7O1FBT25CLElBQU11QixTQUFTLEdBQUdDLElBQUksQ0FBQ0MsS0FBTCxDQUNoQixDQUFDLElBQUlDLElBQUosRUFBQSxDQUFXQyxPQUFYLEVBQXVCOUMsR0FBQUEsT0FBTyxDQUFDbUIsMkJBQVIsQ0FBb0MyQixPQUFwQyxFQUF4QixJQUF5RSxJQUR6RCxDQUFsQixDQUFBO1FBR0E5QyxPQUFPLENBQUNvQix1QkFBUixJQUFtQ3NCLFNBQW5DLENBQUE7UUFDQTFDLE9BQU8sQ0FBQytDLEdBQVIsQ0FBWSxNQUFaLEVBQW9CLDJCQUFwQixFQUFpRC9DLE9BQU8sQ0FBQ29CLHVCQUF6RCxDQUFBLENBWG1COztRQWNuQixJQUFNNEIsUUFBUSxHQUFRZCxRQUFBLENBQUEsRUFBQSxFQUFBbEMsT0FBTyxDQUFDa0IsZUFBUixDQUF0QixDQUFBOztFQUNBLE1BQUEsSUFBTStCLEdBQUcsR0FBRyxJQUFJSixJQUFKLEVBQUEsQ0FBV0ssV0FBWCxFQUFaLENBQUE7RUFFQUYsTUFBQUEsUUFBUSxDQUFDRyxRQUFULEdBQW9CbkQsT0FBTyxDQUFDb0IsdUJBQTVCLENBQUE7RUFDQTRCLE1BQUFBLFFBQVEsQ0FBQ0ksVUFBVCxHQUFzQkgsR0FBdEIsQ0FsQm1COztFQXFCbkIsTUFBQSxJQUFJLENBQUNqRCxPQUFPLENBQUNpQixjQUFSLENBQXVCa0MsUUFBNUIsRUFBc0NuRCxPQUFPLENBQUNpQixjQUFSLENBQXVCa0MsUUFBdkIsR0FBa0MsQ0FBbEMsQ0FBQTtFQUN0Q25ELE1BQUFBLE9BQU8sQ0FBQ2lCLGNBQVIsQ0FBdUJrQyxRQUF2QixJQUFtQ1QsU0FBbkMsQ0FBQTtFQUNBMUMsTUFBQUEsT0FBTyxDQUFDaUIsY0FBUixDQUF1Qm1DLFVBQXZCLEdBQW9DSCxHQUFwQyxDQUFBO0VBQ0FqRCxNQUFBQSxPQUFPLENBQUNxRCxpQkFBUixDQUEwQnJELE9BQU8sQ0FBQ2lCLGNBQWxDLEVBeEJtQjtFQTBCbkI7RUFDQTtFQUNBO0VBQ0E7O0VBRUEsTUFBQSxJQUFNZSxJQUFJLEdBQUc7RUFDWFAsUUFBQUEsSUFBSSxFQUFFLFVBREs7RUFFWHVCLFFBQUFBLFFBQVEsRUFBRUEsUUFGQztVQUdYZixJQUFJLEVBQU9DLFFBQUEsQ0FBQSxFQUFBLEVBQUFsQyxPQUFPLENBQUNjLFdBQVIsQ0FIQTtVQUlYcUIsT0FBTyxFQUFPRCxRQUFBLENBQUEsRUFBQSxFQUFBbEMsT0FBTyxDQUFDaUIsY0FBUixDQUpIO0VBS1hxQyxRQUFBQSxNQUFNLEVBQU9wQixRQUFBLENBQUEsRUFBQSxFQUFBbEMsT0FBTyxDQUFDZ0IsYUFBUixDQUFBO1NBTGYsQ0FBQTtRQU9BaEIsT0FBTyxDQUFDc0IsVUFBUixDQUFtQkMsS0FBbkIsQ0FBeUJhLElBQXpCLENBQThCSixJQUE5QixDQUFBLENBdENtQjs7RUF3Q25CaEMsTUFBQUEsT0FBTyxDQUFDcUMsYUFBUixDQUFzQkMsR0FBdEIsQ0FBMEIsT0FBMUIsRUFBbUNDLElBQUksQ0FBQ0MsU0FBTCxDQUFleEMsT0FBTyxDQUFDc0IsVUFBUixDQUFtQkMsS0FBbEMsQ0FBbkMsQ0FBQSxDQUFBO0VBQ0QsS0FBQTtLQS9GcUI7RUFpR3hCZ0MsRUFBQUEsYUFBYSxFQUFFLEVBakdTO0VBa0d4QkMsRUFBQUEsYUFBYSxFQUFFLEtBbEdTO0VBbUd4QkMsRUFBQUEsWUFBWSxFQUFFLEVBbkdVO0lBcUd4QlYsR0FBRyxFQUFFLFNBQUNXLEdBQUFBLENBQUFBLEtBQUQsRUFBYztNQUFFLElBQU9DLElBQUEsR0FBQSxFQUFQLENBQUE7O1dBQUEsSUFBT0MsRUFBQSxHQUFBLEdBQVBBLEVBQU8sR0FBQUMsU0FBQSxDQUFBQyxRQUFQRixFQUFPLElBQUE7UUFBUEQsSUFBTyxDQUFBQyxFQUFBLEdBQUEsQ0FBQSxDQUFQLEdBQU9DLFNBQUEsQ0FBQUQsRUFBQSxDQUFQLENBQUE7OztFQUNuQixJQUFBLFFBQVFGLEtBQVI7RUFDRSxNQUFBLEtBQUssTUFBTDtFQUNFLFFBQUEsSUFBSSxDQUFDLE1BQUQsRUFBUyxNQUFULEVBQWlCLE9BQWpCLEVBQTBCLE9BQTFCLENBQW1DN0IsQ0FBQUEsUUFBbkMsQ0FBNEM3QixPQUFPLENBQUNDLE1BQVIsQ0FBZVEsU0FBM0QsQ0FBSixFQUEyRTtFQUN6RXNELFVBQUFBLE9BQU8sQ0FBQ0MsSUFBUixDQUFZQyxLQUFaLENBQUFGLE9BQUEsRUFBZ0JKLElBQWhCLENBQUEsQ0FBQTtFQUNELFNBQUE7O0VBQ0QsUUFBQSxNQUFBOztFQUNGLE1BQUEsS0FBSyxNQUFMO0VBQ0UsUUFBQSxJQUFJLENBQUMsTUFBRCxFQUFTLE9BQVQsRUFBa0IsT0FBbEIsQ0FBQSxDQUEyQjlCLFFBQTNCLENBQW9DN0IsT0FBTyxDQUFDQyxNQUFSLENBQWVRLFNBQW5ELENBQUosRUFBbUU7RUFDakVzRCxVQUFBQSxPQUFPLENBQUNHLElBQVIsQ0FBWUQsS0FBWixDQUFBRixPQUFBLEVBQWdCSixJQUFoQixDQUFBLENBQUE7RUFDRCxTQUFBOztFQUNELFFBQUEsTUFBQTs7RUFDRixNQUFBLEtBQUssT0FBTDtFQUNFLFFBQUEsSUFBSSxDQUFDLE9BQUQsRUFBVSxPQUFWLENBQW1COUIsQ0FBQUEsUUFBbkIsQ0FBNEI3QixPQUFPLENBQUNDLE1BQVIsQ0FBZVEsU0FBM0MsQ0FBSixFQUEyRDtFQUN6RHNELFVBQUFBLE9BQU8sQ0FBQ0ksS0FBUixDQUFhRixLQUFiLENBQUFGLE9BQUEsRUFBaUJKLElBQWpCLENBQUEsQ0FBQTtFQUNELFNBQUE7O0VBQ0QsUUFBQSxNQUFBOztFQUNGLE1BQUEsS0FBSyxPQUFMO0VBQ0UsUUFBQSxJQUFJM0QsT0FBTyxDQUFDQyxNQUFSLENBQWVRLFNBQWYsS0FBNkIsT0FBakMsRUFBMEM7RUFDeENzRCxVQUFBQSxPQUFPLENBQUNLLEtBQVIsQ0FBYUgsS0FBYixDQUFBRixPQUFBLEVBQWlCSixJQUFqQixDQUFBLENBQUE7RUFDRCxTQUFBOztFQUNELFFBQUEsTUFBQTtFQUNGOztFQUNBLE1BQUE7RUFDRUksUUFBQUEsT0FBTyxDQUFDTSxLQUFSLENBQWFKLEtBQWIsQ0FBQUYsT0FBQSxFQUFpQkosSUFBakIsQ0FBQSxDQUFBO0VBdkJKLEtBQUE7S0F0R3NCO0VBaUl4QjtJQUNBVyxJQUFJLEVBQUUsU0FBQ0MsSUFBQUEsQ0FBQUEsR0FBRCxFQUFhO0VBQ2pCO01BQ0EsSUFBSUMsUUFBUSxDQUFDQyxVQUFULEtBQXdCLFVBQXhCLElBQXNDRCxRQUFRLENBQUNDLFVBQVQsS0FBd0IsYUFBbEUsRUFBaUY7UUFDL0V6RSxPQUFPLENBQUMwRSxRQUFSLENBQWlCSCxHQUFqQixDQUFBLENBQUE7RUFDRCxLQUpnQjs7O01BT2pCQyxRQUFRLENBQUNHLGtCQUFULEdBQThCLFlBQUE7UUFDNUIsSUFBSUgsUUFBUSxDQUFDQyxVQUFULEtBQXdCLGFBQXhCLElBQXlDRCxRQUFRLENBQUNDLFVBQVQsS0FBd0IsVUFBckUsRUFBaUY7VUFDL0V6RSxPQUFPLENBQUMrQyxHQUFSLENBQVksTUFBWixFQUFvQixpQkFBcEIsRUFBdUN5QixRQUFRLENBQUNDLFVBQWhELENBQUEsQ0FBQTs7VUFDQXpFLE9BQU8sQ0FBQzBFLFFBQVIsQ0FBaUJILEdBQWpCLENBQUEsQ0FBQTtFQUNELE9BQUE7T0FKSCxDQUFBOztFQU9BLElBQUEsSUFBTUssUUFBUSxHQUFHNUUsT0FBTyxDQUFDNEIsU0FBUixDQUFrQjVCLE9BQU8sQ0FBQ0MsTUFBUixDQUFlSSxTQUFmLEdBQTJCLE9BQTdDLENBQWpCLENBQUE7O0VBQ0EsSUFBQSxJQUFJdUUsUUFBSixFQUFjO0VBQ1o1RSxNQUFBQSxPQUFPLENBQUNDLE1BQVIsQ0FBZVEsU0FBZixHQUEyQixNQUEzQixDQUFBO0VBQ0QsS0FBQTtLQW5KcUI7SUFzSnhCb0Usa0JBQWtCLEVBQUUsU0FBQ0Msa0JBQUFBLENBQUFBLE9BQUQsRUFBaUI7RUFDbkM5RSxJQUFBQSxPQUFPLENBQUMrQyxHQUFSLENBQVksTUFBWixFQUFvQiw2QkFBcEIsRUFBbUQrQixPQUFuRCxDQUFBLENBQUE7TUFDQTlFLE9BQU8sQ0FBQ2EsZUFBUixHQUEwQmlFLE9BQTFCLENBQUE7S0F4SnNCO0VBMkp4QjtJQUNBQyxjQUFjLEVBQUUsU0FBQ0MsY0FBQUEsQ0FBQUEsUUFBRCxFQUFnQztFQUM5QztNQUNBLElBQUloRixPQUFPLENBQUNjLFdBQVosRUFBeUI7RUFDdkJrRSxNQUFBQSxRQUFRLENBQUNoRixPQUFPLENBQUNjLFdBQVQsQ0FBUixDQUFBO0VBQ0EsTUFBQSxPQUFBO0VBQ0QsS0FBQTs7TUFFRGQsT0FBTyxDQUFDaUYsY0FBUixDQUF1QixZQUFBO0VBQ3JCRCxNQUFBQSxRQUFRLENBQUNoRixPQUFPLENBQUNjLFdBQVQsQ0FBUixDQUFBO09BREYsQ0FBQSxDQUFBO0tBbktzQjtJQXdLeEJvRSxPQUFPLEVBQUUsU0FBQ0MsT0FBQUEsQ0FBQUEsRUFBRCxFQUFlO01BQ3RCLElBQUluRixPQUFPLENBQUNZLE9BQVosRUFBcUI7UUFDbkJ1RSxFQUFFLEVBQUEsQ0FBQTtFQUNILEtBRkQsTUFFTztFQUNMbkYsTUFBQUEsT0FBTyxDQUFDeUQsWUFBUixDQUFxQnJCLElBQXJCLENBQTBCK0MsRUFBMUIsQ0FBQSxDQUFBO0VBQ0QsS0FBQTtLQTdLcUI7RUFnTHhCO0lBQ0FDLGFBQWEsRUFBRSxTQUFDMUQsYUFBQUEsQ0FBQUEsSUFBRCxFQUFVO0VBQ3ZCLElBQUEsSUFBSSxDQUFDMUIsT0FBTyxDQUFDWSxPQUFiLEVBQXNCO0VBQ3BCWixNQUFBQSxPQUFPLENBQUMrQyxHQUFSLENBQVksT0FBWixFQUFxQiw2REFBckIsQ0FBQSxDQUFBOztRQUNBL0MsT0FBTyxDQUFDaUYsY0FBUixDQUF1QixZQUFBO0VBQU0sUUFBQSxPQUFBakYsT0FBTyxDQUFDb0YsYUFBUixDQUFzQjFELElBQXRCLENBQUEsQ0FBQTtTQUE3QixDQUFBLENBQUE7O0VBQ0EsTUFBQSxPQUFBO0VBQ0QsS0FMc0I7RUFRdkI7OztNQUNBLElBQUkxQixPQUFPLENBQUNrQixlQUFSLElBQTJCLENBQUNsQixPQUFPLENBQUNtQiwyQkFBeEMsRUFBcUU7RUFDbkVuQixNQUFBQSxPQUFPLENBQUMrQyxHQUFSLENBQVksTUFBWixFQUFvQixzREFBcEIsQ0FBQSxDQUFBO1FBQ0EvQyxPQUFPLENBQUNrQixlQUFSLEdBQTBCSCxTQUExQixDQUFBO0VBQ0QsS0FBQTs7TUFFRCxJQUFNa0MsR0FBRyxHQUFHLElBQUlKLElBQUosR0FBV0ssV0FBWCxFQUFaLENBZHVCOztNQWlCdkIsSUFBSWxELE9BQU8sQ0FBQ2tCLGVBQVosRUFBNkI7RUFDM0JsQixNQUFBQSxPQUFPLENBQUMrQyxHQUFSLENBQVksTUFBWixFQUFvQiwwQkFBcEIsRUFEMkI7O1FBSTNCLElBQU1MLFNBQVMsR0FBR0MsSUFBSSxDQUFDQyxLQUFMLENBQ2hCLENBQUMsSUFBSUMsSUFBSixFQUFBLENBQVdDLE9BQVgsRUFBdUI5QyxHQUFBQSxPQUFPLENBQUNtQiwyQkFBUixDQUFvQzJCLE9BQXBDLEVBQXhCLElBQXlFLElBRHpELENBQWxCLENBQUE7UUFHQTlDLE9BQU8sQ0FBQ29CLHVCQUFSLElBQW1Dc0IsU0FBbkMsQ0FBQTtFQUNBMUMsTUFBQUEsT0FBTyxDQUFDa0IsZUFBUixDQUF3QmlDLFFBQXhCLEdBQW1DbkQsT0FBTyxDQUFDb0IsdUJBQTNDLENBQUE7RUFDQXBCLE1BQUFBLE9BQU8sQ0FBQ2tCLGVBQVIsQ0FBd0JrQyxVQUF4QixHQUFxQ0gsR0FBckMsQ0FUMkI7O0VBWTNCLE1BQUEsSUFBSSxDQUFDakQsT0FBTyxDQUFDaUIsY0FBUixDQUF1QmtDLFFBQTVCLEVBQXNDbkQsT0FBTyxDQUFDaUIsY0FBUixDQUF1QmtDLFFBQXZCLEdBQWtDLENBQWxDLENBQUE7RUFDdENuRCxNQUFBQSxPQUFPLENBQUNpQixjQUFSLENBQXVCa0MsUUFBdkIsSUFBbUNULFNBQW5DLENBQUE7RUFDQTFDLE1BQUFBLE9BQU8sQ0FBQ2lCLGNBQVIsQ0FBdUJtQyxVQUF2QixHQUFvQ0gsR0FBcEMsQ0FkMkI7RUFpQjNCOztFQUNBakQsTUFBQUEsT0FBTyxDQUFDc0IsVUFBUixDQUFtQkUsR0FBbkIsQ0FBdUIsVUFBdkIsRUFBaUNVLFFBQUEsQ0FBQSxFQUFBLEVBQU9sQyxPQUFPLENBQUNrQixlQUFmLENBQWpDLEVBbEIyQjs7UUFvQjNCbEIsT0FBTyxDQUFDbUIsMkJBQVIsR0FBc0NKLFNBQXRDLENBQUE7UUFDQWYsT0FBTyxDQUFDb0IsdUJBQVIsR0FBa0MsQ0FBbEMsQ0FBQTtFQUNELEtBdkNzQjs7O0VBMEN2QixJQUFBLElBQU00QixRQUFRLEdBQUdULElBQUksQ0FBQzhDLEtBQUwsQ0FBVzlDLElBQUksQ0FBQ0MsU0FBTCxDQUFlZCxJQUFJLElBQUksRUFBdkIsQ0FBWCxDQUFqQixDQUFBO0VBRUFzQixJQUFBQSxRQUFRLENBQUNzQyxXQUFULEdBQXVCdEYsT0FBTyxDQUFDdUYsTUFBUixFQUF2QixDQUFBO01BQ0F2QyxRQUFRLENBQUN3QyxVQUFULEdBQXNCdkMsR0FBdEIsQ0FBQTtFQUVBLElBQUEsSUFBTXdDLFFBQVEsR0FBR3pGLE9BQU8sQ0FBQzBGLFdBQVIsRUFBakIsQ0FBQTs7RUFDQSxJQUFBLElBQUlELFFBQUosRUFBYztRQUNaekMsUUFBUSxDQUFDeUMsUUFBVCxHQUFvQkEsUUFBcEIsQ0FBQTtFQUNELEtBbERzQjs7O0VBcUR2QixJQUFBLElBQUksQ0FBQ3pDLFFBQVEsQ0FBQzJDLEtBQWQsRUFBcUI7RUFDbkIzQyxNQUFBQSxRQUFRLENBQUMyQyxLQUFULEdBQWlCbkIsUUFBUSxDQUFDbUIsS0FBMUIsQ0FBQTtFQUNELEtBQUE7O0VBQ0QsSUFBQSxJQUFJLENBQUMzQyxRQUFRLENBQUM0QyxPQUFkLEVBQXVCO0VBQ3JCNUMsTUFBQUEsUUFBUSxDQUFDNEMsT0FBVCxHQUFtQkMsTUFBTSxDQUFDQyxRQUFQLENBQWdCQyxJQUFuQyxDQUFBO0VBQ0QsS0ExRHNCOzs7RUE2RHZCL0MsSUFBQUEsUUFBUSxDQUFDMkMsS0FBVCxHQUFpQjNDLFFBQVEsQ0FBQzJDLEtBQVQsQ0FBZUssT0FBZixDQUF1QixLQUF2QixFQUE4QixHQUE5QixDQUFqQixDQTdEdUI7O01BZ0V2QixJQUFJaEQsUUFBUSxDQUFDaUQsYUFBVCxJQUEwQmpELFFBQVEsQ0FBQ2lELGFBQVQsR0FBeUIsQ0FBdkQsRUFBMEQ7RUFDeERqRCxNQUFBQSxRQUFRLENBQUNpRCxhQUFULEdBQXlCdEQsSUFBSSxDQUFDQyxLQUFMLENBQVdJLFFBQVEsQ0FBQ2lELGFBQVQsR0FBeUIsR0FBcEMsQ0FBekIsQ0FBQTtFQUNELEtBbEVzQjs7O0VBcUV2QixJQUFBLElBQUlqRyxPQUFPLENBQUNrRyxhQUFSLEVBQUosRUFBNkI7RUFDM0JsRyxNQUFBQSxPQUFPLENBQUNtQiwyQkFBUixHQUFzQyxJQUFJMEIsSUFBSixFQUF0QyxDQUFBO0VBQ0QsS0F2RXNCO0VBMEV2QjtFQUNBO0VBQ0E7RUFFQTs7O0VBQ0E3QyxJQUFBQSxPQUFPLENBQUNrQixlQUFSLEdBQTBCOEIsUUFBMUIsQ0EvRXVCOztFQWtGdkJoRCxJQUFBQSxPQUFPLENBQUNjLFdBQVIsQ0FBb0JxRixtQkFBcEIsR0FBMENsRCxHQUExQyxDQUFBO0VBQ0FqRCxJQUFBQSxPQUFPLENBQUNvRyxjQUFSLENBQXVCcEcsT0FBTyxDQUFDYyxXQUEvQixFQW5GdUI7RUFxRnZCOztFQUNBLElBQUEsSUFBSSxDQUFDZCxPQUFPLENBQUNpQixjQUFSLENBQXVCb0YsZUFBNUIsRUFBNkNyRyxPQUFPLENBQUNpQixjQUFSLENBQXVCb0YsZUFBdkIsR0FBeUMsQ0FBekMsQ0FBQTtFQUM3QyxJQUFBLElBQUksQ0FBQ3JHLE9BQU8sQ0FBQ2lCLGNBQVIsQ0FBdUJxRixrQkFBNUIsRUFBZ0R0RyxPQUFPLENBQUNpQixjQUFSLENBQXVCcUYsa0JBQXZCLEdBQTRDLENBQTVDLENBQUE7TUFDaER0RyxPQUFPLENBQUNpQixjQUFSLENBQXVCb0YsZUFBdkIsRUFBQSxDQUFBO01BQ0FyRyxPQUFPLENBQUNpQixjQUFSLENBQXVCcUYsa0JBQXZCLEVBQUEsQ0FBQTtFQUNBdEcsSUFBQUEsT0FBTyxDQUFDaUIsY0FBUixDQUF1Qm1DLFVBQXZCLEdBQW9DSCxHQUFwQyxDQUFBO0VBQ0FqRCxJQUFBQSxPQUFPLENBQUNxRCxpQkFBUixDQUEwQnJELE9BQU8sQ0FBQ2lCLGNBQWxDLEVBM0Z1QjtFQTZGdkI7O0VBQ0FqQixJQUFBQSxPQUFPLENBQUNzQixVQUFSLENBQW1CRSxHQUFuQixDQUF1QixVQUF2QixFQUFpQ1UsUUFBQSxDQUFBLEVBQUEsRUFBT2MsUUFBUCxDQUFqQyxDQUFBLENBQUE7S0EvUXNCO0VBa1J4QjtJQUNBdUQsZ0JBQWdCLEVBQUUsU0FBQzdFLGdCQUFBQSxDQUFBQSxJQUFELEVBQVU7RUFDMUIsSUFBQSxJQUFJLENBQUMxQixPQUFPLENBQUNZLE9BQWIsRUFBc0I7RUFDcEJaLE1BQUFBLE9BQU8sQ0FBQytDLEdBQVIsQ0FBWSxPQUFaLEVBQXFCLGdFQUFyQixDQUFBLENBQUE7O1FBQ0EvQyxPQUFPLENBQUNpRixjQUFSLENBQXVCLFlBQUE7RUFBTSxRQUFBLE9BQUFqRixPQUFPLENBQUN1RyxnQkFBUixDQUF5QjdFLElBQXpCLENBQUEsQ0FBQTtTQUE3QixDQUFBLENBQUE7O0VBQ0EsTUFBQSxPQUFBO0VBQ0QsS0FBQTs7RUFFRCxJQUFBLElBQUlBLElBQUksSUFBSSxDQUFDQSxJQUFJLENBQUM4RSxLQUFsQixFQUF5QjtFQUN2QnhHLE1BQUFBLE9BQU8sQ0FBQytDLEdBQVIsQ0FBWSxPQUFaLEVBQXFCLCtCQUFyQixDQUFBLENBQUE7RUFDQSxNQUFBLE9BQUE7RUFDRCxLQVZ5Qjs7O0VBYTFCLElBQUEsSUFBTTBELFdBQVcsR0FBR2xFLElBQUksQ0FBQzhDLEtBQUwsQ0FBVzlDLElBQUksQ0FBQ0MsU0FBTCxDQUFlZCxJQUFJLElBQUksRUFBdkIsQ0FBWCxDQUFwQixDQUFBO0VBQ0EsSUFBQSxJQUFNdUIsR0FBRyxHQUFHLElBQUlKLElBQUosRUFBQSxDQUFXSyxXQUFYLEVBQVosQ0FBQTtFQUVBdUQsSUFBQUEsV0FBVyxDQUFDbkIsV0FBWixHQUEwQnRGLE9BQU8sQ0FBQ3VGLE1BQVIsRUFBMUIsQ0FBQTtFQUNBa0IsSUFBQUEsV0FBVyxDQUFDakIsVUFBWixHQUF5QnZDLEdBQXpCLENBakIwQjs7RUFvQjFCd0QsSUFBQUEsV0FBVyxDQUFDRCxLQUFaLEdBQW9CQyxXQUFXLENBQUNELEtBQVosQ0FBa0JSLE9BQWxCLENBQTBCLEtBQTFCLEVBQWlDLEdBQWpDLENBQXBCLENBQUE7O01BRUEsSUFBSVMsV0FBVyxDQUFDQyxZQUFoQixFQUE4QjtFQUM1QjtFQUNBRCxNQUFBQSxXQUFXLENBQUNDLFlBQVosR0FBMkJELFdBQVcsQ0FBQ0MsWUFBWixDQUF5QlYsT0FBekIsQ0FBaUMsS0FBakMsRUFBd0MsR0FBeEMsQ0FBM0IsQ0FBQTtFQUNELEtBekJ5QjtFQTRCMUI7RUFDQTtFQUNBO0VBRUE7OztFQUNBLElBQUEsSUFBSSxDQUFDUyxXQUFXLENBQUMzRSxlQUFqQixFQUFrQztFQUNoQzlCLE1BQUFBLE9BQU8sQ0FBQ2MsV0FBUixDQUFvQnFGLG1CQUFwQixHQUEwQ2xELEdBQTFDLENBQUE7RUFDQWpELE1BQUFBLE9BQU8sQ0FBQ29HLGNBQVIsQ0FBdUJwRyxPQUFPLENBQUNjLFdBQS9CLEVBRmdDO0VBSWhDOztFQUNBLE1BQUEsSUFBSSxDQUFDZCxPQUFPLENBQUNpQixjQUFSLENBQXVCcUYsa0JBQTVCLEVBQWdEdEcsT0FBTyxDQUFDaUIsY0FBUixDQUF1QnFGLGtCQUF2QixHQUE0QyxDQUE1QyxDQUFBO1FBQ2hEdEcsT0FBTyxDQUFDaUIsY0FBUixDQUF1QnFGLGtCQUF2QixFQUFBLENBQUE7RUFDQXRHLE1BQUFBLE9BQU8sQ0FBQ2lCLGNBQVIsQ0FBdUJtQyxVQUF2QixHQUFvQ0gsR0FBcEMsQ0FBQTtFQUNBakQsTUFBQUEsT0FBTyxDQUFDcUQsaUJBQVIsQ0FBMEJyRCxPQUFPLENBQUNpQixjQUFsQyxFQVJnQztFQVNqQyxLQTFDeUI7OztFQTZDMUJqQixJQUFBQSxPQUFPLENBQUNzQixVQUFSLENBQW1CRSxHQUFuQixDQUF1QixjQUF2QixFQUFxQ1UsUUFBQSxDQUFBLEVBQUEsRUFBT3VFLFdBQVAsQ0FBckMsQ0FBQSxDQUFBO0tBaFVzQjtJQW1VeEJFLFNBQVMsRUFBRSxTQUFDakYsU0FBQUEsQ0FBQUEsSUFBRCxFQUFZO0VBQ3JCLElBQUEsSUFBSSxDQUFDMUIsT0FBTyxDQUFDWSxPQUFiLEVBQXNCO0VBQ3BCWixNQUFBQSxPQUFPLENBQUMrQyxHQUFSLENBQVksT0FBWixFQUFxQix5REFBckIsQ0FBQSxDQUFBOztRQUNBL0MsT0FBTyxDQUFDaUYsY0FBUixDQUF1QixZQUFBO0VBQU0sUUFBQSxPQUFBakYsT0FBTyxDQUFDMkcsU0FBUixDQUFrQmpGLElBQWxCLENBQUEsQ0FBQTtTQUE3QixDQUFBLENBQUE7O0VBQ0EsTUFBQSxPQUFBO0VBQ0QsS0FMb0I7OztFQVFyQixJQUFBLElBQUksQ0FBQ0EsSUFBRCxJQUFTLFFBQU9BLElBQVAsQ0FBQSxLQUFnQixRQUE3QixFQUF1QztFQUNyQzFCLE1BQUFBLE9BQU8sQ0FBQytDLEdBQVIsQ0FBWSxPQUFaLEVBQXFCLG1CQUFyQixDQUFBLENBQUE7RUFDQSxNQUFBLE9BQUE7RUFDRCxLQVhvQjs7O0VBYXJCLElBQUEsSUFBTTZELElBQUksR0FBR3JFLElBQUksQ0FBQzhDLEtBQUwsQ0FBVzlDLElBQUksQ0FBQ0MsU0FBTCxDQUFlZCxJQUFmLENBQVgsQ0FBYixDQUFBO0VBRUExQixJQUFBQSxPQUFPLENBQUMrQyxHQUFSLENBQVksTUFBWixFQUFvQixTQUFwQixFQUErQjZELElBQS9CLENBQUEsQ0FBQTs7RUFFQSxJQUFBLElBQUksQ0FBQ0EsSUFBSSxDQUFDdEIsV0FBVixFQUF1QjtFQUNyQnNCLE1BQUFBLElBQUksQ0FBQ3RCLFdBQUwsR0FBbUJ0RixPQUFPLENBQUN1RixNQUFSLEVBQW5CLENBQUE7RUFDRCxLQUFBOztFQUVEcUIsSUFBQUEsSUFBSSxDQUFDQyxtQkFBTCxHQUEyQjdHLE9BQU8sQ0FBQ2lCLGNBQVIsQ0FBdUJxRSxXQUFsRCxDQUFBOztNQUVBLElBQUlzQixJQUFJLENBQUNyRixLQUFULEVBQWdCO0VBQ2Q7RUFDQXFGLE1BQUFBLElBQUksQ0FBQ3JGLEtBQUwsQ0FBV3VGLE9BQVgsQ0FBbUIsVUFBQzlFLElBQUQsRUFBZ0I7RUFDakNBLFFBQUFBLElBQUksQ0FBQytFLGdCQUFMLEdBQXdCSCxJQUFJLENBQUN0QixXQUE3QixDQUFBOztVQUNBLElBQUl0RCxJQUFJLENBQUNnRixLQUFMLElBQWNoRixJQUFJLENBQUNnRixLQUFMLEdBQWEsQ0FBL0IsRUFBa0M7RUFDaENoRixVQUFBQSxJQUFJLENBQUNnRixLQUFMLEdBQWFyRSxJQUFJLENBQUNDLEtBQUwsQ0FBV1osSUFBSSxDQUFDZ0YsS0FBTCxHQUFhLEdBQXhCLENBQWIsQ0FBQTtFQUNELFNBQUE7U0FKSCxDQUFBLENBQUE7RUFNRCxLQS9Cb0I7RUFrQ3JCO0VBQ0E7RUFDQTtFQUVBOzs7RUFDQSxJQUFBLElBQUksQ0FBQ0osSUFBSSxDQUFDSyxJQUFWLEVBQWdCO1FBQ2RMLElBQUksQ0FBQ0ssSUFBTCxHQUFZakgsT0FBTyxDQUFDa0gsU0FBUixDQUFrQk4sSUFBbEIsQ0FBWixDQUFBO0VBQ0QsS0F6Q29COzs7TUE0Q3JCLElBQUk1RyxPQUFPLENBQUNxQixXQUFaLEVBQXlCO0VBQ3ZCckIsTUFBQUEsT0FBTyxDQUFDK0MsR0FBUixDQUFZLE1BQVosRUFBb0Isc0JBQXBCLEVBQTBDYixRQUFBLENBQUEsRUFBQSxFQUFPbEMsT0FBTyxDQUFDcUIsV0FBZixDQUExQyxFQUR1Qjs7UUFJdkIsSUFBSXJCLE9BQU8sQ0FBQ3FCLFdBQVIsQ0FBb0I0RixJQUFwQixLQUE2QkwsSUFBSSxDQUFDSyxJQUF0QyxFQUE0QztFQUMxQyxRQUFBLE9BQUE7RUFDRCxPQUFBO0VBQ0YsS0FBQTs7RUFFRCxJQUFBLElBQU1oRSxHQUFHLEdBQUcsSUFBSUosSUFBSixFQUFBLENBQVdLLFdBQVgsRUFBWixDQUFBOztFQUVBLElBQUEsSUFBSSxDQUFDMEQsSUFBSSxDQUFDeEQsVUFBVixFQUFzQjtFQUNwQndELE1BQUFBLElBQUksQ0FBQ3hELFVBQUwsR0FBa0IsSUFBSVAsSUFBSixFQUFBLENBQVdLLFdBQVgsRUFBbEIsQ0FBQTtFQUNELEtBQUE7O0VBRURsRCxJQUFBQSxPQUFPLENBQUNxQixXQUFSLEdBQXNCdUYsSUFBdEIsQ0EzRHFCO0VBOERyQjs7RUFDQTVHLElBQUFBLE9BQU8sQ0FBQ2MsV0FBUixDQUFvQnFGLG1CQUFwQixHQUEwQ2xELEdBQTFDLENBQUE7O0VBRUEsSUFBQSxJQUFJLENBQUNqRCxPQUFPLENBQUNpQixjQUFSLENBQXVCcUYsa0JBQTVCLEVBQWdEO0VBQzlDdEcsTUFBQUEsT0FBTyxDQUFDaUIsY0FBUixDQUF1QnFGLGtCQUF2QixHQUE0QyxDQUE1QyxDQUFBO0VBQ0QsS0FBQTs7TUFDRHRHLE9BQU8sQ0FBQ2lCLGNBQVIsQ0FBdUJxRixrQkFBdkIsRUFBQSxDQUFBO0VBQ0F0RyxJQUFBQSxPQUFPLENBQUNpQixjQUFSLENBQXVCbUMsVUFBdkIsR0FBb0NILEdBQXBDLENBQUE7RUFDQWpELElBQUFBLE9BQU8sQ0FBQ3FELGlCQUFSLENBQTBCckQsT0FBTyxDQUFDaUIsY0FBbEMsRUF0RXFCOztFQXdFckJqQixJQUFBQSxPQUFPLENBQUNzQixVQUFSLENBQW1CRSxHQUFuQixDQUF1QixNQUF2QixFQUErQm9GLElBQS9CLENBQUEsQ0FBQTtLQTNZc0I7SUE4WXhCTyxVQUFVLEVBQUUsU0FBQ3pGLFVBQUFBLENBQUFBLElBQUQsRUFBYTtFQUN2QixJQUFBLElBQUksQ0FBQzFCLE9BQU8sQ0FBQ1ksT0FBYixFQUFzQjtFQUNwQlosTUFBQUEsT0FBTyxDQUFDK0MsR0FBUixDQUFZLE9BQVosRUFBcUIsMERBQXJCLENBQUEsQ0FBQTs7UUFDQS9DLE9BQU8sQ0FBQ2lGLGNBQVIsQ0FBdUIsWUFBQTtFQUFNLFFBQUEsT0FBQWpGLE9BQU8sQ0FBQ21ILFVBQVIsQ0FBbUJ6RixJQUFuQixDQUFBLENBQUE7U0FBN0IsQ0FBQSxDQUFBOztFQUNBLE1BQUEsT0FBQTtFQUNELEtBTHNCOzs7RUFRdkIsSUFBQSxJQUFJLENBQUNBLElBQUQsSUFBUyxRQUFPQSxJQUFQLENBQUEsS0FBZ0IsUUFBN0IsRUFBdUM7RUFDckMxQixNQUFBQSxPQUFPLENBQUMrQyxHQUFSLENBQVksT0FBWixFQUFxQixvQkFBckIsQ0FBQSxDQUFBO0VBQ0EsTUFBQSxPQUFBO0VBQ0QsS0FYc0I7OztFQWN2QixJQUFBLElBQU1xRSxLQUFLLEdBQUc3RSxJQUFJLENBQUM4QyxLQUFMLENBQVc5QyxJQUFJLENBQUNDLFNBQUwsQ0FBZWQsSUFBZixDQUFYLENBQWQsQ0FkdUI7O01BaUJ2QixJQUFJMEYsS0FBSyxDQUFDN0YsS0FBVixFQUFpQjtFQUNmNkYsTUFBQUEsS0FBSyxDQUFDN0YsS0FBTixDQUFZdUYsT0FBWixDQUFvQixVQUFDOUUsSUFBRCxFQUFpQjtVQUNuQyxJQUFJQSxJQUFJLENBQUNnRixLQUFMLElBQWNoRixJQUFJLENBQUNnRixLQUFMLEdBQWEsQ0FBL0IsRUFBa0M7RUFDaENoRixVQUFBQSxJQUFJLENBQUNnRixLQUFMLEdBQWFyRSxJQUFJLENBQUNDLEtBQUwsQ0FBV1osSUFBSSxDQUFDZ0YsS0FBTCxHQUFhLEdBQXhCLENBQWIsQ0FBQTtFQUNELFNBQUE7O0VBQ0RoRixRQUFBQSxJQUFJLENBQUNxRixpQkFBTCxHQUF5QkQsS0FBSyxDQUFDOUIsV0FBL0IsQ0FBQTtTQUpGLENBQUEsQ0FBQTtFQU1ELEtBeEJzQjtFQTJCdkI7RUFDQTtFQUNBOzs7TUFFQSxJQUFNckMsR0FBRyxHQUFHLElBQUlKLElBQUosR0FBV0ssV0FBWCxFQUFaLENBL0J1Qjs7RUFrQ3ZCbEQsSUFBQUEsT0FBTyxDQUFDYyxXQUFSLENBQW9CcUYsbUJBQXBCLEdBQTBDbEQsR0FBMUMsQ0FBQTtFQUNBakQsSUFBQUEsT0FBTyxDQUFDb0csY0FBUixDQUF1QnBHLE9BQU8sQ0FBQ2MsV0FBL0IsRUFuQ3VCO0VBcUN2Qjs7RUFDQSxJQUFBLElBQUksQ0FBQ2QsT0FBTyxDQUFDaUIsY0FBUixDQUF1QnFGLGtCQUE1QixFQUFnRHRHLE9BQU8sQ0FBQ2lCLGNBQVIsQ0FBdUJxRixrQkFBdkIsR0FBNEMsQ0FBNUMsQ0FBQTtNQUNoRHRHLE9BQU8sQ0FBQ2lCLGNBQVIsQ0FBdUJxRixrQkFBdkIsRUFBQSxDQUFBO0VBQ0F0RyxJQUFBQSxPQUFPLENBQUNpQixjQUFSLENBQXVCbUMsVUFBdkIsR0FBb0NILEdBQXBDLENBQUE7RUFDQWpELElBQUFBLE9BQU8sQ0FBQ3FELGlCQUFSLENBQTBCckQsT0FBTyxDQUFDaUIsY0FBbEMsRUF6Q3VCO0VBMkN2Qjs7RUFDQWpCLElBQUFBLE9BQU8sQ0FBQ3NCLFVBQVIsQ0FBbUJFLEdBQW5CLENBQXVCLE9BQXZCLEVBQWdDNEYsS0FBaEMsQ0FBQSxDQUFBO0tBMWJzQjtJQTZieEJFLGdCQUFnQixFQUFFLFNBQUM1RixnQkFBQUEsQ0FBQUEsSUFBRCxFQUFjO0VBQzlCLElBQUEsSUFBSSxDQUFDMUIsT0FBTyxDQUFDWSxPQUFiLEVBQXNCO0VBQ3BCWixNQUFBQSxPQUFPLENBQUMrQyxHQUFSLENBQVksT0FBWixFQUFxQixnRUFBckIsQ0FBQSxDQUFBOztRQUNBL0MsT0FBTyxDQUFDaUYsY0FBUixDQUF1QixZQUFBO0VBQU0sUUFBQSxPQUFBakYsT0FBTyxDQUFDc0gsZ0JBQVIsQ0FBeUI1RixJQUF6QixDQUFBLENBQUE7U0FBN0IsQ0FBQSxDQUFBOztFQUNBLE1BQUEsT0FBQTtFQUNELEtBTDZCOzs7RUFROUIsSUFBQSxJQUFNNkYsU0FBUyx5QkFBUXZILE9BQU8sQ0FBQ2dCLGdCQUFrQlUsS0FBakQsQ0FSOEI7OztNQVU5QjZGLFNBQVMsQ0FBQ25FLFVBQVYsR0FBdUIsSUFBSVAsSUFBSixFQUFXSyxDQUFBQSxXQUFYLEVBQXZCLENBVjhCO0VBYTlCO0VBQ0E7RUFDQTtFQUNBOztFQUVBbEQsSUFBQUEsT0FBTyxDQUFDZ0IsYUFBUixHQUF3QnVHLFNBQXhCLENBbEI4Qjs7TUFvQjlCdkgsT0FBTyxDQUFDd0gsU0FBUixDQUNFeEgsT0FBTyxDQUFDQyxNQUFSLENBQWVJLFNBQWYsR0FBMkIsUUFEN0IsRUFFRWtDLElBQUksQ0FBQ0MsU0FBTCxDQUFleEMsT0FBTyxDQUFDZ0IsYUFBdkIsQ0FGRixFQUdFdEIsZ0JBSEYsQ0FBQSxDQUFBO0tBamRzQjtJQXdkeEIyRCxpQkFBaUIsRUFBRSxTQUFDM0IsaUJBQUFBLENBQUFBLElBQUQsRUFBZTtFQUNoQyxJQUFBLElBQUksQ0FBQzFCLE9BQU8sQ0FBQ1ksT0FBYixFQUFzQjtFQUNwQlosTUFBQUEsT0FBTyxDQUFDK0MsR0FBUixDQUFZLE9BQVosRUFBcUIsaUVBQXJCLENBQUEsQ0FBQTs7UUFDQS9DLE9BQU8sQ0FBQ2lGLGNBQVIsQ0FBdUIsWUFBQTtFQUFNLFFBQUEsT0FBQWpGLE9BQU8sQ0FBQ3FELGlCQUFSLENBQTBCM0IsSUFBMUIsQ0FBQSxDQUFBO1NBQTdCLENBQUEsQ0FBQTs7RUFDQSxNQUFBLE9BQUE7RUFDRCxLQUwrQjs7O0VBUWhDLElBQUEsSUFBTStGLFVBQVUseUJBQVF6SCxPQUFPLENBQUNpQixpQkFBbUJTLEtBQW5ELENBUmdDOzs7RUFVaEMrRixJQUFBQSxVQUFVLENBQUNyRSxVQUFYLEdBQXdCLElBQUlQLElBQUosRUFBQSxDQUFXSyxXQUFYLEVBQXhCLENBQUE7TUFFQWxELE9BQU8sQ0FBQ2lCLGNBQVIsR0FBeUJ3RyxVQUF6QixDQUFBO01BRUF6SCxPQUFPLENBQUMrQyxHQUFSLENBQVksTUFBWixFQUFvQix3QkFBcEIsRUFBOEMwRSxVQUE5QyxDQUFBLENBZGdDOztNQWlCaEN6SCxPQUFPLENBQUN3SCxTQUFSLENBQ0V4SCxPQUFPLENBQUNDLE1BQVIsQ0FBZUksU0FBZixHQUEyQixTQUQ3QixFQUVFa0MsSUFBSSxDQUFDQyxTQUFMLENBQWV4QyxPQUFPLENBQUNpQixjQUF2QixDQUZGLEVBR0VqQixPQUFPLENBQUNDLE1BQVIsQ0FBZUcsZUFIakIsQ0FBQSxDQUFBO0tBemVzQjtFQWdmeEI7RUFDQTtFQUNBO0VBQ0E7SUFDQWdHLGNBQWMsRUFBRSxTQUFDMUUsY0FBQUEsQ0FBQUEsSUFBRCxFQUFZO0VBQzFCLElBQUEsSUFBSSxDQUFDMUIsT0FBTyxDQUFDWSxPQUFiLEVBQXNCO0VBQ3BCWixNQUFBQSxPQUFPLENBQUMrQyxHQUFSLENBQVksT0FBWixFQUFxQiw4REFBckIsQ0FBQSxDQUFBOztRQUNBL0MsT0FBTyxDQUFDaUYsY0FBUixDQUF1QixZQUFBO0VBQU0sUUFBQSxPQUFBakYsT0FBTyxDQUFDb0csY0FBUixDQUF1QjFFLElBQXZCLENBQUEsQ0FBQTtTQUE3QixDQUFBLENBQUE7O0VBQ0EsTUFBQSxPQUFBO0VBQ0QsS0FMeUI7OztFQVExQixJQUFBLElBQUlBLElBQUksSUFBSUEsSUFBSSxDQUFDZ0csb0JBQUwsS0FBOEIzRyxTQUExQyxFQUFxRDtFQUNuRFcsTUFBQUEsSUFBSSxDQUFDaUcsV0FBTCxHQUFtQmpHLElBQUksQ0FBQ2dHLG9CQUF4QixDQUFBO1FBQ0EsT0FBT2hHLElBQUksQ0FBQ2dHLG9CQUFaLENBQUE7RUFDRCxLQUFBOztNQUVELElBQ0VoRyxJQUFJLElBQ0pBLElBQUksQ0FBQzRELFdBREwsSUFFQTVELElBQUksQ0FBQzRELFdBQUwsS0FBcUIsRUFGckIsSUFHQTVELElBQUksQ0FBQzRELFdBQUwsS0FBcUJ0RixPQUFPLENBQUNjLFdBQVIsQ0FBb0J3RSxXQUozQyxFQUtFO0VBQ0E7UUFDQSxJQUFJdEYsT0FBTyxDQUFDYyxXQUFSLENBQW9COEcsZ0JBQXBCLElBQXdDbEcsSUFBSSxDQUFDa0csZ0JBQUwsS0FBMEIsSUFBdEUsRUFBNEU7VUFDMUU1SCxPQUFPLENBQUMrQyxHQUFSLENBQVksTUFBWixFQUFvQixnREFBcEIsRUFBc0VyQixJQUF0RSxDQUFBLENBRDBFOztVQUcxRTFCLE9BQU8sQ0FBQ2dCLGFBQVIsR0FBd0JELFNBQXhCLENBQUE7O0VBQ0FmLFFBQUFBLE9BQU8sQ0FBQ3FDLGFBQVIsQ0FBc0J3RixNQUF0QixDQUE2QixRQUE3QixDQUFBLENBQUE7O1VBQ0E3SCxPQUFPLENBQUM4SCxhQUFSLEVBQUEsQ0FMMEU7OztVQVExRTlILE9BQU8sQ0FBQ2lCLGNBQVIsR0FBeUJGLFNBQXpCLENBQUE7VUFDQWYsT0FBTyxDQUFDK0gsWUFBUixDQUFxQi9ILE9BQU8sQ0FBQ0MsTUFBUixDQUFlSSxTQUFmLEdBQTJCLFNBQWhELENBQUEsQ0FBQTs7VUFDQUwsT0FBTyxDQUFDZ0ksY0FBUixFQUFBLENBVjBFOzs7VUFhMUVDLE1BQU0sQ0FBQ0MsSUFBUCxDQUFZeEcsSUFBWixFQUFrQm9GLE9BQWxCLENBQTBCLFVBQUNxQixHQUFELEVBQUk7RUFDNUIsVUFBQSxJQUFJLE9BQU96RyxJQUFJLENBQUN5RyxHQUFELENBQVgsS0FBcUIsUUFBckIsSUFBaUN6RyxJQUFJLENBQUN5RyxHQUFELENBQUosS0FBYyxFQUFuRCxFQUF1RDtjQUNyRCxPQUFPekcsSUFBSSxDQUFDeUcsR0FBRCxDQUFYLENBQUE7RUFDRCxXQUFBO1dBSEgsQ0FBQSxDQUFBO1VBTUFuSSxPQUFPLENBQUNjLFdBQVIsR0FBMkJvQixRQUFBLENBQUEsRUFBQSxFQUFBUixJQUFBLENBQTNCLENBbkIwRTs7RUFzQjFFLFFBQUEsSUFBSTFCLE9BQU8sQ0FBQ2MsV0FBUixDQUFvQjBFLFVBQXBCLEtBQW1DekUsU0FBdkMsRUFBa0Q7WUFDaERmLE9BQU8sQ0FBQ2MsV0FBUixDQUFvQjBFLFVBQXBCLEdBQWlDLElBQUkzQyxJQUFKLEVBQVdLLENBQUFBLFdBQVgsRUFBakMsQ0FBQTtFQUNELFNBeEJ5RTtFQTJCMUU7RUFDQTtFQUNBOzs7RUFFQSxRQUFBLE9BQUE7RUFDRCxPQWxDRDtFQXFDQTs7O0VBQ0EsTUFBQSxJQUFJLENBQUNsRCxPQUFPLENBQUNjLFdBQVIsQ0FBb0I4RyxnQkFBekIsRUFBMkM7RUFDekM1SCxRQUFBQSxPQUFPLENBQUMrQyxHQUFSLENBQ0UsTUFERixFQUVFLHFCQUZGLEVBR0UvQyxPQUFPLENBQUNjLFdBQVIsQ0FBb0J3RSxXQUh0QixFQUlFLFdBSkYsRUFLRTVELElBTEYsQ0FBQSxDQUFBO0VBUUExQixRQUFBQSxPQUFPLENBQUNzQixVQUFSLENBQW1CRSxHQUFuQixDQUF1QixZQUF2QixFQUFxQztFQUNuQzRHLFVBQUFBLHFCQUFxQixFQUFFcEksT0FBTyxDQUFDYyxXQUFSLENBQW9Cd0UsV0FEUjtZQUVuQytDLG1CQUFtQixFQUFFM0csSUFBSSxDQUFDNEQsV0FGUztFQUduQ2dELFVBQUFBLHdCQUF3QixFQUFFNUcsSUFBSSxDQUFDa0csZ0JBQUwsS0FBMEIsSUFIakI7RUFJbkNXLFVBQUFBLGtCQUFrQixFQUFFN0csSUFBSSxDQUFDOEQsVUFBTCxHQUFrQjlELElBQUksQ0FBQzhELFVBQXZCLEdBQW9DLElBQUkzQyxJQUFKLEVBQUEsQ0FBV0ssV0FBWCxFQUFBO1dBSjFELENBQUEsQ0FBQTtFQU1ELE9BQUE7RUFDRixLQXhFeUI7OztFQTJFMUIsSUFBQSxJQUFNc0YsT0FBTyx5QkFBUXhJLE9BQU8sQ0FBQ2MsY0FBZ0JZLEtBQTdDLENBM0UwQjtFQThFMUI7RUFDQTtFQUNBO0VBQ0E7RUFDQTs7O0VBRUExQixJQUFBQSxPQUFPLENBQUNjLFdBQVIsR0FBc0IwSCxPQUF0QixDQXBGMEI7O01BdUYxQnhJLE9BQU8sQ0FBQ3dILFNBQVIsQ0FDRXhILE9BQU8sQ0FBQ0MsTUFBUixDQUFlSSxTQUFmLEdBQTJCLE1BRDdCLEVBRUVrQyxJQUFJLENBQUNDLFNBQUwsQ0FBZXhDLE9BQU8sQ0FBQ2MsV0FBdkIsQ0FGRixFQUdFcEIsZ0JBSEYsQ0FBQSxDQUFBO0tBM2tCc0I7RUFrbEJ4QjtFQUNBK0ksRUFBQUEsZUFBZSxFQUFFLFNBQUEsZUFBQSxHQUFBO0VBQ2YsSUFBQSxJQUFJLENBQUN6SSxPQUFPLENBQUNZLE9BQWIsRUFBc0I7RUFDcEJaLE1BQUFBLE9BQU8sQ0FBQytDLEdBQVIsQ0FBWSxPQUFaLEVBQXFCLCtEQUFyQixDQUFBLENBQUE7O1FBQ0EvQyxPQUFPLENBQUNpRixjQUFSLENBQXVCLFlBQUE7VUFBTSxPQUFBakYsT0FBTyxDQUFDeUksZUFBUixFQUFBLENBQUE7U0FBN0IsQ0FBQSxDQUFBOztFQUNBLE1BQUEsT0FBQTtFQUNELEtBQUE7O0VBRUR6SSxJQUFBQSxPQUFPLENBQUNzQixVQUFSLENBQW1CRSxHQUFuQixDQUF1QixNQUF2QixFQUE2QlUsUUFBQSxDQUFBLEVBQUEsRUFBT2xDLE9BQU8sQ0FBQ2MsV0FBZixDQUE3QixDQUFBLENBQUE7S0ExbEJzQjtFQTZsQnhCb0YsRUFBQUEsYUFBYSxFQUFFLFNBQUEsYUFBQSxHQUFBO0VBQ2IsSUFBQSxPQUFPd0MsQ0FBUyxDQUFDQyxLQUFWLEtBQW9CaEosVUFBVSxDQUFDQyxNQUF0QyxDQUFBO0tBOWxCc0I7RUFpbUJ4QjtJQUNBZ0osUUFBUSxFQUFFLFNBQUNDLFFBQUFBLENBQUFBLFNBQUQsRUFBbUI7RUFDM0I7RUFDQSxJQUFBLElBQUk3SSxPQUFPLENBQUNZLE9BQVIsS0FBb0IsS0FBeEIsRUFBK0I7RUFDN0JaLE1BQUFBLE9BQU8sQ0FBQytDLEdBQVIsQ0FBWSxNQUFaLEVBQW9CLDZDQUFwQixDQUFBLENBQUE7UUFFQThDLE1BQU0sQ0FBQ2lELFVBQVAsQ0FBa0IsWUFBQTtVQUNoQjlJLE9BQU8sQ0FBQzRJLFFBQVIsQ0FBaUJDLFNBQWpCLENBQUEsQ0FBQTtFQUNELE9BRkQsRUFFRyxFQUZILENBQUEsQ0FBQTtFQUdBLE1BQUEsT0FBQTtFQUNELEtBVDBCOzs7RUFZM0IsSUFBQSxJQUFJLENBQUM3SSxPQUFPLENBQUNhLGVBQWIsRUFBOEI7RUFDNUJiLE1BQUFBLE9BQU8sQ0FBQytDLEdBQVIsQ0FBWSxNQUFaLEVBQW9CLHlDQUFwQixDQUFBLENBQUE7RUFDQSxNQUFBLE9BQUE7RUFDRCxLQWYwQjs7O0VBa0IzQixJQUFBLElBQUkvQyxPQUFPLENBQUNzQixVQUFSLENBQW1CQyxLQUFuQixDQUF5QnVDLE1BQXpCLEtBQW9DLENBQXBDLElBQXlDOUQsT0FBTyxDQUFDdUQsYUFBUixDQUFzQk8sTUFBdEIsS0FBaUMsQ0FBOUUsRUFBaUY7RUFDL0U5RCxNQUFBQSxPQUFPLENBQUMrQyxHQUFSLENBQVksTUFBWixFQUFvQiwwQ0FBcEIsQ0FBQSxDQUFBO0VBQ0EsTUFBQSxPQUFBO0VBQ0QsS0FBQTs7TUFFRCxJQUFNZ0csU0FBUyxHQUFRN0csUUFBQSxDQUFBLEVBQUEsRUFBQWxDLE9BQU8sQ0FBQ2dCLGFBQVIsQ0FBdkIsQ0F2QjJCOzs7RUF5QjNCLElBQUEsSUFBSStILFNBQVMsQ0FBQ0MsVUFBVixLQUF5QkMsU0FBUyxDQUFDQyxTQUF2QyxFQUFrRDtFQUNoREgsTUFBQUEsU0FBUyxDQUFDM0YsVUFBVixHQUF1QixJQUFJUCxJQUFKLEVBQUEsQ0FBV0ssV0FBWCxFQUF2QixDQUFBO0VBQ0QsS0FBQTs7RUFDRDZGLElBQUFBLFNBQVMsQ0FBQ0MsVUFBVixHQUF1QkMsU0FBUyxDQUFDQyxTQUFqQyxDQUFBO0VBQ0FILElBQUFBLFNBQVMsQ0FBQ0ksUUFBVixHQUFxQkYsU0FBUyxDQUFDRSxRQUEvQixDQUFBO0VBQ0FKLElBQUFBLFNBQVMsQ0FBQ0ssVUFBVixHQUF1QnBKLE9BQU8sQ0FBQ3FKLFlBQVIsRUFBdkIsQ0FBQTtFQUNBTixJQUFBQSxTQUFTLENBQUNPLFVBQVYsR0FDRXpELE1BQU0sQ0FBQzBELE1BQVAsSUFBaUIxRCxNQUFNLENBQUMwRCxNQUFQLENBQWNDLEtBQS9CLElBQXdDM0QsTUFBTSxDQUFDMEQsTUFBUCxDQUFjRSxNQUF0RCxHQUNJNUQsTUFBTSxDQUFDMEQsTUFBUCxDQUFjRSxNQUFkLEdBQXVCNUQsTUFBTSxDQUFDMEQsTUFBUCxDQUFjQyxLQUFyQyxHQUNFM0QsTUFBTSxDQUFDMEQsTUFBUCxDQUFjRSxNQUFkLEdBQXVCLEdBQXZCLEdBQTZCNUQsTUFBTSxDQUFDMEQsTUFBUCxDQUFjQyxLQUQ3QyxHQUVFM0QsTUFBTSxDQUFDMEQsTUFBUCxDQUFjQyxLQUFkLEdBQXNCLEdBQXRCLEdBQTRCM0QsTUFBTSxDQUFDMEQsTUFBUCxDQUFjRSxNQUhoRCxHQUlJMUksU0FMTixDQS9CMkI7O01BdUMzQixJQUFNMkksT0FBTyxHQUFHLEVBQWhCLENBQUE7TUFDQSxJQUFJQyxVQUFVLEdBQVksRUFBMUIsQ0FBQTs7TUFFQSxPQUFPM0osT0FBTyxDQUFDc0IsVUFBUixDQUFtQkMsS0FBbkIsQ0FBeUJ1QyxNQUF6QixHQUFrQyxDQUF6QyxFQUE0QztRQUMxQzZGLFVBQVUsQ0FBQ3ZILElBQVgsQ0FBZ0JwQyxPQUFPLENBQUNzQixVQUFSLENBQW1CQyxLQUFuQixDQUF5QnFJLEtBQXpCLEVBQWhCLENBQUEsQ0FBQTs7RUFDQSxNQUFBLElBQUlELFVBQVUsQ0FBQzdGLE1BQVgsSUFBcUIsRUFBekIsRUFBNkI7VUFDM0I0RixPQUFPLENBQUN0SCxJQUFSLENBQWF1SCxVQUFiLENBQUEsQ0FBQTtFQUNBQSxRQUFBQSxVQUFVLEdBQUcsRUFBYixDQUFBO0VBQ0QsT0FBQTtFQUNGLEtBaEQwQjs7O0VBbUQzQixJQUFBLElBQUlBLFVBQVUsQ0FBQzdGLE1BQVgsR0FBb0IsQ0FBeEIsRUFBMkI7UUFDekI0RixPQUFPLENBQUN0SCxJQUFSLENBQWF1SCxVQUFiLENBQUEsQ0FBQTtFQUNELEtBckQwQjs7O0VBd0QzQkQsSUFBQUEsT0FBTyxDQUFDNUMsT0FBUixDQUFnQixVQUFDK0MsS0FBRCxFQUFlO0VBQzdCO0VBQ0FBLE1BQUFBLEtBQUssQ0FBQy9DLE9BQU4sQ0FBYyxVQUFDOUUsSUFBRCxFQUFZO1VBQ3hCQSxJQUFJLENBQUNzQixNQUFMLEdBQWN5RixTQUFkLENBQUE7U0FERixDQUFBLENBQUE7RUFJQS9JLE1BQUFBLE9BQU8sQ0FBQ3VELGFBQVIsQ0FBc0JuQixJQUF0QixDQUEyQjtFQUN6QjBILFFBQUFBLEVBQUUsRUFBRTlKLE9BQU8sQ0FBQ3VGLE1BQVIsRUFEcUI7RUFFekJyRixRQUFBQSxZQUFZLEVBQUVGLE9BQU8sQ0FBQ0MsTUFBUixDQUFlQyxZQUZKO0VBR3pCcUIsUUFBQUEsS0FBSyxFQUFFc0ksS0FIa0I7RUFJekJFLFFBQUFBLE9BQU8sRUFBRTtFQUVQO1dBTnVCO0VBUXpCdkUsUUFBQUEsVUFBVSxFQUFFLElBQUkzQyxJQUFKLEVBQUEsQ0FBV0ssV0FBWCxFQUFBO1NBUmQsQ0FBQSxDQUFBO0VBVUQsS0FoQkQsRUF4RDJCOztFQTJFM0IsSUFBQSxJQUFJbEQsT0FBTyxDQUFDdUQsYUFBUixDQUFzQk8sTUFBdEIsS0FBaUMsQ0FBckMsRUFBd0M7RUFDdEM5RCxNQUFBQSxPQUFPLENBQUMrQyxHQUFSLENBQVksTUFBWixFQUFvQixpREFBcEIsQ0FBQSxDQUFBO0VBQ0EsTUFBQSxPQUFBO0VBQ0QsS0E5RTBCOzs7RUFpRjNCL0MsSUFBQUEsT0FBTyxDQUFDcUMsYUFBUixDQUFzQkMsR0FBdEIsQ0FBMEIsZUFBMUIsRUFBMkNDLElBQUksQ0FBQ0MsU0FBTCxDQUFleEMsT0FBTyxDQUFDdUQsYUFBdkIsQ0FBM0MsRUFqRjJCOzs7RUFvRjNCdkQsSUFBQUEsT0FBTyxDQUFDK0MsR0FBUixDQUFZLE1BQVosRUFBb0IsK0JBQXBCLENBQUEsQ0FBQTs7TUFFQS9DLE9BQU8sQ0FBQ2dLLGlCQUFSLENBQTBCbkIsU0FBMUIsQ0FBQSxDQUFBO0tBeHJCc0I7SUEyckJ4Qm1CLGlCQUFpQixFQUFFLFNBQUNuQixpQkFBQUEsQ0FBQUEsU0FBRCxFQUFtQjtFQUNwQztNQUNBLElBQUk3SSxPQUFPLENBQUN3RCxhQUFaLEVBQTJCO0VBQ3pCeEQsTUFBQUEsT0FBTyxDQUFDK0MsR0FBUixDQUFZLE1BQVosRUFBb0IseUNBQXBCLENBQUEsQ0FBQTtFQUNBLE1BQUEsT0FBQTtFQUNELEtBTG1DOzs7RUFRcEMvQyxJQUFBQSxPQUFPLENBQUN3RCxhQUFSLEdBQXdCLElBQXhCLENBUm9DOztNQVdwQ3hELE9BQU8sQ0FBQ3VELGFBQVIsQ0FBc0IwRyxJQUF0QixDQUEyQixVQUFDQyxDQUFELEVBQUlDLENBQUosRUFBSztFQUM5QixNQUFBLElBQUlELENBQUMsQ0FBQzFFLFVBQUYsR0FBZTJFLENBQUMsQ0FBQzNFLFVBQXJCLEVBQWlDO1VBQy9CLE9BQU8sQ0FBQyxDQUFSLENBRCtCO0VBRWhDLE9BQUE7O0VBQ0QsTUFBQSxJQUFJMEUsQ0FBQyxDQUFDMUUsVUFBRixHQUFlMkUsQ0FBQyxDQUFDM0UsVUFBckIsRUFBaUM7RUFDL0IsUUFBQSxPQUFPLENBQVAsQ0FBQTtFQUNELE9BQUE7O0VBQ0QsTUFBQSxPQUFPLENBQVAsQ0FBQTtPQVBGLENBQUEsQ0FBQTtNQVVBLElBQU00RSxZQUFZLEdBQUdwSyxPQUFPLENBQUN1RCxhQUFSLENBQXNCLENBQXRCLENBQXJCLENBckJvQzs7RUF3QnBDdkQsSUFBQUEsT0FBTyxDQUFDcUssWUFBUixDQUFxQkQsWUFBckIsRUFBbUMsQ0FBbkMsRUFBc0N2QixTQUF0QyxDQUFBLENBQUE7S0FudEJzQjtFQXN0QnhCd0IsRUFBQUEsWUFBWSxFQUFFLFNBQUNDLFlBQUFBLENBQUFBLFVBQUQsRUFBMEJDLFVBQTFCLEVBQThDMUIsU0FBOUMsRUFBZ0U7TUFDNUU3SSxPQUFPLENBQUN3SyxLQUFSLENBQWNGLFVBQWQsRUFBMEJ6QixTQUExQixFQUFxQyxVQUFDeEUsS0FBRCxFQUFNO1FBQ3pDLElBQUlvRyxPQUFPLEdBQUcsSUFBZCxDQUFBOztFQUVBLE1BQUEsSUFBSXBHLEtBQUosRUFBVztFQUNUO1VBQ0FyRSxPQUFPLENBQUMrQyxHQUFSLENBQVksT0FBWixFQUFxQix3QkFBckIsRUFBK0NzQixLQUEvQyxDQUFBLENBRlM7O1VBS1QsSUFBSTtZQUNGLElBQUlxRyxTQUFTLEdBQUduSSxJQUFJLENBQUM4QyxLQUFMLENBQVdoQixLQUFYLENBQUEsSUFBcUIsRUFBckMsQ0FERTs7RUFHRixVQUFBLElBQUlxRyxTQUFTLENBQUNDLElBQVYsR0FBaUIsR0FBckIsRUFBMEI7RUFDeEJGLFlBQUFBLE9BQU8sR0FBRyxLQUFWLENBQUE7RUFDRCxXQUFBO1dBTEgsQ0FNRSxPQUFPRyxHQUFQLEVBQVk7RUFDWkgsVUFBQUEsT0FBTyxHQUFHLEtBQVYsQ0FBQTtFQUNELFNBQUE7O1VBRUR6SyxPQUFPLENBQUMrQyxHQUFSLENBQVksTUFBWixFQUFvQixRQUFwQixFQUE4QjBILE9BQTlCLEVBQXVDLGFBQXZDLEVBQXNERixVQUF0RCxDQUFBLENBQUE7O1VBRUEsSUFBSUUsT0FBTyxLQUFLLEtBQWhCLEVBQXVCO0VBQ3JCO0VBQ0EsVUFBQSxJQUFJRixVQUFVLElBQUl2SyxPQUFPLENBQUNDLE1BQVIsQ0FBZVMsU0FBakMsRUFBNEM7Y0FDMUNWLE9BQU8sQ0FBQ3dELGFBQVIsR0FBd0IsS0FBeEIsQ0FBQTtFQUNBeEQsWUFBQUEsT0FBTyxDQUFDK0MsR0FBUixDQUFZLE1BQVosRUFBb0IsNkJBQXBCLENBQUEsQ0FBQTtFQUNBLFlBQUEsT0FBQTtFQUNELFdBTm9COzs7RUFTckIsVUFBQSxJQUFJOEgsS0FBSyxHQUFHLEdBQVosQ0FUcUI7O1lBVXJCLElBQUlDLE1BQU0sR0FBR0QsS0FBYixDQUFBOztZQUVBLElBQUlOLFVBQVUsR0FBRyxDQUFqQixFQUFvQjtjQUNsQixLQUFLLElBQUlRLENBQUMsR0FBRyxDQUFiLEVBQWdCQSxDQUFDLEdBQUdSLFVBQXBCLEVBQWdDUSxDQUFDLEVBQWpDLEVBQXFDO2dCQUNuQ0QsTUFBTSxHQUFHQSxNQUFNLEdBQUcsQ0FBbEIsQ0FBQTtFQUNELGFBQUE7RUFDRixXQUFBOztZQUVEOUssT0FBTyxDQUFDK0MsR0FBUixDQUFZLE9BQVosRUFBcUIsY0FBYytILE1BQWQsR0FBdUIsSUFBNUMsQ0FBQSxDQWxCcUI7O1lBcUJyQmpGLE1BQU0sQ0FBQ2lELFVBQVAsQ0FBa0IsWUFBQTtjQUNoQnlCLFVBQVUsRUFBQSxDQUFBOztFQUNWdkssWUFBQUEsT0FBTyxDQUFDcUssWUFBUixDQUFxQkMsVUFBckIsRUFBaUNDLFVBQWpDLEVBQTZDMUIsU0FBN0MsQ0FBQSxDQUFBO0VBQ0QsV0FIRCxFQUdHaUMsTUFISCxDQUFBLENBQUE7RUFLQSxVQUFBLE9BQUE7RUFDRCxTQUFBO0VBQ0YsT0FBQTs7UUFFRCxJQUFJTCxPQUFPLEtBQUssSUFBaEIsRUFBc0I7RUFDcEI7RUFFQSxRQUFBLElBQU1PLGtCQUFnQixHQUFHLEVBQXpCLENBSG9COztFQU1wQmhMLFFBQUFBLE9BQU8sQ0FBQ3VELGFBQVIsQ0FBc0J1RCxPQUF0QixDQUE4QixVQUFDbUUsRUFBRCxFQUFHO0VBQy9CLFVBQUEsSUFBSUEsRUFBRSxDQUFDbkIsRUFBSCxLQUFVUSxVQUFVLENBQUNSLEVBQXpCLEVBQTZCO2NBQzNCa0Isa0JBQWdCLENBQUM1SSxJQUFqQixDQUFzQjZJLEVBQXRCLENBQUEsQ0FBQTtFQUNELFdBQUE7V0FISCxDQUFBLENBQUE7RUFNQWpMLFFBQUFBLE9BQU8sQ0FBQ3VELGFBQVIsR0FBd0J5SCxrQkFBeEIsQ0Fab0I7O0VBZXBCaEwsUUFBQUEsT0FBTyxDQUFDcUMsYUFBUixDQUFzQkMsR0FBdEIsQ0FBMEIsZUFBMUIsRUFBMkNDLElBQUksQ0FBQ0MsU0FBTCxDQUFleEMsT0FBTyxDQUFDdUQsYUFBdkIsQ0FBM0MsRUFmb0I7OztVQWtCcEJ2RCxPQUFPLENBQUN3RCxhQUFSLEdBQXdCLEtBQXhCLENBQUE7O0VBRUEsUUFBQSxJQUFJeEQsT0FBTyxDQUFDdUQsYUFBUixDQUFzQk8sTUFBdEIsR0FBK0IsQ0FBbkMsRUFBc0M7WUFDcEM5RCxPQUFPLENBQUNnSyxpQkFBUixDQUEwQm5CLFNBQTFCLENBQUEsQ0FBQTtFQUNELFNBQUE7RUFDRixPQUFBO09BekVILENBQUEsQ0FBQTtLQXZ0QnNCO0VBb3lCeEIyQixFQUFBQSxLQUFLLEVBQUUsU0FBQ0YsS0FBQUEsQ0FBQUEsVUFBRCxFQUEwQnpCLFNBQTFCLEVBQThDN0QsUUFBOUMsRUFBK0U7RUFDcEY7TUFDQXNGLFVBQVUsQ0FBQ1AsT0FBWCxDQUFtQm1CLFlBQW5CLEdBQWtDLElBQUlySSxJQUFKLEVBQVdLLENBQUFBLFdBQVgsRUFBbEMsQ0FBQTtNQUNBLElBQU14QixJQUFJLEdBQUdhLElBQUksQ0FBQ0MsU0FBTCxDQUFlOEgsVUFBZixDQUFiLENBSG9GOztNQU1wRnRLLE9BQU8sQ0FBQytDLEdBQVIsQ0FBWSxNQUFaLEVBQW9CLGtCQUFwQixFQUF3Q3JCLElBQXhDLENBQUEsQ0FOb0Y7O0VBU3BGLElBQUEsSUFBSW1ILFNBQVMsSUFBSUksU0FBUyxDQUFDa0MsVUFBM0IsRUFBdUM7UUFDckMsSUFBTUMsTUFBTSxHQUFHbkMsU0FBUyxDQUFDa0MsVUFBVixDQUNibkwsT0FBTyxDQUFDQyxNQUFSLENBQWVFLElBQWYsR0FBc0IsT0FEVCxFQUViLElBQUlrTCxJQUFKLENBQVMsQ0FBQzNKLElBQUQsQ0FBVCxFQUFpQjtFQUFFNEosUUFBQUEsSUFBSSxFQUFFLGtCQUFBO0VBQVIsT0FBakIsQ0FGYSxDQUFmLENBQUE7RUFJQSxNQUFBLE9BQU90RyxRQUFRLENBQUNvRyxNQUFNLEdBQUcsSUFBSCxHQUFVLG1CQUFqQixDQUFmLENBQUE7RUFDRCxLQUFBOztFQUVELElBQUEsSUFBTUcsR0FBRyxHQUFtQixJQUFJQyxjQUFKLEVBQTVCLENBQUE7O01BRUFELEdBQUcsQ0FBQ0UsTUFBSixHQUFhLFlBQUE7RUFDWCxNQUFBLElBQU1DLElBQUksR0FBRyxVQUFjSCxJQUFBQSxHQUFkLEdBQW9CQSxHQUFHLENBQUNJLFFBQXhCLEdBQW1DSixHQUFHLENBQUMsY0FBRCxDQUFuRCxDQUFBOztFQUNBLE1BQUEsSUFBSUEsR0FBRyxDQUFDSyxNQUFKLElBQWMsR0FBbEIsRUFBdUI7VUFDckIsT0FBTzVHLFFBQVEsQ0FBQzBHLElBQUQsQ0FBZixDQUFBO0VBQ0QsT0FBQTs7UUFDRCxPQUFPMUcsUUFBUSxDQUFDLElBQUQsQ0FBZixDQUFBO09BTEYsQ0FBQTs7TUFRQXVHLEdBQUcsQ0FBQ00sT0FBSixHQUFjLFlBQUE7UUFDWixPQUFPN0csUUFBUSxDQUFDLHdCQUFELENBQWYsQ0FBQTtPQURGLENBQUE7O01BSUF1RyxHQUFHLENBQUNPLFNBQUosR0FBZ0IsWUFBQTtRQUNkLE9BQU85RyxRQUFRLENBQUMseUJBQUQsQ0FBZixDQUFBO09BREYsQ0FBQTs7RUFJQXVHLElBQUFBLEdBQUcsQ0FBQ1EsSUFBSixDQUFTLE1BQVQsRUFBaUIvTCxPQUFPLENBQUNDLE1BQVIsQ0FBZUUsSUFBZixHQUFzQixPQUF2QyxFQUFnRCxJQUFoRCxDQUFBLENBQUE7RUFDQW9MLElBQUFBLEdBQUcsQ0FBQ1MsZ0JBQUosQ0FBcUIsY0FBckIsRUFBcUMsa0JBQXJDLENBQUEsQ0FBQTtNQUNBVCxHQUFHLENBQUNVLGVBQUosR0FBc0IsSUFBdEIsQ0FBQTtNQUNBVixHQUFHLENBQUNXLElBQUosQ0FBU3hLLElBQVQsQ0FBQSxDQUFBO0tBMTBCc0I7RUE2MEJ4QjtFQUNBO0VBQ0E7RUFDQTtFQUNBeUssRUFBQUEsV0FBVyxFQUFFLFNBQUEsV0FBQSxHQUFBO0VBQ1g7RUFDQTtFQUNBLElBQUEsSUFBTUMsaUJBQWlCLEdBQUcsQ0FBQyxJQUFELEVBQU8sSUFBUCxFQUFhLElBQWIsRUFBbUIsSUFBbkIsRUFBeUIsSUFBekIsRUFBK0IsSUFBL0IsRUFBcUMsSUFBckMsRUFBMkMsSUFBM0MsQ0FBMUIsQ0FBQTtFQUNBLElBQUEsSUFBSUMsVUFBVSxHQUFHck0sT0FBTyxDQUFDNEIsU0FBUixDQUFrQjVCLE9BQU8sQ0FBQ0MsTUFBUixDQUFlSSxTQUFmLEdBQTJCLE1BQTdDLENBQWpCLENBQUE7TUFDQSxJQUFJaU0sWUFBWSxHQUFRLEVBQXhCLENBQUE7O0VBRUEsSUFBQSxJQUFJRCxVQUFVLElBQUlBLFVBQVUsS0FBSyxFQUFqQyxFQUFxQztRQUNuQ0MsWUFBWSxHQUFHL0osSUFBSSxDQUFDOEMsS0FBTCxDQUFXZ0gsVUFBWCxDQUFmLENBRG1DO0VBSW5DOztFQUNBLE1BQUEsSUFDRUQsaUJBQWlCLENBQUNHLElBQWxCLENBQ0UsVUFBQ0MsT0FBRCxFQUFRO0VBQUssUUFBQSxPQUFBRixZQUFZLENBQUNoSCxXQUFiLElBQTRCZ0gsWUFBWSxDQUFDaEgsV0FBYixDQUF5Qm1ILE9BQXpCLENBQWlDRCxPQUFqQyxDQUFBLEtBQThDLENBQUMsQ0FBM0UsQ0FBQTtFQUE0RSxPQUQzRixDQURGLEVBSUU7RUFDQUYsUUFBQUEsWUFBWSxHQUFHLEVBQWYsQ0FBQTtFQUNBRCxRQUFBQSxVQUFVLEdBQUcsRUFBYixDQUFBO0VBQ0QsT0Faa0M7OztFQWVuQyxNQUFBLElBQ0VyTSxPQUFPLENBQUNDLE1BQVIsQ0FBZVUsT0FBZixLQUEyQixJQUEzQixJQUNBMkwsWUFBWSxDQUFDeEMsRUFBYixLQUFvQi9JLFNBRHBCLElBRUF1TCxZQUFZLENBQUNoSCxXQUFiLEtBQTZCdkUsU0FIL0IsRUFJRTtFQUNBdUwsUUFBQUEsWUFBWSxHQUFHO1lBQ2JoSCxXQUFXLEVBQUVnSCxZQUFZLENBQUN4QyxFQURiO0VBRWJsQyxVQUFBQSxnQkFBZ0IsRUFBRTBFLFlBQVksQ0FBQzFFLGdCQUFiLElBQWlDLEtBRnRDO0VBR2JwQyxVQUFBQSxVQUFVLEVBQUUsSUFBSTNDLElBQUosRUFBQSxDQUFXSyxXQUFYLEVBSEM7RUFJYndKLFVBQUFBLElBQUksRUFBRUosWUFBWSxDQUFDSSxJQUFiLElBQXFCM0wsU0FBQUE7V0FKN0IsQ0FBQTtFQU1ELE9BQUE7RUFDRixLQUFBOztFQUVELElBQUEsSUFBSTRMLE1BQU0sR0FBRzNNLE9BQU8sQ0FBQzRNLGFBQVIsQ0FBc0JwSSxRQUFRLENBQUNxSSxHQUEvQixFQUFvQ3hOLFNBQVMsQ0FBQ0UsZ0JBQTlDLENBQWIsQ0FwQ1c7RUF1Q1g7O0VBQ0EsSUFBQSxJQUNFb04sTUFBTSxJQUNOQSxNQUFNLEtBQUssRUFEWCxJQUVBUCxpQkFBaUIsQ0FBQ1UsS0FBbEIsQ0FBd0IsVUFBQ04sT0FBRCxFQUFRO0VBQUssTUFBQSxPQUFBRyxNQUFNLENBQUNGLE9BQVAsQ0FBZUQsT0FBZixDQUFBLEtBQTRCLENBQUMsQ0FBN0IsQ0FBQTtFQUE4QixLQUFuRSxDQUhGLEVBSUU7RUFDQXhNLE1BQUFBLE9BQU8sQ0FBQytDLEdBQVIsQ0FBWSxNQUFaLEVBQW9CLHNCQUFwQixFQUE0QzRKLE1BQTVDLENBQUEsQ0FBQTtFQUVBLE1BQUEsSUFBTUksZUFBZSxHQUFHL00sT0FBTyxDQUFDNE0sYUFBUixDQUFzQnBJLFFBQVEsQ0FBQ3FJLEdBQS9CLEVBQW9DeE4sU0FBUyxDQUFDRyxxQkFBOUMsQ0FBeEIsQ0FBQTtFQUNBUSxNQUFBQSxPQUFPLENBQUMrQyxHQUFSLENBQVksTUFBWixFQUFvQiw4QkFBcEIsRUFBb0RnSyxlQUFwRCxDQUFBLENBQUE7UUFFQS9NLE9BQU8sQ0FBQ2MsV0FBUixHQUFzQjtFQUNwQndFLFFBQUFBLFdBQVcsRUFBRXFILE1BRE87RUFFcEIvRSxRQUFBQSxnQkFBZ0IsRUFBRW1GLGVBQWUsS0FBSyxNQUFwQixJQUE4QkEsZUFBZSxLQUFLLEdBRmhEO0VBR3BCdkgsUUFBQUEsVUFBVSxFQUFFLElBQUkzQyxJQUFKLEVBQUEsQ0FBV0ssV0FBWCxFQUhRO1VBSXBCd0osSUFBSSxFQUFFMU0sT0FBTyxDQUFDNE0sYUFBUixDQUFzQnBJLFFBQVEsQ0FBQ3FJLEdBQS9CLEVBQW9DeE4sU0FBUyxDQUFDSSxxQkFBOUMsQ0FBQTtFQUpjLE9BQXRCLENBTkE7O0VBZUEsTUFBQSxJQUFJNE0sVUFBVSxJQUFJQSxVQUFVLEtBQUssRUFBakMsRUFBcUM7RUFDbkNyTSxRQUFBQSxPQUFPLENBQUMrQyxHQUFSLENBQVksTUFBWixFQUFvQixpQ0FBcEIsRUFBdURzSixVQUF2RCxDQUFBLENBQUE7O0VBRUEsUUFBQSxJQUNFQyxZQUFZLENBQUMxRSxnQkFBYixLQUFrQyxLQUFsQyxJQUNBMEUsWUFBWSxDQUFDaEgsV0FBYixLQUE2QnRGLE9BQU8sQ0FBQ2MsV0FBUixDQUFvQndFLFdBRm5ELEVBR0U7RUFDQXRGLFVBQUFBLE9BQU8sQ0FBQytDLEdBQVIsQ0FDRSxNQURGLEVBRUUscUJBRkYsRUFHRXVKLFlBQVksQ0FBQ2hILFdBSGYsRUFJRSxXQUpGLEVBS0V0RixPQUFPLENBQUNjLFdBQVIsQ0FBb0J3RSxXQUx0QixDQUFBLENBQUE7RUFRQXRGLFVBQUFBLE9BQU8sQ0FBQ3NCLFVBQVIsQ0FBbUJFLEdBQW5CLENBQXVCLFlBQXZCLEVBQXFDO2NBQ25DNEcscUJBQXFCLEVBQUVrRSxZQUFZLENBQUNoSCxXQUREO0VBRW5DK0MsWUFBQUEsbUJBQW1CLEVBQUVySSxPQUFPLENBQUNjLFdBQVIsQ0FBb0J3RSxXQUZOO0VBR25DZ0QsWUFBQUEsd0JBQXdCLEVBQUV0SSxPQUFPLENBQUNjLFdBQVIsQ0FBb0I4RyxnQkFIWDtFQUluQ1csWUFBQUEsa0JBQWtCLEVBQUV2SSxPQUFPLENBQUNjLFdBQVIsQ0FBb0IwRSxVQUFBQTthQUoxQyxDQUFBLENBQUE7RUFNRCxTQUFBO0VBQ0YsT0FBQTs7UUFFRHhGLE9BQU8sQ0FBQ3dILFNBQVIsQ0FDRXhILE9BQU8sQ0FBQ0MsTUFBUixDQUFlSSxTQUFmLEdBQTJCLE1BRDdCLEVBRUVrQyxJQUFJLENBQUNDLFNBQUwsQ0FBZXhDLE9BQU8sQ0FBQ2MsV0FBdkIsQ0FGRixFQUdFcEIsZ0JBSEYsQ0FBQSxDQUFBO0VBS0EsTUFBQSxPQUFBO0VBQ0QsS0F6RlU7OztFQTZGWCxJQUFBLElBQUkyTSxVQUFVLElBQUlBLFVBQVUsS0FBSyxFQUFqQyxFQUFxQztFQUNuQ3JNLE1BQUFBLE9BQU8sQ0FBQytDLEdBQVIsQ0FBWSxNQUFaLEVBQW9CLHlCQUFwQixFQUErQ3NKLFVBQS9DLENBQUEsQ0FBQTtRQUNBck0sT0FBTyxDQUFDYyxXQUFSLEdBQXNCd0wsWUFBdEIsQ0FBQTtRQUNBdE0sT0FBTyxDQUFDd0gsU0FBUixDQUNFeEgsT0FBTyxDQUFDQyxNQUFSLENBQWVJLFNBQWYsR0FBMkIsTUFEN0IsRUFFRWtDLElBQUksQ0FBQ0MsU0FBTCxDQUFleEMsT0FBTyxDQUFDYyxXQUF2QixDQUZGLEVBR0VwQixnQkFIRixDQUFBLENBQUE7RUFLQSxNQUFBLE9BQUE7RUFDRCxLQXRHVTs7O0VBeUdYTSxJQUFBQSxPQUFPLENBQUNnTixXQUFSLENBQW9CaE4sT0FBTyxDQUFDdUYsTUFBUixFQUFwQixFQUFzQyxLQUF0QyxFQUE2QyxJQUFJMUMsSUFBSixFQUFBLENBQVdLLFdBQVgsRUFBN0MsQ0FBQSxDQUFBO0tBMTdCc0I7RUE2N0J4QjhKLEVBQUFBLFdBQVcsRUFBRSxTQUFDQyxXQUFBQSxDQUFBQSxjQUFELEVBQXlCRixlQUF6QixFQUFtREcsU0FBbkQsRUFBb0U7TUFDL0VsTixPQUFPLENBQUNjLFdBQVIsR0FBc0I7RUFDcEJ3RSxNQUFBQSxXQUFXLEVBQUUySCxjQURPO0VBRXBCckYsTUFBQUEsZ0JBQWdCLEVBQUVtRixlQUZFO0VBR3BCdkgsTUFBQUEsVUFBVSxFQUFFMEgsU0FBQUE7T0FIZCxDQUFBO0VBS0FsTixJQUFBQSxPQUFPLENBQUMrQyxHQUFSLENBQVksTUFBWixFQUFvQixtQkFBcEIsRUFBdUNiLFFBQUEsQ0FBQSxFQUFBLEVBQU9sQyxPQUFPLENBQUNjLFdBQWYsQ0FBdkMsQ0FBQSxDQUFBO01BQ0FkLE9BQU8sQ0FBQ3dILFNBQVIsQ0FDRXhILE9BQU8sQ0FBQ0MsTUFBUixDQUFlSSxTQUFmLEdBQTJCLE1BRDdCLEVBRUVrQyxJQUFJLENBQUNDLFNBQUwsQ0FBZXhDLE9BQU8sQ0FBQ2MsV0FBdkIsQ0FGRixFQUdFcEIsZ0JBSEYsQ0FBQSxDQUFBO0tBcDhCc0I7RUEyOEJ4QjtFQUNBeU4sRUFBQUEsa0JBQWtCLEVBQUUsU0FBQSxrQkFBQSxHQUFBO0VBQ2xCO01BQ0EsSUFBTUMsWUFBWSxHQUFHLENBQ25CO0VBQUVqRixNQUFBQSxHQUFHLEVBQUUsT0FBUDtRQUFnQmtGLEtBQUssRUFBRXJOLE9BQU8sQ0FBQzRNLGFBQVIsQ0FBc0JwSSxRQUFRLENBQUNxSSxHQUEvQixFQUFvQyxRQUFwQyxDQUFBO0VBQXZCLEtBRG1CLEVBRW5CO0VBQUUxRSxNQUFBQSxHQUFHLEVBQUUsV0FBUDtRQUFvQmtGLEtBQUssRUFBRXJOLE9BQU8sQ0FBQzRNLGFBQVIsQ0FBc0JwSSxRQUFRLENBQUNxSSxHQUEvQixFQUFvQyxXQUFwQyxDQUFBO0VBQTNCLEtBRm1CLEVBR25CO0VBQUUxRSxNQUFBQSxHQUFHLEVBQUUsWUFBUDtRQUFxQmtGLEtBQUssRUFBRXJOLE9BQU8sQ0FBQzRNLGFBQVIsQ0FBc0JwSSxRQUFRLENBQUNxSSxHQUEvQixFQUFvQyxZQUFwQyxDQUFBO0VBQTVCLEtBSG1CLEVBSW5CO0VBQUUxRSxNQUFBQSxHQUFHLEVBQUUsY0FBUDtRQUF1QmtGLEtBQUssRUFBRXJOLE9BQU8sQ0FBQzRNLGFBQVIsQ0FBc0JwSSxRQUFRLENBQUNxSSxHQUEvQixFQUFvQyxjQUFwQyxDQUFBO0VBQTlCLEtBSm1CLEVBS25CO0VBQUUxRSxNQUFBQSxHQUFHLEVBQUUsV0FBUDtRQUFvQmtGLEtBQUssRUFBRXJOLE9BQU8sQ0FBQzRNLGFBQVIsQ0FBc0JwSSxRQUFRLENBQUNxSSxHQUEvQixFQUFvQyxZQUFwQyxDQUFBO09BTFIsQ0FBckIsQ0FGa0I7O0VBV2xCTyxJQUFBQSxZQUFZLENBQUN0RyxPQUFiLENBQXFCLFVBQUN3RyxDQUFELEVBQUl2QyxDQUFKLEVBQUs7UUFDeEIsSUFBSXVDLENBQUMsQ0FBQ0QsS0FBRixJQUFXQyxDQUFDLENBQUNELEtBQUYsS0FBWSxFQUEzQixFQUErQjtVQUM3QnJOLE9BQU8sQ0FBQ2MsV0FBUixDQUFvQndNLENBQUMsQ0FBQ25GLEdBQXRCLENBQUEsR0FBNkJtRixDQUFDLENBQUNELEtBQS9CLENBQUE7RUFDRCxPQUFBO0VBQ0YsS0FKRCxFQVhrQjs7TUFrQmxCck4sT0FBTyxDQUFDd0gsU0FBUixDQUNFeEgsT0FBTyxDQUFDQyxNQUFSLENBQWVJLFNBQWYsR0FBMkIsTUFEN0IsRUFFRWtDLElBQUksQ0FBQ0MsU0FBTCxDQUFleEMsT0FBTyxDQUFDYyxXQUF2QixDQUZGLEVBR0VwQixnQkFIRixDQUFBLENBQUE7S0E5OUJzQjtFQXErQnhCO0VBQ0E7RUFDQTtFQUNBO0VBQ0E7RUFDQTZOLEVBQUFBLGFBQWEsRUFBRSxTQUFBLGFBQUEsR0FBQTtFQUNiLElBQUEsSUFBSUMsUUFBUSxHQUFHeE4sT0FBTyxDQUFDNE0sYUFBUixDQUFzQnBJLFFBQVEsQ0FBQ3FJLEdBQS9CLEVBQW9DeE4sU0FBUyxDQUFDQyxrQkFBOUMsQ0FBZixDQURhOztFQUliLElBQUEsSUFBSWtPLFFBQVEsSUFBSUEsUUFBUSxLQUFLLEVBQTdCLEVBQWlDO0VBQy9CeE4sTUFBQUEsT0FBTyxDQUFDK0MsR0FBUixDQUFZLE1BQVosRUFBb0Isd0JBQXBCLEVBQThDeUssUUFBOUMsQ0FBQSxDQUFBO1FBQ0F4TixPQUFPLENBQUNnQixhQUFSLEdBQXdCO0VBQ3RCc0UsUUFBQUEsV0FBVyxFQUFFa0ksUUFEUztFQUV0QmhJLFFBQUFBLFVBQVUsRUFBRSxJQUFJM0MsSUFBSixFQUFBLENBQVdLLFdBQVgsRUFGVTtVQUd0QjhGLFVBQVUsRUFBRUMsU0FBUyxDQUFDQyxTQUFBQTtTQUh4QixDQUFBO1FBS0FsSixPQUFPLENBQUN3SCxTQUFSLENBQ0V4SCxPQUFPLENBQUNDLE1BQVIsQ0FBZUksU0FBZixHQUEyQixRQUQ3QixFQUVFa0MsSUFBSSxDQUFDQyxTQUFMLENBQWV4QyxPQUFPLENBQUNnQixhQUF2QixDQUZGLEVBR0V0QixnQkFIRixDQUFBLENBQUE7RUFLQSxNQUFBLE9BQUE7RUFDRCxLQWpCWTs7O0VBb0JiLElBQUEsSUFBTStOLFlBQVksR0FBR3pOLE9BQU8sQ0FBQzRCLFNBQVIsQ0FBa0I1QixPQUFPLENBQUNDLE1BQVIsQ0FBZUksU0FBZixHQUEyQixRQUE3QyxDQUFyQixDQUFBOztFQUVBLElBQUEsSUFBSW9OLFlBQVksSUFBSUEsWUFBWSxLQUFLLEVBQXJDLEVBQXlDO0VBQ3ZDek4sTUFBQUEsT0FBTyxDQUFDK0MsR0FBUixDQUFZLE1BQVosRUFBb0IsMkJBQXBCLEVBQWlEMEssWUFBakQsQ0FBQSxDQUFBO1FBQ0F6TixPQUFPLENBQUNnQixhQUFSLEdBQXdCdUIsSUFBSSxDQUFDOEMsS0FBTCxDQUFXb0ksWUFBWCxDQUF4QixDQUZ1Qzs7RUFLdkMsTUFBQSxJQUFJek4sT0FBTyxDQUFDQyxNQUFSLENBQWVVLE9BQWYsS0FBMkIsSUFBL0IsRUFBcUM7RUFDbkMsUUFBQSxJQUFNK00sY0FBYyxHQUFHMU4sT0FBTyxDQUFDNEIsU0FBUixDQUFrQixTQUFsQixDQUF2QixDQUFBOztFQUVBLFFBQUEsSUFBSThMLGNBQWMsSUFBSUEsY0FBYyxLQUFLLEVBQXpDLEVBQTZDO0VBQzNDMU4sVUFBQUEsT0FBTyxDQUFDZ0IsYUFBUixDQUFzQnNFLFdBQXRCLEdBQW9Db0ksY0FBcEMsQ0FEMkM7O0VBRzNDLFVBQUEsSUFBTUMscUJBQXFCLEdBQUczTixPQUFPLENBQUM0QixTQUFSLENBQWtCLFdBQWxCLENBQTlCLENBQUE7O0VBQ0EsVUFBQSxJQUFJK0wscUJBQXFCLElBQUlBLHFCQUFxQixLQUFLLEVBQXZELEVBQTJEO0VBQ3pEM04sWUFBQUEsT0FBTyxDQUFDZ0IsYUFBUixDQUFzQndFLFVBQXRCLEdBQW1DLElBQUkzQyxJQUFKLENBQ2pDK0ssUUFBUSxDQUFDRCxxQkFBRCxFQUF3QixFQUF4QixDQUR5QixDQUFBLENBRWpDekssV0FGaUMsRUFBbkMsQ0FBQTtFQUdELFdBQUE7RUFDRixTQUFBO0VBQ0YsT0FBQTs7UUFFRGxELE9BQU8sQ0FBQ3dILFNBQVIsQ0FDRXhILE9BQU8sQ0FBQ0MsTUFBUixDQUFlSSxTQUFmLEdBQTJCLFFBRDdCLEVBRUVrQyxJQUFJLENBQUNDLFNBQUwsQ0FBZXhDLE9BQU8sQ0FBQ2dCLGFBQXZCLENBRkYsRUFHRXRCLGdCQUhGLENBQUEsQ0FBQTtFQUtBLE1BQUEsT0FBQTtFQUNELEtBaERZOzs7RUFtRGJNLElBQUFBLE9BQU8sQ0FBQzhILGFBQVIsRUFBQSxDQUFBO0tBN2hDc0I7RUFnaUN4QkEsRUFBQUEsYUFBYSxFQUFFLFNBQUEsYUFBQSxHQUFBO0VBQ2I5SCxJQUFBQSxPQUFPLENBQUMrQyxHQUFSLENBQVksTUFBWixFQUFvQix3QkFBcEIsQ0FBQSxDQUFBO01BQ0EvQyxPQUFPLENBQUNnQixhQUFSLEdBQXdCO0VBQ3RCc0UsTUFBQUEsV0FBVyxFQUFFdEYsT0FBTyxDQUFDdUYsTUFBUixFQURTO0VBRXRCQyxNQUFBQSxVQUFVLEVBQUUsSUFBSTNDLElBQUosRUFBQSxDQUFXSyxXQUFYLEVBRlU7UUFHdEI4RixVQUFVLEVBQUVDLFNBQVMsQ0FBQ0MsU0FBQUE7T0FIeEIsQ0FBQTtNQUtBbEosT0FBTyxDQUFDd0gsU0FBUixDQUNFeEgsT0FBTyxDQUFDQyxNQUFSLENBQWVJLFNBQWYsR0FBMkIsUUFEN0IsRUFFRWtDLElBQUksQ0FBQ0MsU0FBTCxDQUFleEMsT0FBTyxDQUFDZ0IsYUFBdkIsQ0FGRixFQUdFdEIsZ0JBSEYsQ0FBQSxDQUFBO0tBdmlDc0I7RUE4aUN4QjtJQUNBbU8saUJBQWlCLEVBQUUsMkJBQUNDLE9BQUQsRUFBVUMsU0FBVixFQUFxQkMsWUFBckIsRUFBbUNDLFVBQW5DLEVBQTZDO01BQzlELElBQUlILE9BQU8sQ0FBQ0ksZ0JBQVosRUFBOEI7RUFDNUJKLE1BQUFBLE9BQU8sQ0FBQ0ksZ0JBQVIsQ0FBeUJILFNBQXpCLEVBQW9DQyxZQUFwQyxFQUFrREMsVUFBbEQsQ0FBQSxDQUFBO0VBQ0EsTUFBQSxPQUFPLElBQVAsQ0FBQTtFQUNELEtBQUE7O01BQ0QsSUFBSUgsT0FBTyxDQUFDSyxXQUFaLEVBQXlCO1FBQ3ZCLE9BQU9MLE9BQU8sQ0FBQ0ssV0FBUixDQUFvQixPQUFPSixTQUEzQixFQUFzQ0MsWUFBdEMsQ0FBUCxDQUFBO0VBQ0QsS0FBQTs7RUFDREYsSUFBQUEsT0FBTyxDQUFDLElBQUEsR0FBT0MsU0FBUixDQUFQLEdBQTRCQyxZQUE1QixDQUFBO0tBdmpDc0I7RUEwakN4QjtJQUNBdEosUUFBUSxFQUFFLFNBQUNILFFBQUFBLENBQUFBLEdBQUQsRUFBYTtFQUNyQjtNQUNBLElBQUl2RSxPQUFPLENBQUNZLE9BQVosRUFBcUIsT0FBQTtNQUNyQlosT0FBTyxDQUFDWSxPQUFSLEdBQWtCLElBQWxCLENBQUE7RUFFQVosSUFBQUEsT0FBTyxDQUFDK0MsR0FBUixDQUFZLE1BQVosRUFBb0Isa0JBQXBCLEVBTHFCOztFQVFyQixJQUFBLElBQUksQ0FBQy9DLE9BQU8sQ0FBQ29PLGNBQVIsRUFBTCxFQUErQjtFQUM3QnBPLE1BQUFBLE9BQU8sQ0FBQytDLEdBQVIsQ0FBWSxNQUFaLEVBQW9CLHNCQUFwQixDQUFBLENBQUE7RUFDQSxNQUFBLE9BQUE7RUFDRCxLQVhvQjs7O0VBY3JCLElBQUEsSUFBTTlDLE1BQU0seUJBQVFELE9BQU8sQ0FBQ0MsU0FBV3NFLElBQXZDLENBZHFCO0VBaUJyQjtFQUNBO0VBQ0E7RUFFQTs7O01BQ0F2RSxPQUFPLENBQUNDLE1BQVIsR0FBaUJBLE1BQWpCLENBQUE7TUFDQUQsT0FBTyxDQUFDK0MsR0FBUixDQUFZLE1BQVosRUFBb0IsZ0JBQXBCLEVBQXNDL0MsT0FBTyxDQUFDQyxNQUE5QyxDQUFBLENBdkJxQjs7TUEwQnJCLElBQUk0RixNQUFNLENBQUN3SSxZQUFYLEVBQXlCO1FBQ3ZCck8sT0FBTyxDQUFDc0IsVUFBUixDQUFtQkMsS0FBbkIsR0FBMkJnQixJQUFJLENBQUM4QyxLQUFMLENBQ3pCZ0osWUFBWSxDQUFDQyxPQUFiLENBQXFCdE8sT0FBTyxDQUFDQyxNQUFSLENBQWVJLFNBQWYsR0FBMkIsT0FBaEQsQ0FBNEQsSUFBQSxJQURuQyxDQUEzQixDQUFBO1FBR0FMLE9BQU8sQ0FBQ3VELGFBQVIsR0FBd0JoQixJQUFJLENBQUM4QyxLQUFMLENBQ3RCZ0osWUFBWSxDQUFDQyxPQUFiLENBQXFCdE8sT0FBTyxDQUFDQyxNQUFSLENBQWVJLFNBQWYsR0FBMkIsZUFBaEQsQ0FBQSxJQUFvRSxJQUQ5QyxDQUF4QixDQUFBO0VBR0QsS0FqQ29COzs7RUFvQ3JCTCxJQUFBQSxPQUFPLENBQUNtTSxXQUFSLEVBQUEsQ0FBQTs7TUFDQW5NLE9BQU8sQ0FBQ21OLGtCQUFSLEVBQUEsQ0FyQ3FCOzs7TUF1Q3JCbk4sT0FBTyxDQUFDdU4sYUFBUixFQUFBLENBdkNxQjs7O01BeUNyQnZOLE9BQU8sQ0FBQ2dJLGNBQVIsRUFBQSxDQXpDcUI7OztFQTRDckIsSUFBQSxJQUFJaEksT0FBTyxDQUFDeUQsWUFBUixDQUFxQkssTUFBckIsR0FBOEIsQ0FBbEMsRUFBcUM7UUFDbkM5RCxPQUFPLENBQUN5RCxZQUFSLENBQXFCcUQsT0FBckIsQ0FBNkIsVUFBQ3dHLENBQUQsRUFBSXZDLENBQUosRUFBSztFQUNoQy9LLFFBQUFBLE9BQU8sQ0FBQytDLEdBQVIsQ0FBWSxPQUFaLEVBQXFCLDJCQUFyQixFQUFrRHVLLENBQWxELENBQUEsQ0FBQTtVQUNBQSxDQUFDLEVBQUEsQ0FBQTtTQUZILENBQUEsQ0FBQTtRQUtBdE4sT0FBTyxDQUFDeUQsWUFBUixHQUF1QixFQUF2QixDQUFBO0VBQ0QsS0FuRG9CO0VBc0RyQjtFQUNBO0VBQ0E7RUFFQTtFQUNBO0VBQ0E7RUFFQTtFQUNBO0VBQ0E7RUFFQTs7O01BQ0FvQyxNQUFNLENBQUMwSSxXQUFQLENBQW1CLFlBQUE7UUFDakIsSUFBSXZPLE9BQU8sQ0FBQ2tCLGVBQVIsSUFBMkJsQixPQUFPLENBQUNrRyxhQUFSLEVBQS9CLEVBQXdEO0VBQ3REO0VBQ0EsUUFBQSxJQUFNc0ksYUFBYSxHQUFHeE8sT0FBTyxDQUFDNEIsU0FBUixDQUFrQjVCLE9BQU8sQ0FBQ0MsTUFBUixDQUFlSSxTQUFmLEdBQTJCLFNBQTdDLENBQXRCLENBQUE7O0VBQ0EsUUFBQSxJQUFJbU8sYUFBSixFQUFtQjtFQUNqQjtFQUNBeE8sVUFBQUEsT0FBTyxDQUFDd0gsU0FBUixDQUNFeEgsT0FBTyxDQUFDQyxNQUFSLENBQWVJLFNBQWYsR0FBMkIsU0FEN0IsRUFFRW1PLGFBRkYsRUFHRXhPLE9BQU8sQ0FBQ0MsTUFBUixDQUFlRyxlQUhqQixDQUFBLENBQUE7RUFLRCxTQUdBO0VBQ0YsT0FBQTtPQWZILEVBZ0JHLEtBaEJILENBQUEsQ0FuRXFCO0VBcUZyQjs7TUFDQSxJQUFJSixPQUFPLENBQUNDLE1BQVIsQ0FBZUssYUFBZixDQUE2QndELE1BQTdCLEdBQXNDLENBQTFDLEVBQTZDO0VBQzNDO0VBQ0EsTUFBQSxLQUFLLElBQUlpSCxDQUFDLEdBQUcsQ0FBYixFQUFnQkEsQ0FBQyxHQUFHdkcsUUFBUSxDQUFDaUssS0FBVCxDQUFlM0ssTUFBbkMsRUFBMkNpSCxDQUFDLEVBQTVDLEVBQWdEO0VBQzlDLFFBQUEsSUFBSTJELEdBQUcsR0FBR2xLLFFBQVEsQ0FBQ2lLLEtBQVQsQ0FBZTFELENBQWYsQ0FBVixDQUFBO1VBRUEvSyxPQUFPLENBQUNDLE1BQVIsQ0FBZUssYUFBZixDQUE2QndHLE9BQTdCLENBQXFDLFVBQUM2SCxDQUFELEVBQUU7RUFDckM7WUFDQSxJQUFJRCxHQUFHLENBQUMzSSxJQUFKLENBQVMwRyxPQUFULENBQWlCa0MsQ0FBakIsQ0FBQSxLQUF3QixDQUFDLENBQTdCLEVBQWdDO2NBQzlCM08sT0FBTyxDQUFDNk4saUJBQVIsQ0FBMEJhLEdBQTFCLEVBQStCLE9BQS9CLEVBQXdDMU8sT0FBTyxDQUFDNE8sWUFBaEQsRUFBOEQsSUFBOUQsQ0FBQSxDQUFBOztjQUNBNU8sT0FBTyxDQUFDNk4saUJBQVIsQ0FBMEJhLEdBQTFCLEVBQStCLFdBQS9CLEVBQTRDMU8sT0FBTyxDQUFDNE8sWUFBcEQsRUFBa0UsSUFBbEUsQ0FBQSxDQUFBO0VBQ0QsV0FBQTtXQUxILENBQUEsQ0FBQTtFQU9ELE9BQUE7RUFDRixLQW5Hb0I7RUFzR3JCO0VBQ0E7RUFFQTs7O0VBQ0FsRyxJQUFBQSxDQUFTLENBQUN3RixnQkFBVixDQUEyQixhQUEzQixFQUEwQyxVQUFDVyxLQUFELEVBQU07RUFDOUM3TyxNQUFBQSxPQUFPLENBQUMrQyxHQUFSLENBQVksTUFBWixFQUFvQix5QkFBcEIsRUFBK0M4TCxLQUFLLENBQUNDLFFBQXJELEVBQStELElBQS9ELEVBQXFFRCxLQUFLLENBQUNFLFFBQTNFLENBQUEsQ0FBQTs7RUFFQSxNQUFBLElBQUlGLEtBQUssQ0FBQ0MsUUFBTixLQUFtQm5QLFVBQVUsQ0FBQ0MsTUFBOUIsSUFBd0NpUCxLQUFLLENBQUNFLFFBQU4sS0FBbUJwUCxVQUFVLENBQUNFLE9BQTFFLEVBQW1GO0VBQ2pGRyxRQUFBQSxPQUFPLENBQUNnUCxjQUFSLEVBQUEsQ0FBQTtFQUNELE9BRkQsTUFFTyxJQUFJSCxLQUFLLENBQUNDLFFBQU4sS0FBbUJuUCxVQUFVLENBQUNFLE9BQTlCLElBQXlDZ1AsS0FBSyxDQUFDRSxRQUFOLEtBQW1CcFAsVUFBVSxDQUFDQyxNQUEzRSxFQUFtRjtFQUN4RkksUUFBQUEsT0FBTyxDQUFDaVAsYUFBUixFQUFBLENBQUE7RUFDRCxPQUFBO0VBQ0YsS0FSRCxFQTFHcUI7RUFxSHJCO0VBQ0E7RUFDQTtLQWxyQ3NCO0lBcXJDeEJoSyxjQUFjLEVBQUUsU0FBQ0UsY0FBQUEsQ0FBQUEsRUFBRCxFQUFhO01BQzNCLElBQUluRixPQUFPLENBQUNZLE9BQVosRUFBcUI7UUFDbkJ1RSxFQUFFLEVBQUEsQ0FBQTtFQUNILEtBRkQsTUFFTztFQUNMbkYsTUFBQUEsT0FBTyxDQUFDeUQsWUFBUixDQUFxQnJCLElBQXJCLENBQTBCK0MsRUFBMUIsQ0FBQSxDQUFBO0VBQ0QsS0FBQTtLQTFyQ3FCO0VBNnJDeEI7RUFDQTtFQUNBO0VBQ0E7RUFDQTtFQUNBNkMsRUFBQUEsY0FBYyxFQUFFLFNBQUEsY0FBQSxHQUFBO0VBQ2Q7TUFDQSxJQUFJa0gsVUFBVSxHQUNabFAsT0FBTyxDQUFDNE0sYUFBUixDQUFzQnBJLFFBQVEsQ0FBQ3FJLEdBQS9CLEVBQW9DLFlBQXBDLEtBQ0E3TSxPQUFPLENBQUNtUCxZQUFSLENBQXFCdEosTUFBTSxDQUFDQyxRQUFQLENBQWdCbUIsSUFBckMsRUFBMkMsWUFBM0MsQ0FGRixDQUFBO01BR0EsSUFBSW1JLFVBQVUsR0FDWnBQLE9BQU8sQ0FBQzRNLGFBQVIsQ0FBc0JwSSxRQUFRLENBQUNxSSxHQUEvQixFQUFvQyxZQUFwQyxLQUNBN00sT0FBTyxDQUFDbVAsWUFBUixDQUFxQnRKLE1BQU0sQ0FBQ0MsUUFBUCxDQUFnQm1CLElBQXJDLEVBQTJDLFlBQTNDLENBRkYsQ0FBQTtNQUdBLElBQUlvSSxZQUFZLEdBQ2RyUCxPQUFPLENBQUM0TSxhQUFSLENBQXNCcEksUUFBUSxDQUFDcUksR0FBL0IsRUFBb0MsY0FBcEMsS0FDQTdNLE9BQU8sQ0FBQ21QLFlBQVIsQ0FBcUJ0SixNQUFNLENBQUNDLFFBQVAsQ0FBZ0JtQixJQUFyQyxFQUEyQyxjQUEzQyxDQUZGLENBQUE7TUFHQSxJQUFJcUksV0FBVyxHQUNidFAsT0FBTyxDQUFDNE0sYUFBUixDQUFzQnBJLFFBQVEsQ0FBQ3FJLEdBQS9CLEVBQW9DLGFBQXBDLEtBQ0E3TSxPQUFPLENBQUNtUCxZQUFSLENBQXFCdEosTUFBTSxDQUFDQyxRQUFQLENBQWdCbUIsSUFBckMsRUFBMkMsYUFBM0MsQ0FGRixDQUFBO01BR0EsSUFBSXNJLFFBQVEsR0FDVnZQLE9BQU8sQ0FBQzRNLGFBQVIsQ0FBc0JwSSxRQUFRLENBQUNxSSxHQUEvQixFQUFvQyxVQUFwQyxLQUNBN00sT0FBTyxDQUFDbVAsWUFBUixDQUFxQnRKLE1BQU0sQ0FBQ0MsUUFBUCxDQUFnQm1CLElBQXJDLEVBQTJDLFVBQTNDLENBRkYsQ0FBQTtNQUdBLElBQUl1SSxNQUFNLEdBQ1J4UCxPQUFPLENBQUM0TSxhQUFSLENBQXNCcEksUUFBUSxDQUFDcUksR0FBL0IsRUFBb0MsUUFBcEMsS0FDQTdNLE9BQU8sQ0FBQ21QLFlBQVIsQ0FBcUJ0SixNQUFNLENBQUNDLFFBQVAsQ0FBZ0JtQixJQUFyQyxFQUEyQyxRQUEzQyxDQUZGLENBQUE7TUFHQSxJQUFJd0ksV0FBVyxHQUNielAsT0FBTyxDQUFDNE0sYUFBUixDQUFzQnBJLFFBQVEsQ0FBQ3FJLEdBQS9CLEVBQW9DLGFBQXBDLENBQUEsSUFDQTdNLE9BQU8sQ0FBQ21QLFlBQVIsQ0FBcUJ0SixNQUFNLENBQUNDLFFBQVAsQ0FBZ0JtQixJQUFyQyxFQUEyQyxhQUEzQyxDQUZGLENBcEJjOztFQXlCZCxJQUFBLElBQU14QixRQUFRLEdBQUd6RixPQUFPLENBQUMwRixXQUFSLEVBQWpCLENBQUE7O0VBRUEsSUFBQSxJQUFJRCxRQUFKLEVBQWM7RUFDWjtFQUNBLE1BQUEsSUFBSWlLLFdBQVcsR0FBR2xMLFFBQVEsQ0FBQ21MLGFBQVQsQ0FBdUIsR0FBdkIsQ0FBbEIsQ0FBQTtFQUNBRCxNQUFBQSxXQUFXLENBQUMzSixJQUFaLEdBQW1CTixRQUFuQixDQUhZOztFQU1aLE1BQUEsSUFBSW1LLGlCQUFpQixHQUFHLEtBQXhCLENBTlk7O0VBU1osTUFBQSxJQUFJRixXQUFXLENBQUNHLFFBQVosSUFBd0JILFdBQVcsQ0FBQ0csUUFBWixLQUF5QmhLLE1BQU0sQ0FBQ0MsUUFBUCxDQUFnQitKLFFBQXJFLEVBQStFO1VBQzdFLElBQUlDLGVBQWEsR0FBRyxLQUFwQixDQUFBOztFQUVBLFFBQUEsSUFBSTlQLE9BQU8sQ0FBQ0MsTUFBUixDQUFlSyxhQUFmLElBQWdDTixPQUFPLENBQUNDLE1BQVIsQ0FBZUssYUFBZixDQUE2QndELE1BQWpFLEVBQXlFO1lBQ3ZFOUQsT0FBTyxDQUFDQyxNQUFSLENBQWVLLGFBQWYsQ0FBNkJ3RyxPQUE3QixDQUFxQyxVQUFDaUosR0FBRCxFQUFJO2NBQ3ZDLElBQUlMLFdBQVcsQ0FBQzNKLElBQVosQ0FBaUIwRyxPQUFqQixDQUF5QnNELEdBQXpCLENBQUEsS0FBa0MsQ0FBQyxDQUF2QyxFQUEwQztFQUN4Q0QsY0FBQUEsZUFBYSxHQUFHLElBQWhCLENBQUE7RUFDRCxhQUFBO2FBSEgsQ0FBQSxDQUFBO0VBS0QsU0FBQTs7VUFFRCxJQUFJQSxlQUFhLEtBQUssS0FBdEIsRUFBNkI7RUFDM0JGLFVBQUFBLGlCQUFpQixHQUFHLElBQXBCLENBQUE7RUFDRCxTQUFBO0VBQ0YsT0FBQTs7RUFFRCxNQUFBLElBQUlBLGlCQUFKLEVBQXVCO0VBQ3JCVixRQUFBQSxVQUFVLEdBQUdRLFdBQVcsQ0FBQ0csUUFBekIsQ0FEcUI7O0VBSXJCLFFBQUEsSUFBSSxDQUFDVCxVQUFELElBQWVBLFVBQVUsS0FBSyxFQUFsQyxFQUFzQztFQUNwQ0EsVUFBQUEsVUFBVSxHQUFHLFVBQWIsQ0FBQTtFQUNELFNBTm9COzs7RUFTckIsUUFBQSxJQUFJM0osUUFBUSxDQUFDdUssTUFBVCxDQUFnQiw4QkFBaEIsQ0FBQSxLQUFvRCxDQUF4RCxFQUEyRDtZQUN6RFQsUUFBUSxHQUFHdlAsT0FBTyxDQUFDNE0sYUFBUixDQUFzQm5ILFFBQXRCLEVBQWdDLEdBQWhDLENBQVgsQ0FBQTtXQURGLE1BRU8sSUFBSUEsUUFBUSxDQUFDdUssTUFBVCxDQUFnQix1QkFBaEIsQ0FBNkMsS0FBQSxDQUFqRCxFQUFvRDtZQUN6RFQsUUFBUSxHQUFHdlAsT0FBTyxDQUFDNE0sYUFBUixDQUFzQm5ILFFBQXRCLEVBQWdDLEdBQWhDLENBQVgsQ0FBQTtXQURLLE1BRUEsSUFBSUEsUUFBUSxDQUFDdUssTUFBVCxDQUFnQiwrQkFBaEIsQ0FBcUQsS0FBQSxDQUF6RCxFQUE0RDtZQUNqRVQsUUFBUSxHQUFHdlAsT0FBTyxDQUFDNE0sYUFBUixDQUFzQm5ILFFBQXRCLEVBQWdDLEdBQWhDLENBQVgsQ0FBQTtXQURLLE1BRUEsSUFBSUEsUUFBUSxDQUFDdUssTUFBVCxDQUFnQixzQkFBaEIsQ0FBNEMsS0FBQSxDQUFoRCxFQUFtRDtZQUN4RFQsUUFBUSxHQUFHdlAsT0FBTyxDQUFDNE0sYUFBUixDQUFzQm5ILFFBQXRCLEVBQWdDLEdBQWhDLENBQVgsQ0FBQTtXQURLLE1BRUEsSUFBSUEsUUFBUSxDQUFDdUssTUFBVCxDQUFnQiw2QkFBaEIsQ0FBbUQsS0FBQSxDQUF2RCxFQUEwRDtZQUMvRFQsUUFBUSxHQUFHdlAsT0FBTyxDQUFDNE0sYUFBUixDQUFzQm5ILFFBQXRCLEVBQWdDLEdBQWhDLENBQVgsQ0FBQTtXQURLLE1BRUEsSUFBSUEsUUFBUSxDQUFDdUssTUFBVCxDQUFnQiw2QkFBaEIsQ0FBbUQsS0FBQSxDQUF2RCxFQUEwRDtZQUMvRFQsUUFBUSxHQUFHdlAsT0FBTyxDQUFDNE0sYUFBUixDQUFzQm5ILFFBQXRCLEVBQWdDLEdBQWhDLENBQVgsQ0FBQTtFQUNELFNBQUE7RUFDRixPQUFBO0VBQ0YsS0EzRWE7OztNQThFZCxJQUFNd0ssR0FBRyxHQUFHLENBQUMsT0FBRCxFQUFVLFFBQVYsRUFBb0IsU0FBcEIsQ0FBWixDQUFBO0VBQ0FBLElBQUFBLEdBQUcsQ0FBQ25KLE9BQUosQ0FBWSxVQUFDb0osS0FBRCxFQUFNO1FBQ2hCLElBQU03QyxLQUFLLEdBQ1RyTixPQUFPLENBQUM0TSxhQUFSLENBQXNCcEksUUFBUSxDQUFDcUksR0FBL0IsRUFBb0NxRCxLQUFwQyxLQUNBbFEsT0FBTyxDQUFDbVAsWUFBUixDQUFxQnRKLE1BQU0sQ0FBQ0MsUUFBUCxDQUFnQm1CLElBQXJDLEVBQTJDaUosS0FBM0MsQ0FGRixDQUFBOztFQUdBLE1BQUEsSUFBSTdDLEtBQUosRUFBVztFQUNUbUMsUUFBQUEsTUFBTSxHQUFHbkMsS0FBVCxDQUFBO0VBQ0FvQyxRQUFBQSxXQUFXLEdBQUdTLEtBQWQsQ0FBQTtFQUNELE9BQUE7RUFDRixLQVJELEVBL0VjO0VBMEZkO0VBQ0E7O0VBQ0EsSUFBQSxJQUFJVCxXQUFXLEtBQUssT0FBaEIsSUFBMkJMLFVBQVUsS0FBSyxVQUE5QyxFQUEwRDtFQUN4REEsTUFBQUEsVUFBVSxHQUFHLEtBQWIsQ0FBQTtFQUNELEtBQUE7O0VBRURwUCxJQUFBQSxPQUFPLENBQUMrQyxHQUFSLENBQVksTUFBWixFQUFvQixvQkFBcEIsRUFBMENtTSxVQUExQyxDQUFBLENBQUE7RUFDQWxQLElBQUFBLE9BQU8sQ0FBQytDLEdBQVIsQ0FBWSxNQUFaLEVBQW9CLG9CQUFwQixFQUEwQ3FNLFVBQTFDLENBQUEsQ0FBQTtFQUNBcFAsSUFBQUEsT0FBTyxDQUFDK0MsR0FBUixDQUFZLE1BQVosRUFBb0Isc0JBQXBCLEVBQTRDc00sWUFBNUMsQ0FBQSxDQUFBO0VBQ0FyUCxJQUFBQSxPQUFPLENBQUMrQyxHQUFSLENBQVksTUFBWixFQUFvQixxQkFBcEIsRUFBMkN1TSxXQUEzQyxDQUFBLENBQUE7RUFDQXRQLElBQUFBLE9BQU8sQ0FBQytDLEdBQVIsQ0FBWSxNQUFaLEVBQW9CLGtCQUFwQixFQUF3Q3dNLFFBQXhDLENBQUEsQ0FBQTtFQUNBdlAsSUFBQUEsT0FBTyxDQUFDK0MsR0FBUixDQUFZLE1BQVosRUFBb0IsZ0JBQXBCLEVBQXNDeU0sTUFBdEMsQ0FBQSxDQUFBO01BQ0F4UCxPQUFPLENBQUMrQyxHQUFSLENBQVksTUFBWixFQUFvQixxQkFBcEIsRUFBMkMwTSxXQUEzQyxDQUFBLENBdEdjOztFQXlHZCxJQUFBLElBQU05TixhQUFhLEdBQUczQixPQUFPLENBQUM0QixTQUFSLENBQWtCNUIsT0FBTyxDQUFDQyxNQUFSLENBQWVJLFNBQWYsR0FBMkIsU0FBN0MsQ0FBdEIsQ0F6R2M7O0VBNEdkLElBQUEsSUFBSSxDQUFDc0IsYUFBRCxJQUFrQkEsYUFBYSxLQUFLLEVBQXhDLEVBQTRDO0VBQzFDM0IsTUFBQUEsT0FBTyxDQUFDK0MsR0FBUixDQUFZLE1BQVosRUFBb0IsOEJBQXBCLENBQUEsQ0FBQTs7UUFFQS9DLE9BQU8sQ0FBQytCLGdCQUFSLENBQXlCO0VBQ3ZCbU4sUUFBQUEsVUFBVSxFQUFBQSxVQURhO0VBRXZCRSxRQUFBQSxVQUFVLEVBQUFBLFVBRmE7RUFHdkJDLFFBQUFBLFlBQVksRUFBQUEsWUFIVztFQUl2QkMsUUFBQUEsV0FBVyxFQUFBQSxXQUpZO0VBS3ZCQyxRQUFBQSxRQUFRLEVBQUFBLFFBTGU7RUFNdkJDLFFBQUFBLE1BQU0sRUFBQUEsTUFOaUI7RUFPdkJDLFFBQUFBLFdBQVcsRUFBQUEsV0FBQUE7U0FQYixDQUFBLENBQUE7O0VBU0EsTUFBQSxPQUFBO0VBQ0QsS0F6SGE7OztFQTRIZCxJQUFBLElBQUlVLGFBQUosQ0FBQTs7RUFFQSxJQUFBLElBQUlqQixVQUFVLElBQUlBLFVBQVUsS0FBSyxFQUE3QixJQUFtQ2xQLE9BQU8sQ0FBQ0MsTUFBUixDQUFlTSxlQUFmLENBQStCdUQsTUFBL0IsR0FBd0MsQ0FBL0UsRUFBa0Y7RUFDaEY7UUFDQXFNLGFBQWEsR0FBR25RLE9BQU8sQ0FBQ0MsTUFBUixDQUFlTSxlQUFmLENBQStCNlAsSUFBL0IsQ0FBb0MsVUFBQ0MsTUFBRCxFQUFPO0VBQ3pEO1VBQ0EsSUFBSUEsTUFBTSxDQUFDbkIsVUFBUCxLQUFzQkEsVUFBdEIsSUFBb0NtQixNQUFNLENBQUNqQixVQUFQLEtBQXNCQSxVQUE5RCxFQUEwRTtFQUN4RTtZQUNBLElBQUlpQixNQUFNLENBQUNoQixZQUFQLElBQXVCZ0IsTUFBTSxDQUFDaEIsWUFBUCxLQUF3QixFQUFuRCxFQUF1RDtFQUNyRCxZQUFBLElBQUlBLFlBQVksSUFBSWdCLE1BQU0sQ0FBQ2hCLFlBQVAsS0FBd0JBLFlBQTVDLEVBQTBEO0VBQ3hELGNBQUEsT0FBTyxJQUFQLENBQUE7RUFDRCxhQUhvRDs7O0VBS3JELFlBQUEsT0FBTyxLQUFQLENBQUE7RUFDRCxXQVJ1RTs7O0VBVXhFLFVBQUEsT0FBTyxJQUFQLENBQUE7RUFDRCxTQUFBOztFQUNELFFBQUEsT0FBTyxLQUFQLENBQUE7RUFDRCxPQWZlLENBQWhCLENBQUE7RUFnQkQsS0FoSmE7OztFQW1KZCxJQUFBLElBQUlpQixlQUFlLEdBQUcvTixJQUFJLENBQUM4QyxLQUFMLENBQVcxRCxhQUFYLENBQXRCLENBQUE7TUFDQTNCLE9BQU8sQ0FBQytDLEdBQVIsQ0FBWSxNQUFaLEVBQW9CLDBCQUFwQixFQUFnRHVOLGVBQWhELENBQUEsQ0FwSmM7O01BdUpkLElBQUlDLE9BQU8sR0FBRyxJQUFkLENBQUE7RUFDQSxJQUFBLElBQUlyQixVQUFVLElBQUlBLFVBQVUsS0FBSyxFQUE3QixJQUFtQ29CLGVBQWUsQ0FBQ3BCLFVBQWhCLEtBQStCQSxVQUF0RSxFQUNFcUIsT0FBTyxHQUFHLEtBQVYsQ0FBQTtFQUNGLElBQUEsSUFBSW5CLFVBQVUsSUFBSUEsVUFBVSxLQUFLLEVBQTdCLElBQW1Da0IsZUFBZSxDQUFDbEIsVUFBaEIsS0FBK0JBLFVBQXRFLEVBQ0VtQixPQUFPLEdBQUcsS0FBVixDQUFBO0VBQ0YsSUFBQSxJQUFJbEIsWUFBWSxJQUFJQSxZQUFZLEtBQUssRUFBakMsSUFBdUNpQixlQUFlLENBQUNqQixZQUFoQixLQUFpQ0EsWUFBNUUsRUFDRWtCLE9BQU8sR0FBRyxLQUFWLENBQUE7RUFDRixJQUFBLElBQUlqQixXQUFXLElBQUlBLFdBQVcsS0FBSyxFQUEvQixJQUFxQ2dCLGVBQWUsQ0FBQ2hCLFdBQWhCLEtBQWdDQSxXQUF6RSxFQUNFaUIsT0FBTyxHQUFHLEtBQVYsQ0FBQTtFQUNGLElBQUEsSUFBSWhCLFFBQVEsSUFBSUEsUUFBUSxLQUFLLEVBQXpCLElBQStCZSxlQUFlLENBQUNmLFFBQWhCLEtBQTZCQSxRQUFoRSxFQUEwRWdCLE9BQU8sR0FBRyxLQUFWLENBQUE7RUFDMUUsSUFBQSxJQUFJZixNQUFNLElBQUlBLE1BQU0sS0FBSyxFQUFyQixJQUEyQmMsZUFBZSxDQUFDZCxNQUFoQixLQUEyQkEsTUFBMUQsRUFBa0VlLE9BQU8sR0FBRyxLQUFWLENBaktwRDs7TUFvS2QsSUFBSUosYUFBYSxJQUFJSSxPQUFqQixJQUE0QixDQUFDckIsVUFBN0IsSUFBMkNBLFVBQVUsS0FBSyxFQUE5RCxFQUFrRTtRQUNoRWxQLE9BQU8sQ0FBQytDLEdBQVIsQ0FDRSxNQURGLEVBRUUsOEJBQ0dvTixJQUFBQSxhQUFhLEdBQUcsS0FBSCxHQUFXLElBRDNCLENBRUUsR0FBQSxZQUZGLEdBR0VJLE9BSEYsR0FJRSxlQUpGLEdBS0VyQixVQUxGLEdBTUUsR0FSSixDQUFBLENBQUE7UUFVQWxQLE9BQU8sQ0FBQ2lCLGNBQVIsR0FBeUJxUCxlQUF6QixDQUFBO1FBQ0F0USxPQUFPLENBQUN3SCxTQUFSLENBQ0V4SCxPQUFPLENBQUNDLE1BQVIsQ0FBZUksU0FBZixHQUEyQixTQUQ3QixFQUVFa0MsSUFBSSxDQUFDQyxTQUFMLENBQWV4QyxPQUFPLENBQUNpQixjQUF2QixDQUZGLEVBR0VqQixPQUFPLENBQUNDLE1BQVIsQ0FBZUcsZUFIakIsQ0FBQSxDQUFBO0VBS0EsTUFBQSxPQUFBO0VBQ0QsS0F0TGE7OztNQXlMZEosT0FBTyxDQUFDK0IsZ0JBQVIsQ0FBeUI7RUFDdkJtTixNQUFBQSxVQUFVLEVBQUFBLFVBRGE7RUFFdkJFLE1BQUFBLFVBQVUsRUFBQUEsVUFGYTtFQUd2QkMsTUFBQUEsWUFBWSxFQUFBQSxZQUhXO0VBSXZCQyxNQUFBQSxXQUFXLEVBQUFBLFdBSlk7RUFLdkJDLE1BQUFBLFFBQVEsRUFBQUEsUUFMZTtFQU12QkMsTUFBQUEsTUFBTSxFQUFBQSxNQU5pQjtFQU92QkMsTUFBQUEsV0FBVyxFQUFBQSxXQUFBQTtPQVBiLENBQUEsQ0FBQTtLQTMzQ3NCO0VBczRDeEJULEVBQUFBLGNBQWMsRUFBRSxTQUFBLGNBQUEsR0FBQTtFQUNkaFAsSUFBQUEsT0FBTyxDQUFDK0MsR0FBUixDQUFZLE1BQVosRUFBb0IsdUJBQXBCLENBQUEsQ0FBQTtNQUNBL0MsT0FBTyxDQUFDc0IsVUFBUixDQUFtQm1CLG1CQUFuQixFQUFBLENBQUE7RUFDQXpDLElBQUFBLE9BQU8sQ0FBQzRJLFFBQVIsQ0FBaUIsSUFBakIsRUFIYztLQXQ0Q1E7RUE0NEN4QnFHLEVBQUFBLGFBQWEsRUFBRSxTQUFBLGFBQUEsR0FBQTtFQUNialAsSUFBQUEsT0FBTyxDQUFDK0MsR0FBUixDQUFZLE1BQVosRUFBb0Isc0JBQXBCLEVBRGE7O0VBSWIsSUFBQSxJQUFJLENBQUMvQyxPQUFPLENBQUNrQixlQUFiLEVBQThCO0VBQzVCLE1BQUEsT0FBQTtFQUNELEtBQUE7O0VBRURsQixJQUFBQSxPQUFPLENBQUNtQiwyQkFBUixHQUFzQyxJQUFJMEIsSUFBSixFQUF0QyxDQVJhO0tBNTRDUztFQXU1Q3hCMk4sRUFBQUEsV0FBVyxFQUFFLFNBQUEsV0FBQSxHQUFBOzs7TUFDWCxJQUFNQyxjQUFjLEdBQUcsQ0FBQUMsRUFBQSxHQUFBN0ssTUFBTSxDQUFDOEssSUFBUCxNQUFXLElBQVgsSUFBV0QsRUFBQSxLQUFBLEtBQUEsQ0FBWCxHQUFXLEtBQUEsQ0FBWCxHQUFXQSxFQUFBLENBQUVELGNBQXBDLENBQUE7O0VBQ0EsSUFBQSxJQUFJQSxjQUFKLEVBQW9CO0VBQ2xCLE1BQUEsSUFBTUcsUUFBUSxHQUFHLElBQUlILGNBQUosRUFBcUJJLENBQUFBLGVBQXJCLEdBQXVDQyxRQUF4RCxDQUFBOztFQUNBLE1BQUEsSUFBSUYsUUFBSixFQUFjO0VBQ1osUUFBQSxPQUFPQSxRQUFQLENBQUE7RUFDRCxPQUFBO0VBQ0YsS0FBQTs7RUFDRCxJQUFBLE9BQU83UCxTQUFQLENBQUE7S0EvNUNzQjtFQWs2Q3hCNkwsRUFBQUEsYUFBYSxFQUFFLFNBQUEsYUFBQSxDQUFDbUUsR0FBRCxFQUFjQyxJQUFkLEVBQTBCO01BQ3ZDLElBQUk7RUFDRixNQUFBLElBQUlDLFNBQVMsR0FBRyxJQUFJcEUsR0FBSixDQUFRa0UsR0FBUixDQUFoQixDQUFBO1FBQ0EsSUFBSUcsTUFBTSxHQUFHLElBQUlDLGVBQUosQ0FBb0JGLFNBQVMsQ0FBQ2pCLE1BQTlCLENBQWIsQ0FBQTtFQUNBLE1BQUEsT0FBT2tCLE1BQU0sQ0FBQ0UsR0FBUCxDQUFXSixJQUFYLEtBQW9CalEsU0FBM0IsQ0FBQTtPQUhGLENBSUUsT0FBT3NRLENBQVAsRUFBVTtFQUNWLE1BQUEsT0FBT3RRLFNBQVAsQ0FBQTtFQUNELEtBQUE7S0F6NkNxQjtFQTQ2Q3hCb08sRUFBQUEsWUFBWSxFQUFFLFNBQUEsWUFBQSxDQUFDbEksSUFBRCxFQUFlK0osSUFBZixFQUEyQjtFQUN2QyxJQUFBLElBQUlNLE9BQU8sR0FBR3JLLElBQUksQ0FBQ3NLLEtBQUwsQ0FBVyxJQUFJQyxNQUFKLENBQVdSLElBQUksR0FBRyxVQUFsQixDQUFYLENBQWQsQ0FBQTtFQUNBLElBQUEsT0FBT00sT0FBTyxHQUFHQSxPQUFPLENBQUMsQ0FBRCxDQUFWLEdBQWdCdlEsU0FBOUIsQ0FBQTtLQTk2Q3NCO0VBaTdDeEIwUSxFQUFBQSxjQUFjLEVBQUUsU0FBQ1YsY0FBQUEsQ0FBQUEsR0FBRCxFQUFjQyxJQUFkLEVBQTRCM0QsS0FBNUIsRUFBeUM7RUFDdkQsSUFBQSxJQUFJNEQsU0FBUyxHQUFHLElBQUlwRSxHQUFKLENBQVFrRSxHQUFSLENBQWhCLENBQUE7TUFDQSxJQUFJRyxNQUFNLEdBQUcsSUFBSUMsZUFBSixDQUFvQkYsU0FBUyxDQUFDakIsTUFBOUIsQ0FBYixDQUFBO0VBQ0FrQixJQUFBQSxNQUFNLENBQUM1TyxHQUFQLENBQVcwTyxJQUFYLEVBQWlCM0QsS0FBakIsQ0FBQSxDQUFBO0VBQ0E0RCxJQUFBQSxTQUFTLENBQUNqQixNQUFWLEdBQW1Ca0IsTUFBTSxDQUFDUSxRQUFQLEVBQW5CLENBQUE7TUFDQSxPQUFPVCxTQUFTLENBQUNTLFFBQVYsRUFBUCxDQUFBO0tBdDdDc0I7RUF5N0N4QnJJLEVBQUFBLFlBQVksRUFBRSxTQUFBLFlBQUEsR0FBQTtFQUNaLElBQUEsSUFBTXNJLEdBQUcsR0FBR25OLFFBQVEsQ0FBQ21MLGFBQVQsQ0FBdUIsS0FBdkIsQ0FBWixDQUFBO01BQ0FnQyxHQUFHLENBQUNDLFNBQUosR0FBZ0IsUUFBaEIsQ0FBQTtNQUNBRCxHQUFHLENBQUNFLFNBQUosR0FBZ0IsUUFBaEIsQ0FBQTtNQUNBLElBQUlDLE9BQU8sR0FBRyxLQUFkLENBQUE7O01BQ0EsSUFBSTtFQUNGO0VBQ0F0TixNQUFBQSxRQUFRLENBQUNrSCxJQUFULENBQWNxRyxXQUFkLENBQTBCSixHQUExQixDQUFBLENBQUE7UUFDQUcsT0FBTyxHQUFJdE4sUUFBUSxDQUFDd04sc0JBQVQsQ0FBZ0MsUUFBaEMsQ0FBQSxDQUEwQyxDQUExQyxDQUFBLENBQTZEQyxZQUE3RCxLQUE4RSxDQUF6RixDQUFBO0VBQ0F6TixNQUFBQSxRQUFRLENBQUNrSCxJQUFULENBQWN3RyxXQUFkLENBQTBCUCxHQUExQixDQUFBLENBQUE7T0FKRixDQUtFLE9BQU9RLEVBQVAsRUFBVztFQUNYTCxNQUFBQSxPQUFPLEdBQUcsS0FBVixDQUFBO0VBQ0QsS0FBQTs7RUFDRCxJQUFBLE9BQU9BLE9BQVAsQ0FBQTtLQXQ4Q3NCO0VBeThDeEIxRCxFQUFBQSxjQUFjLEVBQUUsU0FBQSxjQUFBLEdBQUE7RUFDZDtFQUNBLElBQUEsSUFBSWdFLEVBQUUsR0FBR25KLFNBQVMsQ0FBQ0MsU0FBVixDQUFvQm1KLFdBQXBCLEVBQVQsQ0FBQTs7TUFDQSxJQUFJRCxFQUFFLENBQUMzRixPQUFILENBQVcsTUFBWCxDQUF1QixLQUFBLENBQUMsQ0FBNUIsRUFBK0I7RUFDN0IsTUFBQSxJQUFJbUIsUUFBUSxDQUFDd0UsRUFBRSxDQUFDRSxLQUFILENBQVMsTUFBVCxDQUFpQixDQUFBLENBQWpCLENBQUQsRUFBc0IsRUFBdEIsQ0FBUixJQUFxQyxDQUF6QyxFQUE0QztFQUMxQyxRQUFBLE9BQU8sS0FBUCxDQUFBO0VBQ0QsT0FBQTtFQUNGLEtBUGE7OztFQVVkLElBQUEsSUFDRSw4SUFBOElDLElBQTlJLENBQ0V0SixTQUFTLENBQUNDLFNBRFosQ0FERixFQUlFO0VBQ0EsTUFBQSxPQUFPLEtBQVAsQ0FBQTtFQUNELEtBaEJhOzs7TUFtQmQsSUFBSUQsU0FBUyxDQUFDdUosU0FBZCxFQUF5QjtFQUN2QixNQUFBLE9BQU8sS0FBUCxDQUFBO0VBQ0QsS0FBQTs7RUFDRCxJQUFBLE9BQU8sSUFBUCxDQUFBO0tBLzlDc0I7RUFrK0N4QmpOLEVBQUFBLE1BQU0sRUFBRSxTQUFBLE1BQUEsR0FBQTtFQUNOLElBQUEsT0FBTyx1Q0FBdUNTLE9BQXZDLENBQStDLE9BQS9DLEVBQXdELFVBQVV5TSxDQUFWLEVBQVc7UUFDeEUsSUFBSUMsQ0FBQyxHQUFJL1AsSUFBSSxDQUFDZ1EsTUFBTCxFQUFnQixHQUFBLEVBQWpCLEdBQXVCLENBQS9CO0VBQUEsVUFDRUMsQ0FBQyxHQUFHSCxDQUFDLElBQUksR0FBTCxHQUFXQyxDQUFYLEdBQWdCQSxDQUFDLEdBQUcsR0FBTCxHQUFZLEdBRGpDLENBQUE7RUFFQSxNQUFBLE9BQU9FLENBQUMsQ0FBQ2xCLFFBQUYsQ0FBVyxFQUFYLENBQVAsQ0FBQTtFQUNELEtBSk0sQ0FBUCxDQUFBO0tBbitDc0I7SUEwK0N4Qm1CLEdBQUcsRUFBRSxTQUFDQyxHQUFBQSxDQUFBQSxHQUFELEVBQVk7TUFDZixPQUFPRCxHQUFHLENBQUNDLEdBQUQsQ0FBVixDQUFBO0tBMytDc0I7RUE4K0N4QnBOLEVBQUFBLFdBQVcsRUFBRSxTQUFBLFdBQUEsR0FBQTtNQUNYLElBQUlELFFBQVEsR0FBRzFFLFNBQWYsQ0FBQTs7TUFDQSxJQUFJO0VBQ0YwRSxNQUFBQSxRQUFRLEdBQUdJLE1BQU0sQ0FBQ2tOLEdBQVAsQ0FBV3ZPLFFBQVgsQ0FBb0JpQixRQUFwQixLQUFpQyxFQUFqQyxHQUFzQ0ksTUFBTSxDQUFDa04sR0FBUCxDQUFXdk8sUUFBWCxDQUFvQmlCLFFBQTFELEdBQXFFMUUsU0FBaEYsQ0FBQTtPQURGLENBRUUsT0FBT3NRLENBQVAsRUFBVTtRQUNWLElBQUl4TCxNQUFNLENBQUNtTixNQUFYLEVBQW1CO1VBQ2pCLElBQUk7RUFDRnZOLFVBQUFBLFFBQVEsR0FDTkksTUFBTSxDQUFDbU4sTUFBUCxDQUFjeE8sUUFBZCxDQUF1QmlCLFFBQXZCLEtBQW9DLEVBQXBDLEdBQXlDSSxNQUFNLENBQUNtTixNQUFQLENBQWN4TyxRQUFkLENBQXVCaUIsUUFBaEUsR0FBMkUxRSxTQUQ3RSxDQUFBO1dBREYsQ0FHRSxPQUFPb1IsRUFBUCxFQUFXO0VBQ1gxTSxVQUFBQSxRQUFRLEdBQUcxRSxTQUFYLENBQUE7RUFDRCxTQUFBO0VBQ0YsT0FBQTtFQUNGLEtBQUE7O01BQ0QsSUFBSSxDQUFDMEUsUUFBTCxFQUFlO1FBQ2JBLFFBQVEsR0FBR2pCLFFBQVEsQ0FBQ2lCLFFBQVQsS0FBc0IsRUFBdEIsR0FBMkJqQixRQUFRLENBQUNpQixRQUFwQyxHQUErQzFFLFNBQTFELENBQUE7RUFDRCxLQUFBOztFQUNELElBQUEsT0FBTzBFLFFBQVAsQ0FBQTtLQS8vQ3NCO0lBa2dEeEIxRCxnQkFBZ0IsRUFBRSxTQUFDbVAsZ0JBQUFBLENBQUFBLE1BQUQsRUFBb0I7RUFDcEM7TUFDQSxJQUFJLENBQUNBLE1BQU0sQ0FBQ2hDLFVBQVIsSUFBc0JnQyxNQUFNLENBQUNoQyxVQUFQLEtBQXNCLEVBQWhELEVBQW9EO1FBQ2xEZ0MsTUFBTSxDQUFDaEMsVUFBUCxHQUFvQixRQUFwQixDQUFBO0VBQ0QsS0FBQTs7TUFDRCxJQUFJLENBQUNnQyxNQUFNLENBQUM5QixVQUFSLElBQXNCOEIsTUFBTSxDQUFDOUIsVUFBUCxLQUFzQixFQUFoRCxFQUFvRDtRQUNsRDhCLE1BQU0sQ0FBQzlCLFVBQVAsR0FBb0IsTUFBcEIsQ0FBQTtFQUNELEtBQUE7O01BRURwUCxPQUFPLENBQUNpQixjQUFSLEdBQXlCO0VBQ3ZCcUUsTUFBQUEsV0FBVyxFQUFFdEYsT0FBTyxDQUFDdUYsTUFBUixFQURVO0VBRXZCQyxNQUFBQSxVQUFVLEVBQUUsSUFBSTNDLElBQUosRUFBQSxDQUFXSyxXQUFYLEVBRlc7RUFHdkI1RCxNQUFBQSxrQkFBa0IsRUFBRVUsT0FBTyxDQUFDZ0IsYUFBUixDQUFzQnNFLFdBSG5CO0VBSXZCMk4sTUFBQUEsWUFBWSxFQUFFcE4sTUFBTSxDQUFDQyxRQUFQLENBQWdCQyxJQUpQO0VBS3ZCTixNQUFBQSxRQUFRLEVBQUV6RixPQUFPLENBQUMwRixXQUFSLEVBTGE7RUFNdkJrTCxNQUFBQSxRQUFRLEVBQUU1USxPQUFPLENBQUN3USxXQUFSLEVBTmE7UUFRdkJ0QixVQUFVLEVBQUVnQyxNQUFNLENBQUNoQyxVQVJJO1FBU3ZCRSxVQUFVLEVBQUU4QixNQUFNLENBQUM5QixVQVRJO1FBVXZCQyxZQUFZLEVBQUU2QixNQUFNLENBQUM3QixZQVZFO1FBV3ZCQyxXQUFXLEVBQUU0QixNQUFNLENBQUM1QixXQVhHO1FBWXZCQyxRQUFRLEVBQUUyQixNQUFNLENBQUMzQixRQVpNO1FBYXZCQyxNQUFNLEVBQUUwQixNQUFNLENBQUMxQixNQWJRO1FBY3ZCQyxXQUFXLEVBQUV5QixNQUFNLENBQUN6QixXQWRHO0VBZ0J2QnRNLE1BQUFBLFFBQVEsRUFBRSxDQWhCYTtFQWlCdkJrRCxNQUFBQSxlQUFlLEVBQUUsQ0FqQk07RUFrQnZCQyxNQUFBQSxrQkFBa0IsRUFBRSxDQUFBO09BbEJ0QixDQUFBO01BcUJBdEcsT0FBTyxDQUFDK0MsR0FBUixDQUFZLE1BQVosRUFBb0IscUJBQXBCLEVBQTJDL0MsT0FBTyxDQUFDaUIsY0FBbkQsQ0FBQSxDQTlCb0M7O01BaUNwQ2pCLE9BQU8sQ0FBQ3dILFNBQVIsQ0FDRXhILE9BQU8sQ0FBQ0MsTUFBUixDQUFlSSxTQUFmLEdBQTJCLFNBRDdCLEVBRUVrQyxJQUFJLENBQUNDLFNBQUwsQ0FBZXhDLE9BQU8sQ0FBQ2lCLGNBQXZCLENBRkYsRUFHRWpCLE9BQU8sQ0FBQ0MsTUFBUixDQUFlRyxlQUhqQixDQUFBLENBQUE7S0FuaURzQjtJQTBpRHhCd0IsU0FBUyxFQUFFLFNBQUNvUCxTQUFBQSxDQUFBQSxJQUFELEVBQWE7RUFDdEIsSUFBQSxPQUNFa0Msa0JBQWtCLENBQ2hCMU8sUUFBUSxDQUFDMk8sTUFBVCxDQUFnQm5OLE9BQWhCLENBQ0UsSUFBSXdMLE1BQUosQ0FDRSxrQkFBQSxHQUNFNEIsa0JBQWtCLENBQUNwQyxJQUFELENBQWxCLENBQXlCaEwsT0FBekIsQ0FBaUMsU0FBakMsRUFBNEMsTUFBNUMsQ0FERixHQUVFLDZCQUhKLENBREYsRUFNRSxJQU5GLENBRGdCLENBQWxCLElBU0ssSUFWUCxDQUFBO0tBM2lEc0I7RUF5akR4QjtFQUNBd0IsRUFBQUEsU0FBUyxFQUFFLFNBQUN3SixTQUFBQSxDQUFBQSxJQUFELEVBQWUzRCxLQUFmLEVBQThCZ0csT0FBOUIsRUFBNkM7RUFDdEQ7TUFDQSxJQUFNL0IsT0FBTyxHQUFHekwsTUFBTSxDQUFDQyxRQUFQLENBQWdCK0osUUFBaEIsQ0FBeUIwQixLQUF6QixDQUErQixxQ0FBL0IsQ0FBaEIsQ0FBQTtNQUNBLElBQU0rQixNQUFNLEdBQUdoQyxPQUFPLEdBQUdBLE9BQU8sQ0FBQyxDQUFELENBQVYsR0FBZ0IsRUFBdEMsQ0FBQTtFQUNBLElBQUEsSUFBTWlDLE9BQU8sR0FBR0QsTUFBTSxHQUFHLFlBQWVBLEdBQUFBLE1BQWxCLEdBQTJCLEVBQWpELENBQUE7RUFFQSxJQUFBLElBQU1yUSxHQUFHLEdBQUcsSUFBSUosSUFBSixFQUFaLENBQUE7TUFDQUksR0FBRyxDQUFDdVEsT0FBSixDQUFZdlEsR0FBRyxDQUFDSCxPQUFKLEVBQUEsR0FBZ0J1USxPQUFPLEdBQUcsSUFBdEMsQ0FBQSxDQUFBO0VBQ0EsSUFBQSxJQUFNSSxPQUFPLEdBQUcsWUFBQSxHQUFleFEsR0FBRyxDQUFDeVEsV0FBSixFQUEvQixDQUFBO0VBRUEsSUFBQSxJQUFNQyxZQUFZLEdBQ2hCM0MsSUFBSSxHQUFHLEdBQVAsR0FBYW9DLGtCQUFrQixDQUFDL0YsS0FBRCxDQUEvQixHQUF5Q29HLE9BQXpDLEdBQW1ELFVBQW5ELEdBQWdFRixPQUFoRSxHQUEwRSxVQUQ1RSxDQUFBO01BRUEvTyxRQUFRLENBQUMyTyxNQUFULEdBQWtCUSxZQUFsQixDQUFBO0VBQ0EsSUFBQSxPQUFBO0tBdmtEc0I7SUEwa0R4QjVMLFlBQVksRUFBRSxTQUFDaUosWUFBQUEsQ0FBQUEsSUFBRCxFQUFhO01BQ3pCaFIsT0FBTyxDQUFDd0gsU0FBUixDQUFrQndKLElBQWxCLEVBQXdCLEVBQXhCLEVBQTRCLENBQUMsQ0FBN0IsQ0FBQSxDQUFBO0tBM2tEc0I7RUE4a0R4QjNPLEVBQUFBLGFBQWEsRUFBRTtNQUNiK08sR0FBRyxFQUFFLFNBQUNqSixHQUFBQSxDQUFBQSxHQUFELEVBQVk7UUFDZixPQUFPa0csWUFBWSxDQUFDQyxPQUFiLENBQXFCdE8sT0FBTyxDQUFDQyxNQUFSLENBQWVJLFNBQWYsR0FBMkI4SCxHQUFoRCxDQUFQLENBQUE7T0FGVztFQUliN0YsSUFBQUEsR0FBRyxFQUFFLFNBQUEsR0FBQSxDQUFDNkYsR0FBRCxFQUFja0YsS0FBZCxFQUEyQjtRQUM5QixJQUFJO1VBQ0ZnQixZQUFZLENBQUN1RixPQUFiLENBQXFCNVQsT0FBTyxDQUFDQyxNQUFSLENBQWVJLFNBQWYsR0FBMkI4SCxHQUFoRCxFQUFxRGtGLEtBQXJELENBQUEsQ0FBQTtTQURGLENBRUUsT0FBT2dFLENBQVAsRUFBVTtFQUNWclIsUUFBQUEsT0FBTyxDQUFDK0MsR0FBUixDQUFZLE9BQVosRUFBcUIscUJBQXJCLEVBQTRDc08sQ0FBNUMsQ0FBQSxDQUFBO0VBQ0QsT0FBQTtPQVRVO01BV2J4SixNQUFNLEVBQUUsU0FBQ00sTUFBQUEsQ0FBQUEsR0FBRCxFQUFZO1FBQ2xCa0csWUFBWSxDQUFDd0YsVUFBYixDQUF3QjdULE9BQU8sQ0FBQ0MsTUFBUixDQUFlSSxTQUFmLEdBQTJCOEgsR0FBbkQsQ0FBQSxDQUFBO0VBQ0QsS0FBQTtLQTNsRHFCO0VBOGxEeEI7SUFDQXlHLFlBQVksRUFBRSxTQUFDeUMsWUFBQUEsQ0FBQUEsQ0FBRCxFQUFjO0VBQzFCLElBQUEsSUFBTXlDLE1BQU0sR0FBR3pDLENBQUMsQ0FBQ3lDLE1BQWpCLENBQUE7TUFDQUEsTUFBTSxDQUFDL04sSUFBUCxHQUFjL0YsT0FBTyxDQUFDeVIsY0FBUixDQUNacUMsTUFBTSxDQUFDL04sSUFESyxFQUVaMUcsU0FBUyxDQUFDQyxrQkFGRSxFQUdaVSxPQUFPLENBQUNnQixhQUFSLENBQXNCc0UsV0FIVixDQUFkLENBQUE7TUFLQXdPLE1BQU0sQ0FBQy9OLElBQVAsR0FBYy9GLE9BQU8sQ0FBQ3lSLGNBQVIsQ0FDWnFDLE1BQU0sQ0FBQy9OLElBREssRUFFWjFHLFNBQVMsQ0FBQ0UsZ0JBRkUsRUFHWlMsT0FBTyxDQUFDYyxXQUFSLENBQW9Cd0UsV0FIUixDQUFkLENBQUE7TUFLQXdPLE1BQU0sQ0FBQy9OLElBQVAsR0FBYy9GLE9BQU8sQ0FBQ3lSLGNBQVIsQ0FDWnFDLE1BQU0sQ0FBQy9OLElBREssRUFFWjFHLFNBQVMsQ0FBQ0cscUJBRkUsRUFHWlEsT0FBTyxDQUFDYyxXQUFSLENBQW9COEcsZ0JBQXBCLENBQXFDOEosUUFBckMsRUFIWSxDQUFkLENBQUE7O0VBS0EsSUFBQSxJQUFJMVIsT0FBTyxDQUFDYyxXQUFSLENBQW9CNEwsSUFBeEIsRUFBOEI7UUFDNUJvSCxNQUFNLENBQUMvTixJQUFQLEdBQWMvRixPQUFPLENBQUN5UixjQUFSLENBQ1pxQyxNQUFNLENBQUMvTixJQURLLEVBRVoxRyxTQUFTLENBQUNJLHFCQUZFLEVBR1pPLE9BQU8sQ0FBQ2MsV0FBUixDQUFvQjRMLElBSFIsQ0FBZCxDQUFBO0VBS0QsS0FBQTtLQXRuRHFCO0VBeW5EeEI7SUFDQXhGLFNBQVMsRUFBRSxTQUFDeEYsU0FBQUEsQ0FBQUEsSUFBRCxFQUFZO01BQ3JCLElBQUlxUyxRQUFRLEdBQUdyUyxJQUFJLENBQUNzUyxVQUFMLEdBQWtCdFMsSUFBSSxDQUFDc1MsVUFBdkIsR0FBb0MsRUFBbkQsQ0FBQTs7TUFFQSxJQUFJdFMsSUFBSSxDQUFDSCxLQUFMLElBQWNHLElBQUksQ0FBQ0gsS0FBTCxDQUFXdUMsTUFBWCxHQUFvQixDQUF0QyxFQUF5QztFQUN2Q3BDLE1BQUFBLElBQUksQ0FBQ0gsS0FBTCxDQUFXdUYsT0FBWCxDQUFtQixVQUFDOUUsSUFBRCxFQUFLO1VBQ3RCK1IsUUFBUSxHQUNOQSxRQUFRLEdBQ1IvUixJQUFJLENBQUNpUyxtQkFETCxJQUVDalMsSUFBSSxDQUFDa1MsbUJBQUwsR0FBMkJsUyxJQUFJLENBQUNrUyxtQkFBaEMsR0FBc0QsRUFGdkQsQ0FBQSxJQUdDbFMsSUFBSSxDQUFDbVMsUUFBTCxJQUFpQixHQUhsQixDQURGLENBQUE7U0FERixDQUFBLENBQUE7RUFPRCxLQUFBOztNQUVELE9BQU90QixHQUFHLENBQUNrQixRQUFELENBQVYsQ0FBQTtLQXZvRHNCO0VBMG9EeEJLLEVBQUFBLFFBQVEsRUFBRSxTQUFBLFFBQUEsR0FBQTtFQUNSO0VBQ0EsSUFBQSxJQUFJdk8sTUFBTSxDQUFDd08sT0FBUCxDQUFlLGlDQUFmLENBQUosRUFBdUQ7RUFDckQ7UUFDQXJVLE9BQU8sQ0FBQytILFlBQVIsQ0FBcUIvSCxPQUFPLENBQUNDLE1BQVIsQ0FBZUksU0FBZixHQUEyQixRQUFoRCxDQUFBLENBQUE7UUFDQUwsT0FBTyxDQUFDK0gsWUFBUixDQUFxQi9ILE9BQU8sQ0FBQ0MsTUFBUixDQUFlSSxTQUFmLEdBQTJCLE1BQWhELENBQUEsQ0FBQTtRQUNBTCxPQUFPLENBQUMrSCxZQUFSLENBQXFCL0gsT0FBTyxDQUFDQyxNQUFSLENBQWVJLFNBQWYsR0FBMkIsU0FBaEQsQ0FBQSxDQUpxRDs7RUFNckRMLE1BQUFBLE9BQU8sQ0FBQ3FDLGFBQVIsQ0FBc0J3RixNQUF0QixDQUE2QixPQUE3QixDQUFBLENBQUE7O0VBQ0E3SCxNQUFBQSxPQUFPLENBQUNxQyxhQUFSLENBQXNCd0YsTUFBdEIsQ0FBNkIsZUFBN0IsRUFQcUQ7OztRQVNyRDdILE9BQU8sQ0FBQ2MsV0FBUixHQUFzQkMsU0FBdEIsQ0FBQTtRQUNBZixPQUFPLENBQUNnQixhQUFSLEdBQXdCRCxTQUF4QixDQUFBO1FBQ0FmLE9BQU8sQ0FBQ2lCLGNBQVIsR0FBeUJGLFNBQXpCLENBQUE7UUFDQWYsT0FBTyxDQUFDcUIsV0FBUixHQUFzQk4sU0FBdEIsQ0FBQTtRQUNBZixPQUFPLENBQUNrQixlQUFSLEdBQTBCSCxTQUExQixDQUFBO1FBQ0FmLE9BQU8sQ0FBQ1ksT0FBUixHQUFrQixLQUFsQixDQUFBOztFQUNBWixNQUFBQSxPQUFPLENBQUMwRSxRQUFSLENBQWlCMUUsT0FBTyxDQUFDQyxNQUF6QixDQUFBLENBQUE7RUFDRCxLQUFBO0VBQ0YsR0FBQTtFQTdwRHVCOzs7Ozs7OzsifQ==
