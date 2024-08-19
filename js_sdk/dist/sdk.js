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
  var URLParams={device_external_id:"_did",user_authenticated_external_id:"_authuid",user_external_id:"_uid",user_is_authenticated:"_auth",user_external_id_hmac:"_uidh"},OneYearInSeconds=31536e3,PageStates={active:"active",passive:"passive",hidden:"hidden",frozen:"frozen"},Rimdian={config:{workspace_id:"",host:"https://collector-eu.rimdian.com",session_timeout:1800,namespace:"_rmd_",cross_domains:[],ignored_origins:[],version:"2.11.0",log_level:"error",max_retry:10,from_cm:!1},isReady:!1,dispatchConsent:!1,currentUser:void 0,currentDevice:void 0,currentSession:void 0,currentPageview:void 0,currentPageviewVisibleSince:void 0,currentPageviewDuration:0,currentCart:void 0,itemsQueue:{items:[],add:function add(a,b){// if current session expired, and received an interaction event, create a new session
  var c=Rimdian.getCookie(Rimdian.config.namespace+"session");!c&&(["pageview","cart","order"].includes(a)||"custom_event"===a&&!0===b.non_interactive)&&Rimdian._startNewSession({});var d={kind:a};// persist items in local storage
  d[a]=b,["pageview","custom_event","cart","order"].includes(a)&&(d.user=__assign({},Rimdian.currentUser),d.session=__assign({},Rimdian.currentSession)),Rimdian.itemsQueue.items.push(d),Rimdian._localStorage.set("items",JSON.stringify(Rimdian.itemsQueue.items));},// addPageviewDuration() is called when page becomes passive/not focused
  // pageview duration should not trigger a new session if the session expired
  // that's why we handle it separately
  addPageviewDuration:function addPageviewDuration(){// abort if we are not tracking the current pageview
  if(Rimdian.currentPageview&&Rimdian.currentPageviewVisibleSince){// increment the time spent
  var a=Math.round((new Date().getTime()-Rimdian.currentPageviewVisibleSince.getTime())/1e3);Rimdian.currentPageviewDuration+=a,Rimdian.log("info","time spent on page is now",Rimdian.currentPageviewDuration);// add pageview to items queue
  var b=__assign({},Rimdian.currentPageview),c=new Date().toISOString();b.duration=Rimdian.currentPageviewDuration,b.updated_at=c,Rimdian.currentSession.duration||(Rimdian.currentSession.duration=0),Rimdian.currentSession.duration+=a,Rimdian.currentSession.updated_at=c,Rimdian.setSessionContext(Rimdian.currentSession);// cookie update
  // error on invalid pageview
  // if (!ValidatePageview(pageview)) {
  //   Rimdian.log('error', 'invalid pageview', ValidatePageview.errors)
  // }
  var d={kind:"pageview",pageview:b,user:__assign({},Rimdian.currentUser),session:__assign({},Rimdian.currentSession),device:__assign({},Rimdian.currentDevice)};Rimdian.itemsQueue.items.push(d),Rimdian._localStorage.set("items",JSON.stringify(Rimdian.itemsQueue.items));}}},dispatchQueue:[],isDispatching:!1,onReadyQueue:[],log:function log(a){for(var b=[],c=1;c<arguments.length;c++)b[c-1]=arguments[c];"warn"===a?["warn","info","debug","trace"].includes(Rimdian.config.log_level)&&console.warn.apply(console,b):"info"===a?["info","debug","trace"].includes(Rimdian.config.log_level)&&console.info.apply(console,b):"debug"===a?["debug","trace"].includes(Rimdian.config.log_level)&&console.debug.apply(console,b):"trace"===a?"trace"===Rimdian.config.log_level&&console.trace.apply(console,b):console.error.apply(console,b);},// watch for DOM to be ready and call onReady()
  init:function init(a){("complete"===document.readyState||"interactive"===document.readyState)&&Rimdian._onReady(a),document.onreadystatechange=function(){("interactive"===document.readyState||"complete"===document.readyState)&&(Rimdian.log("info","document is now",document.readyState),Rimdian._onReady(a));};var b=Rimdian.getCookie(Rimdian.config.namespace+"debug");b&&(Rimdian.config.log_level="info");},setDispatchConsent:function setDispatchConsent(a){Rimdian.log("info","RMD dispatch consent is now",a),Rimdian.dispatchConsent=a;},// return callback when the user is ready
  getCurrentUser:function getCurrentUser(a){// return callback when the user is ready
  return Rimdian.currentUser?void a(Rimdian.currentUser):void Rimdian._execWhenReady(function(){a(Rimdian.currentUser);})},onReady:function onReady(a){Rimdian.isReady?a():Rimdian.onReadyQueue.push(a);},// tracks the current pageview
  trackPageview:function trackPageview(a){if(!Rimdian.isReady)return Rimdian.log("debug","RMD is not yet ready for trackPageview, queuing function..."),void Rimdian._execWhenReady(function(){return Rimdian.trackPageview(a)});// persist existing pageview duration if any, or discard it if not active
  // because the pageview duration is already persisted when pages become passive
  Rimdian.currentPageview&&!Rimdian.currentPageviewVisibleSince&&(Rimdian.log("info","previous pageview has been discarded, was not active"),Rimdian.currentPageview=void 0);var b=new Date().toISOString();// check if a pageview is already tracked (Single Page App)
  if(Rimdian.currentPageview){Rimdian.log("info","pageview already tracked");// compute pageview duration
  var c=Math.round((new Date().getTime()-Rimdian.currentPageviewVisibleSince.getTime())/1e3);// session is persisted in cookie below
  // add current pageview to items queue
  // reset timer
  Rimdian.currentPageviewDuration+=c,Rimdian.currentPageview.duration=Rimdian.currentPageviewDuration,Rimdian.currentPageview.updated_at=b,Rimdian.currentSession.duration||(Rimdian.currentSession.duration=0),Rimdian.currentSession.duration+=c,Rimdian.currentSession.updated_at=b,Rimdian.itemsQueue.add("pageview",__assign({},Rimdian.currentPageview)),Rimdian.currentPageviewVisibleSince=void 0,Rimdian.currentPageviewDuration=0;}// deep clone pageview
  var d=JSON.parse(JSON.stringify(a||{}));d.external_id=Rimdian.uuidv4(),d.created_at=b;var e=Rimdian.getReferrer();// escape title for JSON
  // // error on invalid pageview
  // if (!ValidatePageview(pageview)) {
  //   Rimdian.log('error', 'invalid pageview', ValidatePageview.errors)
  // }
  // set current pageview
  // update user last interaction
  // cookie update
  // enqueue pageview
  e&&(d.referrer=e),d.title||(d.title=document.title),d.page_id||(d.page_id=window.location.href),d.title=d.title.replace(/\\"/,"\""),d.product_price&&0<d.product_price&&(d.product_price=Math.round(100*d.product_price)),Rimdian.isPageVisible()&&(Rimdian.currentPageviewVisibleSince=new Date),Rimdian.currentPageview=d,Rimdian.currentUser.last_interaction_at=b,Rimdian.setUserContext(Rimdian.currentUser),Rimdian.currentSession.pageviews_count||(Rimdian.currentSession.pageviews_count=0),Rimdian.currentSession.interactions_count||(Rimdian.currentSession.interactions_count=0),Rimdian.currentSession.pageviews_count++,Rimdian.currentSession.interactions_count++,Rimdian.currentSession.updated_at=b,Rimdian.setSessionContext(Rimdian.currentSession),Rimdian.itemsQueue.add("pageview",__assign({},d));},// tracks the current customEvent
  trackCustomEvent:function trackCustomEvent(a){if(!Rimdian.isReady)return Rimdian.log("debug","RMD is not yet ready for trackCustomEvent, queuing function..."),void Rimdian._execWhenReady(function(){return Rimdian.trackCustomEvent(a)});if(a&&!a.label)return void Rimdian.log("error","customEvent label is required");// deep clone customEvent
  var b=JSON.parse(JSON.stringify(a||{})),c=new Date().toISOString();// escape label for JSON
  // enqueue customEvent
  b.external_id=Rimdian.uuidv4(),b.created_at=c,b.label=b.label.replace(/\\"/,"\""),b.string_value&&(b.string_value=b.string_value.replace(/\\"/,"\"")),b.non_interactive||(Rimdian.currentUser.last_interaction_at=c,Rimdian.setUserContext(Rimdian.currentUser),!Rimdian.currentSession.interactions_count&&(Rimdian.currentSession.interactions_count=0),Rimdian.currentSession.interactions_count++,Rimdian.currentSession.updated_at=c,Rimdian.setSessionContext(Rimdian.currentSession)),Rimdian.itemsQueue.add("custom_event",__assign({},b));},trackCart:function trackCart(a){if(!Rimdian.isReady)return Rimdian.log("debug","RMD is not yet ready for trackCart, queuing function..."),void Rimdian._execWhenReady(function(){return Rimdian.trackCart(a)});// check if data exists and is an object
  if(!a||"object"!==_typeof(a))return void Rimdian.log("error","invalid cart data");// deep clone cart
  var b=JSON.parse(JSON.stringify(a));// check if a cart is already tracked (Single Page App)
  if(Rimdian.log("info","cart is",b),b.external_id||(b.external_id=Rimdian.uuidv4()),b.session_external_id=Rimdian.currentSession.external_id,b.items&&b.items.forEach(function(a){a.cart_external_id=b.external_id,a.price&&0<a.price&&(a.price=Math.round(100*a.price));}),b.hash||(b.hash=Rimdian._cartHash(b)),!(Rimdian.currentCart&&(Rimdian.log("info","cart already tracked",__assign({},Rimdian.currentCart)),Rimdian.currentCart.hash===b.hash)))// skip if cart hash is the same
  {var c=new Date().toISOString();b.updated_at||(b.updated_at=new Date().toISOString()),Rimdian.currentCart=b,Rimdian.currentUser.last_interaction_at=c,Rimdian.currentSession.interactions_count||(Rimdian.currentSession.interactions_count=0),Rimdian.currentSession.interactions_count++,Rimdian.currentSession.updated_at=c,Rimdian.setSessionContext(Rimdian.currentSession),Rimdian.itemsQueue.add("cart",b);}},trackOrder:function trackOrder(a){if(!Rimdian.isReady)return Rimdian.log("debug","RMD is not yet ready for trackOrder, queuing function..."),void Rimdian._execWhenReady(function(){return Rimdian.trackOrder(a)});// check if data exists and is an object
  if(!a||"object"!==_typeof(a))return void Rimdian.log("error","invalid order data");// deep clone order
  var b=JSON.parse(JSON.stringify(a));// convert item prices to cents
  b.items&&b.items.forEach(function(a){a.price&&0<a.price&&(a.price=Math.round(100*a.price)),a.order_external_id=b.external_id;});// // error on invalid order, but don't abort
  // if (!ValidateOrder(order)) {
  //   Rimdian.log('error', 'invalid order', ValidateOrder.errors)
  // }
  var c=new Date().toISOString();// update user last interaction
  // cookie update
  // enqueue order after user+session update
  Rimdian.currentUser.last_interaction_at=c,Rimdian.setUserContext(Rimdian.currentUser),Rimdian.currentSession.interactions_count||(Rimdian.currentSession.interactions_count=0),Rimdian.currentSession.interactions_count++,Rimdian.currentSession.updated_at=c,Rimdian.setSessionContext(Rimdian.currentSession),Rimdian.itemsQueue.add("order",b);},setDeviceContext:function setDeviceContext(a){if(!Rimdian.isReady)return Rimdian.log("debug","RMD is not yet ready for setDeviceContext, queuing function..."),void Rimdian._execWhenReady(function(){return Rimdian.setDeviceContext(a)});// merge currentDevice with data
  var b=__assign(__assign({},Rimdian.currentDevice),a);// enriching the device data should trigger a DB update with updated_at
  // // validate current device
  // if (!ValidateDevice(Rimdian.currentDevice)) {
  //   Rimdian.log('error', 'invalid device', ValidateDevice.errors)
  //   return
  // }
  // persist updated device
  b.updated_at=new Date().toISOString(),Rimdian.currentDevice=b,Rimdian.setCookie(Rimdian.config.namespace+"device",JSON.stringify(Rimdian.currentDevice),OneYearInSeconds);},setSessionContext:function setSessionContext(a){if(!Rimdian.isReady)return Rimdian.log("debug","RMD is not yet ready for setSessionContext, queuing function..."),void Rimdian._execWhenReady(function(){return Rimdian.setSessionContext(a)});// merge currentSession with data
  var b=__assign(__assign({},Rimdian.currentSession),a);// enriching the device data should trigger a DB update with updated_at
  // persist updated session
  b.updated_at=new Date().toISOString(),Rimdian.currentSession=b,Rimdian.log("info","RMD updated session is",b),Rimdian.setCookie(Rimdian.config.namespace+"session",JSON.stringify(Rimdian.currentSession),Rimdian.config.session_timeout);},// check if user ID has changed
  // reset context if new user is also authenticated
  // or user_alias if previous was anonymous
  // set new user fields and eventually enqueue user
  setUserContext:function setUserContext(a){if(!Rimdian.isReady)return Rimdian.log("debug","RMD is not yet ready for setUserContext, queuing function..."),void Rimdian._execWhenReady(function(){return Rimdian.setUserContext(a)});// replace user_centric_consent by consent_all
  if(a&&void 0!==a.user_centric_consent&&(a.consent_all=a.user_centric_consent,delete a.user_centric_consent),a&&a.external_id&&""!==a.external_id&&a.external_id!==Rimdian.currentUser.external_id){// if previous user was authenticated, reset new session / device, as we can't alias 2 authenticated users
  if(Rimdian.currentUser.is_authenticated&&!0===a.is_authenticated)// // validate current user
  // if (!ValidateUser(Rimdian.currentUser)) {
  //   Rimdian.log('error', 'invalid user', ValidateUser.errors)
  // }
  return Rimdian.log("info","new authenticated user detected, reset context",a),Rimdian.currentDevice=void 0,Rimdian._localStorage.remove("device"),Rimdian._createDevice(),Rimdian.currentSession=void 0,Rimdian.deleteCookie(Rimdian.config.namespace+"session"),Rimdian._handleSession(),Object.keys(a).forEach(function(b){"string"==typeof a[b]&&""===a[b]&&delete a[b];}),Rimdian.currentUser=__assign({},a),void(void 0===Rimdian.currentUser.created_at&&(Rimdian.currentUser.created_at=new Date().toISOString()));// create a user_alias if previous user was anonymous
  // and merge new user data & enqueue user
  Rimdian.currentUser.is_authenticated||(Rimdian.log("info","alias previous user",Rimdian.currentUser.external_id,"with user",a),Rimdian.itemsQueue.add("user_alias",{from_user_external_id:Rimdian.currentUser.external_id,to_user_external_id:a.external_id,to_user_is_authenticated:!0===a.is_authenticated,to_user_created_at:a.created_at?a.created_at:new Date().toISOString()}));}// update user context with new data
  var b=__assign(__assign({},Rimdian.currentUser),a);// // validate current user
  // if (!ValidateUser(Rimdian.currentUser)) {
  //   // abort on error
  //   Rimdian.log('error', 'invalid user', ValidateUser.errors)
  //   return
  // }
  Rimdian.currentUser=b,Rimdian.setCookie(Rimdian.config.namespace+"user",JSON.stringify(Rimdian.currentUser),OneYearInSeconds);},// enqueue the current user
  saveUserProfile:function saveUserProfile(){return Rimdian.isReady?void Rimdian.itemsQueue.add("user",__assign({},Rimdian.currentUser)):(Rimdian.log("debug","RMD is not yet ready for saveUserProfile, queuing function..."),void Rimdian._execWhenReady(function(){return Rimdian.saveUserProfile()}))},isPageVisible:function isPageVisible(){return w.state===PageStates.active},// dispatch items to the collector
  dispatch:function dispatch(a){// wait for tracker to be ready
  if(!1===Rimdian.isReady)return Rimdian.log("info","RMD is not ready, retrying dispatch soon..."),void window.setTimeout(function(){Rimdian.dispatch(a);},50);// abort if we don't have user consent
  if(!Rimdian.dispatchConsent)return void Rimdian.log("info","RMD abort dispatch, no dispatch consent");// abort if we don't have items to dispatch
  if(0===Rimdian.itemsQueue.items.length&&0===Rimdian.dispatchQueue.length)return void Rimdian.log("info","RMD abort dispatch, no items to dispatch");var b=__assign({},Rimdian.currentDevice);// set updated_at only if user_agent changed, it will trigger a DB update
  b.user_agent!==navigator.userAgent&&(b.updated_at=new Date().toISOString()),b.user_agent=navigator.userAgent,b.language=navigator.language,b.ad_blocker=Rimdian.hasAdBlocker(),b.resolution=window.screen&&window.screen.width&&window.screen.height?window.screen.height>window.screen.width?window.screen.height+"x"+window.screen.width:window.screen.width+"x"+window.screen.height:void 0;for(// create dataImport batches of 20 items maximum
  var c=[],d=[];0<Rimdian.itemsQueue.items.length;)d.push(Rimdian.itemsQueue.items.shift()),20<=d.length&&(c.push(d),d=[]);// add remaining items to last batch
  // abort if we don't have data imports to dispatch
  return 0<d.length&&c.push(d),c.forEach(function(a){a.forEach(function(a){a.device=b;}),Rimdian.dispatchQueue.push({id:Rimdian.uuidv4(),workspace_id:Rimdian.config.workspace_id,items:a,context:{// set this field right before sending data
  // data_sent_at: new Date().toISOString()
  },created_at:new Date().toISOString()});}),0===Rimdian.dispatchQueue.length?void Rimdian.log("info","RMD abort dispatch, no data imports to dispatch"):void(// persist dispatch queue in local storage
  // send data to collector
  Rimdian._localStorage.set("dispatchQueue",JSON.stringify(Rimdian.dispatchQueue)),Rimdian.log("info","RMD sending data to collector"),Rimdian._initDispatchLoop(a))},_initDispatchLoop:function _initDispatchLoop(a){// abort if we are already dispatching
  if(Rimdian.isDispatching)return void Rimdian.log("info","RMD abort dispatch, already dispatching");// dispatch items
  Rimdian.isDispatching=!0,Rimdian.dispatchQueue.sort(function(c,a){return c.created_at<a.created_at?-1:c.created_at>a.created_at?1:0});var b=Rimdian.dispatchQueue[0];// post payload
  Rimdian._postPayload(b,0,a);},_postPayload:function _postPayload(a,b,c){Rimdian._post(a,c,function(d){var e=!0;if(d){Rimdian.log("error","RMD post payload error",d);// parse error, ignore error 400
  try{var f=JSON.parse(d)||{};// error is not a bad hit
  400<f.code&&(e=!1);}catch(a){e=!1;}if(Rimdian.log("info","sucess",e,"retry count",b),!1===e){// stop after 10 retry
  if(b>=Rimdian.config.max_retry)return Rimdian.isDispatching=!1,void Rimdian.log("info","max retry reached, aborting");// exponential backoff
  var g=100;// ms
  if(1<b)for(var h=0;h<b;h++)g*=2;return Rimdian.log("debug","retry in "+g+"ms"),void window.setTimeout(function(){b++,Rimdian._postPayload(a,b,c);},g)}}if(!0==e){// remove from the dispatch queue
  var j=[];// keep other batches
  Rimdian.dispatchQueue.forEach(function(b){b.id!==a.id&&j.push(b);}),Rimdian.dispatchQueue=j,Rimdian._localStorage.set("dispatchQueue",JSON.stringify(Rimdian.dispatchQueue)),Rimdian.isDispatching=!1,0<Rimdian.dispatchQueue.length&&Rimdian._initDispatchLoop(c);}});},_post:function _post(a,b,c){a.context.data_sent_at=new Date().toISOString();var d=JSON.stringify(a);// log info data
  // send dataImport to collector using beacon
  if(Rimdian.log("info","RMD sending data",d),b&&navigator.sendBeacon){var e=navigator.sendBeacon(Rimdian.config.host+"/live",new Blob([d],{type:"application/json"}));return c(e?null:"sendBeacon failed")}var f=new XMLHttpRequest;f.onload=function(){var a="response"in f?f.response:f.responseText;return 300<=f.status?c(a):c(null)},f.onerror=function(){return c("network request failed")},f.ontimeout=function(){return c("network request timeout")},f.open("POST",Rimdian.config.host+"/live",!0),f.setRequestHeader("Content-Type","application/json"),f.withCredentials=!0,f.send(d);},// retrieve existing user ID or create a new one
  // 1. id found in URL
  // 2. OR id exists in a cookie
  // 3. OR id is created
  _handleUser:function _handleUser(){// sometimes user IDs are injected by template engines and are not parsed - ie: {{ user_id }}
  // to avoid using/merging invalid user IDs, we forbid some patterns
  var a=["{{","}}","{%","%}","{#","#}","*|","|*"],b=Rimdian.getCookie(Rimdian.config.namespace+"user"),c={};b&&""!==b&&(c=JSON.parse(b),a.some(function(a){return c.external_id&&-1!==c.external_id.indexOf(a)})&&(c={},b=""),!0===Rimdian.config.from_cm&&c.id!==void 0&&c.external_id===void 0&&(c={external_id:c.id,is_authenticated:c.is_authenticated||!1,created_at:new Date().toISOString(),hmac:c.hmac||void 0}));// an authenticated user id can be the param "_authuid" or the combination of "_uid" + "_auth=true"
  var d=Rimdian.getQueryParam(document.URL,URLParams.user_authenticated_external_id),e=d||Rimdian.getQueryParam(document.URL,URLParams.user_external_id);// 1. ID found in URL
  // check if the user ID does not contain forbidden patterns
  if(e&&""!==e&&a.every(function(a){return -1===e.indexOf(a)})){Rimdian.log("info","found user ID in URL",e);var f=d?"true":Rimdian.getQueryParam(document.URL,URLParams.user_is_authenticated);return Rimdian.log("info","user authenticated URL value",f),Rimdian.currentUser={external_id:e,is_authenticated:"true"===f||"1"===f,created_at:new Date().toISOString(),hmac:Rimdian.getQueryParam(document.URL,URLParams.user_external_id_hmac)},b&&""!==b&&(Rimdian.log("info","found another user ID in cookie",b),!1===c.is_authenticated&&c.external_id!==Rimdian.currentUser.external_id&&(Rimdian.log("info","alias previous user",c.external_id,"with user",Rimdian.currentUser.external_id),Rimdian.itemsQueue.add("user_alias",{from_user_external_id:c.external_id,to_user_external_id:Rimdian.currentUser.external_id,to_user_is_authenticated:Rimdian.currentUser.is_authenticated,to_user_created_at:Rimdian.currentUser.created_at}))),void Rimdian.setCookie(Rimdian.config.namespace+"user",JSON.stringify(Rimdian.currentUser),OneYearInSeconds)}// 2. ID found in cookie
  return b&&""!==b?(Rimdian.log("info","found user ID in cookie",b),Rimdian.currentUser=c,void Rimdian.setCookie(Rimdian.config.namespace+"user",JSON.stringify(Rimdian.currentUser),OneYearInSeconds)):void// 3. ID is created, with anonymous user
  Rimdian._createUser(Rimdian.uuidv4(),!1,new Date().toISOString())},_createUser:function _createUser(a,b,c){Rimdian.currentUser={external_id:a,is_authenticated:b,created_at:c},Rimdian.log("info","creating new user",__assign({},Rimdian.currentUser)),Rimdian.setCookie(Rimdian.config.namespace+"user",JSON.stringify(Rimdian.currentUser),OneYearInSeconds);},// retrieve eventual user details provided in URL
  _enrichUserContext:function _enrichUserContext(){// update current user profile with other parameters
  var a=[{key:"email",value:Rimdian.getQueryParam(document.URL,"_email")},{key:"email_md5",value:Rimdian.getQueryParam(document.URL,"_emailmd5")},{key:"email_sha1",value:Rimdian.getQueryParam(document.URL,"_emailsha1")},{key:"email_sha256",value:Rimdian.getQueryParam(document.URL,"_emailsha256")},{key:"telephone",value:Rimdian.getQueryParam(document.URL,"_telephone")}];// enrich user context with other parameters
  // update user cookie
  a.forEach(function(a){a.value&&""!==a.value&&(Rimdian.currentUser[a.key]=a.value);}),Rimdian.setCookie(Rimdian.config.namespace+"user",JSON.stringify(Rimdian.currentUser),OneYearInSeconds);},// retrieve existing device ID or create a new one
  // 1. id found in URL
  // 2. id exists in a cookie
  // 3. id is created
  // to save cookie space, device context is enriched while dispatching data
  _handleDevice:function _handleDevice(){var a=Rimdian.getQueryParam(document.URL,URLParams.device_external_id);// 1. ID found in URL
  if(a&&""!==a)return Rimdian.log("info","found device ID in URL",a),Rimdian.currentDevice={external_id:a,created_at:new Date().toISOString(),user_agent:navigator.userAgent},void Rimdian.setCookie(Rimdian.config.namespace+"device",JSON.stringify(Rimdian.currentDevice),OneYearInSeconds);// 2. ID found in cookie
  var b=Rimdian.getCookie(Rimdian.config.namespace+"device");if(b&&""!==b){// check if we already had a legacy client ID in a cookie
  if(Rimdian.log("info","found device ID in cookie",b),Rimdian.currentDevice=JSON.parse(b),!0===Rimdian.config.from_cm){var c=Rimdian.getCookie("_cm_cid");if(c&&""!==c){Rimdian.currentDevice.external_id=c;// extract its creation timestamp
  var d=Rimdian.getCookie("_cm_cidat");d&&""!==d&&(Rimdian.currentDevice.created_at=new Date(parseInt(d,10)).toISOString());}}return void Rimdian.setCookie(Rimdian.config.namespace+"device",JSON.stringify(Rimdian.currentDevice),OneYearInSeconds)}// 3. ID is created
  Rimdian._createDevice();},_createDevice:function _createDevice(){Rimdian.log("info","creating new device ID"),Rimdian.currentDevice={external_id:Rimdian.uuidv4(),created_at:new Date().toISOString(),user_agent:navigator.userAgent},Rimdian.setCookie(Rimdian.config.namespace+"device",JSON.stringify(Rimdian.currentDevice),OneYearInSeconds);},// polyfill
  _addEventListener:function _addEventListener(a,b,c,d){return a.addEventListener?(a.addEventListener(b,c,d),!0):a.attachEvent?a.attachEvent("on"+b,c):void(a["on"+b]=c)},// load config, existing data and start tracking
  _onReady:function _onReady(a){// avoid calling onReady twice
  if(!Rimdian.isReady){// check if browser is legit
  if(Rimdian.isReady=!0,Rimdian.log("info","onReady() called"),!Rimdian.isBrowserLegit())return void Rimdian.log("warn","Browser is not legit");// merge cfg with default config & validate
  var b=__assign(__assign({},Rimdian.config),a);// if (!ValidateConfig(config)) {
  //   Rimdian.log('error', 'RMD Config error:', ValidateConfig.errors)
  //   return
  // }
  // save config
  // every minute
  // decorate cross domains links
  if(Rimdian.config=b,Rimdian.log("info","RMD Config is:",Rimdian.config),window.localStorage&&(Rimdian.itemsQueue.items=JSON.parse(localStorage.getItem(Rimdian.config.namespace+"items")||"[]"),Rimdian.dispatchQueue=JSON.parse(localStorage.getItem(Rimdian.config.namespace+"dispatchQueue")||"[]")),Rimdian._handleUser(),Rimdian._enrichUserContext(),Rimdian._handleDevice(),Rimdian._handleSession(),0<Rimdian.onReadyQueue.length&&(Rimdian.onReadyQueue.forEach(function(a){Rimdian.log("debug","executing queued function",a),a();}),Rimdian.onReadyQueue=[]),window.setInterval(function(){if(Rimdian.currentPageview&&Rimdian.isPageVisible()){// get current session from cookie to see if it expired
  var a=Rimdian.getCookie(Rimdian.config.namespace+"session");a&&Rimdian.setCookie(Rimdian.config.namespace+"session",a,Rimdian.config.session_timeout);}},6e4),0<Rimdian.config.cross_domains.length)// loop over every <a> tag and add the decorateURL function
  for(var c,d=0;d<document.links.length;d++)c=document.links[d],Rimdian.config.cross_domains.forEach(function(a){-1!==c.href.indexOf(a)&&(Rimdian._addEventListener(c,"click",Rimdian._decorateURL,!0),Rimdian._addEventListener(c,"mousedown",Rimdian._decorateURL,!0));});// use visibilitychange event to send pageview secs in beacon
  // use https://github.com/GoogleChromeLabs/page-lifecycle
  // https://developer.chrome.com/blog/page-lifecycle-api/#advice-hidden
  // watch page state changes
  w.addEventListener("statechange",function(a){Rimdian.log("info","page state changed from",a.oldState,"to",a.newState),a.oldState===PageStates.active&&a.newState===PageStates.passive?Rimdian._onPagePassive():a.oldState===PageStates.passive&&a.newState===PageStates.active&&Rimdian._onPageActive();});}},_execWhenReady:function _execWhenReady(a){Rimdian.isReady?a():Rimdian.onReadyQueue.push(a);},_normalizeUTMSource:function _normalizeUTMSource(a){// replace 98ad0bb6e8ada73c81aab4e8c2637e7f.safeframe.googlesyndication.com by safeframe.googlesyndication.com
  return a&&-1!==a.indexOf("safeframe.googlesyndication.com")?"safeframe.googlesyndication.com":a},// session is stored in a cookie and will expire with its cookie
  // scenarii:
  // 1. no existing session -> create new session
  // 2. existing session with same referrer origin -> continue session
  // 3. existing session with different referrer origin -> create new session
  _handleSession:function _handleSession(){// extract current utm params from url
  var a=Rimdian.getQueryParam(document.URL,"utm_source")||Rimdian.getHashParam(window.location.hash,"utm_source"),b=Rimdian.getQueryParam(document.URL,"utm_medium")||Rimdian.getHashParam(window.location.hash,"utm_medium"),c=Rimdian.getQueryParam(document.URL,"utm_campaign")||Rimdian.getHashParam(window.location.hash,"utm_campaign"),d=Rimdian.getQueryParam(document.URL,"utm_content")||Rimdian.getHashParam(window.location.hash,"utm_content"),e=Rimdian.getQueryParam(document.URL,"utm_term")||Rimdian.getHashParam(window.location.hash,"utm_term"),f=Rimdian.getQueryParam(document.URL,"utm_id")||Rimdian.getHashParam(window.location.hash,"utm_id"),g=Rimdian.getQueryParam(document.URL,"utm_id_from")||Rimdian.getHashParam(window.location.hash,"utm_id_from"),h=Rimdian.getReferrer();if(h){// parse referrer URL with <a> tag
  var i=document.createElement("a");i.href=h;// check if referrer came from another domain
  var j=!1;// check if different domain is not is the cross_domain config
  if(i.hostname&&i.hostname!==window.location.hostname){var k=!1;Rimdian.config.cross_domains&&Rimdian.config.cross_domains.length&&Rimdian.config.cross_domains.forEach(function(a){-1!==i.href.indexOf(a)&&(k=!0);}),!1==k&&(j=!0);}j&&(a=i.hostname,(!b||""===b)&&(b="referral"),0===h.search("https?://(.*)google.([^/?]*)")?e=Rimdian.getQueryParam(h,"q"):0===h.search("https?://(.*)bing.com")?e=Rimdian.getQueryParam(h,"q"):0===h.search("https?://(.*)search.yahoo.com")?e=Rimdian.getQueryParam(h,"p"):0===h.search("https?://(.*)ask.com")?e=Rimdian.getQueryParam(h,"q"):0===h.search("https?://(.*)search.aol.com")?e=Rimdian.getQueryParam(h,"q"):0===h.search("https?://(.*)duckduckgo.com")&&(e=Rimdian.getQueryParam(h,"q")));}// extract gclid+fbclid+MSCLKID from url into utm_id + utm_id_from
  ["gclid","fbclid","msclkid","aecid"].forEach(function(a){var b=Rimdian.getQueryParam(document.URL,a)||Rimdian.getHashParam(window.location.hash,a);b&&(f=b,g=a);}),"gclid"===g&&"referral"===b&&(b="ads"),a=Rimdian._normalizeUTMSource(a),Rimdian.log("info","RMD utm_source is:",a),Rimdian.log("info","RMD utm_medium is:",b),Rimdian.log("info","RMD utm_campaign is:",c),Rimdian.log("info","RMD utm_content is:",d),Rimdian.log("info","RMD utm_term is:",e),Rimdian.log("info","RMD utm_id is:",f),Rimdian.log("info","RMD utm_id_from is:",g);// read session cookie
  var l=Rimdian.getCookie(Rimdian.config.namespace+"session");// 1. no existing session -> create new session
  if(!l||""===l)return Rimdian.log("info","RMD session cookie not found"),void Rimdian._startNewSession({utm_source:a,utm_medium:b,utm_campaign:c,utm_content:d,utm_term:e,utm_id:f,utm_id_from:g});// check if this origin should be ignored
  var m;a&&""!==a&&0<Rimdian.config.ignored_origins.length&&(m=Rimdian.config.ignored_origins.find(function(d){// source medium matches
  return d.utm_source===a&&d.utm_medium===b&&(!(d.utm_campaign&&""!==d.utm_campaign)||!!(c&&d.utm_campaign===c))}));// process existing session
  var n=JSON.parse(l);Rimdian.log("info","RMD existing session is:",n);// check if session origin has changed from previous page
  var o=!0;// 2. if this origin is ignored, or same origin, or empty origin, resume session
  return a&&""!==a&&n.utm_source!==a&&(o=!1),b&&""!==b&&n.utm_medium!==b&&(o=!1),c&&""!==c&&n.utm_campaign!==c&&(o=!1),d&&""!==d&&n.utm_content!==d&&(o=!1),e&&""!==e&&n.utm_term!==e&&(o=!1),f&&""!==f&&n.utm_id!==f&&(o=!1),m||o||!a||""===a?(Rimdian.log("info","RMD resume session (ignored:"+(m?"yes":"no")+", isEqual:"+o+", utm_source:"+a+")"),Rimdian.currentSession=n,void Rimdian.setCookie(Rimdian.config.namespace+"session",JSON.stringify(Rimdian.currentSession),Rimdian.config.session_timeout)):void// 3. origin has changed, start new session
  Rimdian._startNewSession({utm_source:a,utm_medium:b,utm_campaign:c,utm_content:d,utm_term:e,utm_id:f,utm_id_from:g})},_onPagePassive:function _onPagePassive(){Rimdian.log("info","page is passive state"),Rimdian.itemsQueue.addPageviewDuration(),Rimdian.dispatch(!0);},_onPageActive:function _onPageActive(){Rimdian.log("info","page is active state");// abort if we are not tracking the current pageview
  Rimdian.currentPageview&&(Rimdian.currentPageviewVisibleSince=new Date);},getTimezone:function getTimezone(){var a,b=null===(a=window.Intl)||void 0===a?void 0:a.DateTimeFormat;if(b){var c=new b().resolvedOptions().timeZone;if(c)return c}},getQueryParam:function getQueryParam(a,b){try{var c=new URL(a),d=new URLSearchParams(c.search);return d.get(b)||void 0}catch(a){}},getHashParam:function getHashParam(a,b){var c=a.match(new RegExp(b+"=([^&]*)"));return c?c[1]:void 0},updateURLParam:function updateURLParam(a,b,c){var d=new URL(a),e=new URLSearchParams(d.search);return e.set(b,c),d.search=e.toString(),d.toString()},hasAdBlocker:function hasAdBlocker(){var a=document.createElement("div");a.innerHTML="&nbsp;",a.className="adsbox";var b=!1;try{// body may not exist, that's why we need try/catch
  document.body.appendChild(a),b=0===document.getElementsByClassName("adsbox")[0].offsetHeight,document.body.removeChild(a);}catch(a){b=!1;}return b},isBrowserLegit:function isBrowserLegit(){// detect IE 9
  var a=navigator.userAgent.toLowerCase();return !(-1!==a.indexOf("msie")&&9>=parseInt(a.split("msie")[1],10))&&!/(google web preview|baiduspider|yandexbot|bingbot|googlebot|yahoo! slurp|nuhk|yammybot|openbot|slurp|msnBot|ask jeeves\/teoma|ia_archiver)/i.test(navigator.userAgent)&&!navigator.webdriver;// detect known bot
  // detect headless chrome
  },uuidv4:function uuidv4(){return "xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx".replace(/[xy]/g,function(a){var b=0|16*Math.random(),c="x"==a?b:8|3&b;return c.toString(16)})},md5:function md5(a){return _md(a)},getReferrer:function getReferrer(){var a;try{a=""===window.top.document.referrer?void 0:window.top.document.referrer;}catch(b){if(window.parent)try{a=""===window.parent.document.referrer?void 0:window.parent.document.referrer;}catch(b){a=void 0;}}return a||(a=""===document.referrer?void 0:document.referrer),a},_startNewSession:function _startNewSession(a){// persist session to cookie
  a.utm_source&&""!==a.utm_source||(a.utm_source="direct"),a.utm_medium&&""!==a.utm_medium||(a.utm_medium="none"),Rimdian.currentSession={external_id:Rimdian.uuidv4(),created_at:new Date().toISOString(),device_external_id:Rimdian.currentDevice.external_id,landing_page:window.location.href,referrer:Rimdian.getReferrer(),timezone:Rimdian.getTimezone(),utm_source:a.utm_source,utm_medium:a.utm_medium,utm_campaign:a.utm_campaign,utm_content:a.utm_content,utm_term:a.utm_term,utm_id:a.utm_id,utm_id_from:a.utm_id_from,duration:0,pageviews_count:0,interactions_count:0},Rimdian.log("info","RMD new session is:",Rimdian.currentSession),Rimdian.setCookie(Rimdian.config.namespace+"session",JSON.stringify(Rimdian.currentSession),Rimdian.config.session_timeout);},getCookie:function getCookie(a){return decodeURIComponent(document.cookie.replace(new RegExp("(?:(?:^|.*;)\\s*"+encodeURIComponent(a).replace(/[-.+*]/g,"\\$&")+"\\s*\\=\\s*([^;]*).*$)|^.*$"),"$1"))||null},// cookies are secured and cross-domain by default
  setCookie:function setCookie(a,b,c){// cross_domain
  var d=window.location.hostname.match(/[a-z0-9][a-z0-9\-]+\.[a-z\.]{2,6}$/i),e=d?d[0]:"",f=e?"; domain=."+e:"",g=new Date;g.setTime(g.getTime()+1e3*c);var h="; expires="+g.toUTCString(),i=a+"="+encodeURIComponent(b)+h+"; path=/"+f+"; secure";document.cookie=i;},deleteCookie:function deleteCookie(a){Rimdian.setCookie(a,"",-1);},_localStorage:{get:function get(a){return localStorage.getItem(Rimdian.config.namespace+a)},set:function set(a,b){try{localStorage.setItem(Rimdian.config.namespace+a,b);}catch(a){Rimdian.log("error","localStorage error:",a);}},remove:function remove(a){localStorage.removeItem(Rimdian.config.namespace+a);}},// inject the device + user ids on the fly
  _decorateURL:function _decorateURL(a){var b=a.target;b.href=Rimdian.updateURLParam(b.href,URLParams.device_external_id,Rimdian.currentDevice.external_id),b.href=Rimdian.updateURLParam(b.href,URLParams.user_external_id,Rimdian.currentUser.external_id),b.href=Rimdian.updateURLParam(b.href,URLParams.user_is_authenticated,Rimdian.currentUser.is_authenticated.toString()),Rimdian.currentUser.hmac&&(b.href=Rimdian.updateURLParam(b.href,URLParams.user_external_id_hmac,Rimdian.currentUser.hmac));},// the cart hash is a combination of public_url + products id + items variant id + items quantity
  _cartHash:function _cartHash(a){var b=a.public_url?a.public_url:"";return a.items&&0<a.items.length&&a.items.forEach(function(a){b=b+a.product_external_id+(a.variant_external_id?a.variant_external_id:"")+(a.quantity||"0");}),_md(b)},_wipeAll:function _wipeAll(){window.confirm("Do you know what you are doing?")&&(Rimdian.deleteCookie(Rimdian.config.namespace+"device"),Rimdian.deleteCookie(Rimdian.config.namespace+"user"),Rimdian.deleteCookie(Rimdian.config.namespace+"session"),Rimdian._localStorage.remove("items"),Rimdian._localStorage.remove("dispatchQueue"),Rimdian.currentUser=void 0,Rimdian.currentDevice=void 0,Rimdian.currentSession=void 0,Rimdian.currentCart=void 0,Rimdian.currentPageview=void 0,Rimdian.isReady=!1,Rimdian._onReady(Rimdian.config));}};

  return Rimdian;

})();
