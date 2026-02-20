```html
<canvas data-engine="three.js r164"
width="1580" height="747"
style="width: 1264px; height: 711px; margin-left: 0px; margin-top: -56.5px;"></canvas>
<video autoplay="" muted="" playsinline="" id="arjs-video"
style="width: 1264px; height: 711px; position: absolute; top: 0px; left: 0px; z-index: -2; margin-top: -56.5px; margin-left: 0px;"></video>
```
- videoはabsoluteでcanvasの上に重ねる
- おそらくcanvasがレイアウトされ、videoの見た目の幅と高さは合わせられる
- aspect維持のためのmargin-top、margin-leftが双方に差し込まれる

