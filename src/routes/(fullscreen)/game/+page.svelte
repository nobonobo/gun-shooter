<script>
  import { onMount, onDestroy } from "svelte";
  import { browser } from "$app/environment";
  import { page } from "$app/stores";
  export let data;
  const hostID = browser ? $page.url.searchParams.get("hostID") : "";
  let canvas;
  let ctx;
  let width = 0;
  let height = 0;
  let rafId = 0;

  // イベントリスナーの参照を保持
  let resizeHandler;

  function resize() {
    width = window.innerWidth;
    height = window.innerHeight;
    if (canvas) {
      canvas.width = width;
      canvas.height = height;
    }
  }

  function draw(entries) {
    if (!ctx) return;

    // クリア
    ctx.fillStyle = "#f0f0f0";
    ctx.fillRect(0, 0, width, height);

    // 各エントリを描画
    entries.forEach((entry) => {
      const px = entry.x * width;
      const py = entry.y * height;

      // マーカー（円）
      ctx.fillStyle = entry.fire ? "#ff6b35" : "#4ecdc4";
      ctx.beginPath();
      ctx.arc(px, py, 6, 0, Math.PI * 2);
      ctx.fill();
      ctx.strokeStyle = "#333";
      ctx.lineWidth = 1;
      ctx.stroke();

      // 名前を斜め右上に1文字分ずらして表示
      ctx.font = "bold 14px sans-serif";
      ctx.fillStyle = "#333";
      ctx.textAlign = "left";
      ctx.textBaseline = "middle";

      const textX = px + 12; // 右12px（1文字分相当）
      const textY = py - 10; // 上10px
      ctx.fillText(entry.name, textX, textY);
    });
  }

  async function animate() {
    let s = await globalThis.Go.Inform();
    const entries = Object.values(JSON.parse(s));
    if (entries.length > 0) {
      draw(entries);
    }
    rafId = requestAnimationFrame(animate);
  }

  onMount(() => {
    resizeHandler = resize;
    resize();
    window.addEventListener("resize", resizeHandler);
    ctx = canvas.getContext("2d");
    rafId = requestAnimationFrame(animate);
  });

  onDestroy(async () => {
    if (resizeHandler) {
      window.removeEventListener("resize", resizeHandler);
    }
    if (rafId) cancelAnimationFrame(rafId);
    if (browser) {
      await globalThis.Go.Close();
    }
  });
</script>

<svelte:head>
  <title>{data.title}</title>
</svelte:head>
<main class="page-container">
  <canvas
    id="canvas"
    bind:this={canvas}
    {width}
    {height}
    style="width:100%; height:100%; display:block; cursor:none;"
  ></canvas>
  <!-- 4隅に固定配置する画像 -->
  <img
    src="./images/pattern-marker_0.png"
    class="corner-image top-left"
    alt="左上"
  />
  <img
    src="./images/pattern-marker_1.png"
    class="corner-image top-right"
    alt="右上"
  />
  <img
    src="./images/pattern-marker_2.png"
    class="corner-image bottom-left"
    alt="左下"
  />
  <img
    src="./images/pattern-marker_3.png"
    class="corner-image bottom-right"
    alt="右下"
  />
</main>

<style>
  .page-container {
    margin: 0;
    min-height: 100vh;
    background-color: #f0f0f0;
    background-image:
      /* 市松模様1（白黒チェック） */
      linear-gradient(
        45deg,
        #ddd 25%,
        transparent 25%,
        transparent 75%,
        #ddd 75%
      ),
      /* 市松模様2（位置をずらして） */
        linear-gradient(
          45deg,
          #ddd 25%,
          transparent 25%,
          transparent 75%,
          #ddd 75%
        );
    background-size: 40px 40px; /* 横10×縦10相当（40px=4マス分） */
    background-position:
      0 0,
      20px 20px; /* 2つ目の模様を半分ずらす */
    position: relative;
    overflow: hidden;
  }
  #canvas {
    position: absolute;
    top: 0;
    left: 0;
    width: 100%;
    height: 100%;
  }
  /* 4隅配置用の共通スタイル */
  .corner-image {
    position: fixed;
    width: 10vw; /* サイズ調整 */
    height: 10vw;
    z-index: 1000; /* 最前面に */
    pointer-events: none; /* クリック透過 */
  }

  /* 各隅の位置指定 */
  .top-left {
    top: 0;
    left: 0;
  }

  .top-right {
    top: 0;
    right: 0;
  }

  .bottom-left {
    bottom: 0;
    left: 0;
  }

  .bottom-right {
    bottom: 0;
    right: 0;
  }
</style>
