import{S as q,i as B,s as F,e as b,b as v,f as h,g as y,h as f,P as w,Q as J,k as N,R as O,n as Q,t as I,a as M,o as E,w as S,r as j,u as Y,x as P,O as z,c as A,m as G,d as H}from"./index.97f016a1.js";function C(o,t,l){const n=o.slice();return n[5]=t[l],n}function D(o,t,l){const n=o.slice();return n[5]=t[l],n}function K(o,t){let l,n,_=t[5].title+"",u,r,s,c;function m(){return t[4](t[5])}return{key:o,first:null,c(){l=b("button"),n=b("div"),u=S(_),r=v(),h(n,"class","txt"),h(l,"class","tab-item svelte-1maocj6"),j(l,"active",t[0]===t[5].language),this.first=l},m(g,d){y(g,l,d),f(l,n),f(n,u),f(l,r),s||(c=Y(l,"click",m),s=!0)},p(g,d){t=g,d&2&&_!==(_=t[5].title+"")&&P(u,_),d&3&&j(l,"active",t[0]===t[5].language)},d(g){g&&E(l),s=!1,c()}}}function R(o,t){let l,n,_,u,r,s,c=t[5].title+"",m,g,d,p,k;return n=new z({props:{language:t[5].language,content:t[5].content}}),{key:o,first:null,c(){l=b("div"),A(n.$$.fragment),_=v(),u=b("div"),r=b("em"),s=b("a"),m=S(c),g=S(" SDK"),p=v(),h(s,"href",d=t[5].url),h(s,"target","_blank"),h(s,"rel","noopener noreferrer"),h(r,"class","txt-sm txt-hint"),h(u,"class","txt-right"),h(l,"class","tab-item svelte-1maocj6"),j(l,"active",t[0]===t[5].language),this.first=l},m(e,a){y(e,l,a),G(n,l,null),f(l,_),f(l,u),f(u,r),f(r,s),f(s,m),f(s,g),f(l,p),k=!0},p(e,a){t=e;const i={};a&2&&(i.language=t[5].language),a&2&&(i.content=t[5].content),n.$set(i),(!k||a&2)&&c!==(c=t[5].title+"")&&P(m,c),(!k||a&2&&d!==(d=t[5].url))&&h(s,"href",d),(!k||a&3)&&j(l,"active",t[0]===t[5].language)},i(e){k||(I(n.$$.fragment,e),k=!0)},o(e){M(n.$$.fragment,e),k=!1},d(e){e&&E(l),H(n)}}}function L(o){let t,l,n=[],_=new Map,u,r,s=[],c=new Map,m,g=o[1];const d=e=>e[5].language;for(let e=0;e<g.length;e+=1){let a=D(o,g,e),i=d(a);_.set(i,n[e]=K(i,a))}let p=o[1];const k=e=>e[5].language;for(let e=0;e<p.length;e+=1){let a=C(o,p,e),i=k(a);c.set(i,s[e]=R(i,a))}return{c(){t=b("div"),l=b("div");for(let e=0;e<n.length;e+=1)n[e].c();u=v(),r=b("div");for(let e=0;e<s.length;e+=1)s[e].c();h(l,"class","tabs-header compact left"),h(r,"class","tabs-content"),h(t,"class","tabs sdk-tabs m-b-base svelte-1maocj6")},m(e,a){y(e,t,a),f(t,l);for(let i=0;i<n.length;i+=1)n[i].m(l,null);f(t,u),f(t,r);for(let i=0;i<s.length;i+=1)s[i].m(r,null);m=!0},p(e,[a]){a&3&&(g=e[1],n=w(n,a,d,1,e,g,_,l,J,K,null,D)),a&3&&(p=e[1],N(),s=w(s,a,k,1,e,p,c,r,O,R,null,C),Q())},i(e){if(!m){for(let a=0;a<p.length;a+=1)I(s[a]);m=!0}},o(e){for(let a=0;a<s.length;a+=1)M(s[a]);m=!1},d(e){e&&E(t);for(let a=0;a<n.length;a+=1)n[a].d();for(let a=0;a<s.length;a+=1)s[a].d()}}}const T="pb_sdk_preference";function U(o,t,l){let n,{js:_=""}=t,{dart:u=""}=t,r=localStorage.getItem(T)||"javascript";const s=c=>l(0,r=c.language);return o.$$set=c=>{"js"in c&&l(2,_=c.js),"dart"in c&&l(3,u=c.dart)},o.$$.update=()=>{o.$$.dirty&1&&r&&localStorage.setItem(T,r),o.$$.dirty&12&&l(1,n=[{title:"JavaScript",language:"javascript",content:_,url:"https://github.com/pocketbase/js-sdk/tree/rc"},{title:"Dart",language:"dart",content:u,url:"https://github.com/pocketbase/dart-sdk/tree/rc"}])},[r,n,_,u,s]}class W extends q{constructor(t){super(),B(this,t,U,L,F,{js:2,dart:3})}}export{W as S};
