<script>
  import "./layout.css";
  import favicon from "$lib/assets/favicon.svg";
  import { CircleUserIcon, MenuIcon } from "@lucide/svelte";
  import { AppBar } from "@skeletonlabs/skeleton-svelte";
  let { children } = $props();
</script>

<svelte:head>
  <link rel="icon" href={favicon} />
</svelte:head>
<AppBar>
  <AppBar.Toolbar class="grid-cols-[auto_1fr_auto]">
    <AppBar.Lead>
      <button type="button" class="btn-icon btn-icon-lg hover:preset-tonal"
        ><MenuIcon /></button
      >
    </AppBar.Lead>
    <AppBar.Headline>
      <p class="text-2xl">Gun Shooter</p>
    </AppBar.Headline>
    <AppBar.Trail>
      <button type="button" class="btn-icon hover:preset-tonal"
        ><CircleUserIcon class="size-6" /></button
      >
    </AppBar.Trail>
  </AppBar.Toolbar>
</AppBar>
<div class="page-container">
  {@render children()}
</div>
