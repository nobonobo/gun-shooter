<script>
  import { onMount, onDestroy } from "svelte";
  import { browser } from "$app/environment";
  import { page } from "$app/stores";
  import QRCode from "qrcode";
  import { base } from "$app/paths";
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
  onMount(async () => {
    if (hostID) {
      const baseUrl = $page.url.origin;
      const qrUrl = `${baseUrl}${base}/join?hostID=${hostID}`;
      generateQR(qrUrl);
      navigator.locks.request("wasm-load", async (lock) => {
        await globalThis.Go.Listen(hostID);
      });
      timer = setInterval(interval, 1000);
    }
  });
  onDestroy(async () => {
    if (hostID) {
      await globalThis.Go.Stop();
      clearInterval(timer);
    }
    console.log("destroyed");
  });
  async function interval() {
    inform = JSON.parse(await globalThis.Go.Inform());
    console.log(inform);
  }
</script>

<svelte:head>
  <title>{data.title}</title>
</svelte:head>
<div class="justify-center">
  <nav class="btn-group p-4 md:flex-row">
    <a href="{base}/" class="btn preset-filled-error-500">Quit</a>
    <a href="{base}/game?hostID={hostID}" class="btn preset-filled-primary-500"
      >Game</a
    >
  </nav>
  <div class="flex gap-4">
    <div class="flex-none card p-4 preset-filled-surface-100-900 inline-block">
      <a href="{base}/join?hostID={hostID}" class="inline-block">
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
