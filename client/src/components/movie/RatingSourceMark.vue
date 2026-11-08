<template>
  <span class="inline-flex shrink-0 items-center justify-center" aria-hidden="true" :data-mark="source">
    <span
      v-if="source === 'tmdb'"
      class="rounded-[4px] bg-linear-to-r from-mark-tmdb-from to-mark-tmdb-to px-1.5 py-[3px] text-[9px] font-extrabold leading-none tracking-wide text-mark-tmdb-ink"
    >TMDB</span>
    <span
      v-else-if="source === 'imdb'"
      class="rounded-[4px] bg-mark-imdb px-1.5 py-[3px] text-[10px] font-extrabold leading-none text-black"
    >IMDb</span>
    <svg v-else viewBox="0 0 20 20" class="h-[18px] w-[18px]" focusable="false">
      <circle cx="10" cy="11.5" r="7.5" class="fill-mark-rt" />
      <path d="M10 5 C8.5 2.5 6 2.5 5 3.5 C7 3.5 8 4.5 8.5 5.5 Z M10 5 C11.5 2.5 14 2.5 15 3.5 C13 3.5 12 4.5 11.5 5.5 Z" class="fill-mark-rt-leaf" />
    </svg>
  </span>
</template>

<script setup lang="ts">
defineProps<{ source: "tmdb" | "imdb" | "rt" }>()
</script>
