<script>
  import { onMount } from "svelte";
  import { browser } from "$app/environment";
  import { page } from "$app/stores";
  export let data;
  const hostID = browser ? $page.url.searchParams.get("hostID") : "";
  onMount(async () => {
    await globalThis.Go.Stop();
  });
</script>

<svelte:head>
  <title>{data.title}</title>
</svelte:head>
<nav class="btn-group p-4 md:flex-row">
  <a href="/room?hostID={hostID}" class="btn preset-filled-error-500">Quit</a>
</nav>

<form method="GET" action="/scope.html" class="w-full max-w-md space-y-4 p-4">
  <input type="hidden" name="hostID" value={hostID} />
  <label class="label">
    <span class="label-text">Name</span>
    <input class="input" type="text" name="name" value="No Name" />
  </label>
  <button class="btn preset-filled-primary-500">Connect</button>
</form>
