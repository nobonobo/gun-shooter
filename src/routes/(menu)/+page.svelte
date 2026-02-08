<script>
  import { onMount } from "svelte";
  export let data;
  let ID;
  onMount(() => {
    ID = globalThis.Go.UUID();
    globalThis.Go.Stop();
  });
</script>

<svelte:head>
  <title>{data.title}</title>
</svelte:head>
<p>HostID: {ID}</p>
<a href="/room?hostID={ID}" class="btn preset-filled-primary-500">Room</a>
