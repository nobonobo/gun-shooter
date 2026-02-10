<script>
  import { onMount } from "svelte";
  import { base } from "$app/paths";
  export let data;
  let ID;
  onMount(async () => {
    ID = await (await globalThis.WaitGo()).UUID();
    await globalThis.Go.Stop();
  });
</script>

<svelte:head>
  <title>{data.title}</title>
</svelte:head>
<p>HostID: {ID}</p>
<a href="{base}/room?hostID={ID}" class="btn preset-filled-primary-500">Room</a>
