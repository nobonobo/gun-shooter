<script>
  import { onMount, onDestroy } from "svelte";
  import { browser } from "$app/environment";
  import { page } from "$app/stores";
  import QRCode from "qrcode";
  import { goto } from "$app/navigation";
  export let data;
  const hostID = browser ? $page.url.searchParams.get("hostID") : "";
  let inform = {};
  let qrSvg;
  async function generateQR(text) {
    try {
      qrSvg = await QRCode.toString(text, {
        type: "svg",
        width: 256,
        margin: 2,
      });
    } catch (error) {
      console.error("QR生成エラー:", error);
    }
  }
  let timer = 0;
  onMount(() => {
    if (hostID) {
      const baseUrl = $page.url.origin + $page.url.pathname;
      const qrUrl = `${baseUrl}/scope?hostID=${hostID}`;
      generateQR(qrUrl);
      globalThis.Go.Listen(hostID);
      timer = setInterval(interval, 1000);
    }
  });
  onDestroy(() => {
    if (hostID) {
      globalThis.Go.Stop();
      clearInterval(timer);
    }
    console.log("destroyed");
  });
  function interval() {
    inform = JSON.parse(globalThis.Go.Inform());
    console.log(inform);
  }
</script>

<svelte:head>
  <title>{data.title}</title>
</svelte:head>
<div class="justify-center">
  <nav class="btn-group p-4 md:flex-row">
    <a href="/" class="btn preset-filled-error-500">Quit</a>
    <a href="/game?hostID={hostID}" class="btn preset-filled-primary-500">Game</a>
  </nav>
  <div class="flex gap-4">
  <div class="flex-none card p-4 preset-filled-surface-100-900 inline-block">
    <a href="/join?hostID={hostID}" class="inline-block">
      {#if qrSvg}
        {@html qrSvg}
      {:else}
        <p>生成中...</p>
      {/if}
    </a>
  </div>
  <div class="flex-1 card p-4 preset-filled-surface-100-900 inline-block">
    <ul>
    {#each Object.values(inform) as item (item.id)}
      <li>{item.name}</li>
    {/each}
    </ul>
  </div>
  </div>
</div>
