<template>
  <span class="msb" data-test="manga-status-badge">{{ t(LABEL_KEY[status]) }}</span>
</template>

<script setup lang="ts">
import { useI18n } from "vue-i18n"

/** Manga reads as "Want to read" / "Read" on mixed profile tabs, where the tab name would be wrong. */
defineProps<{ status: "watchlist" | "watched" }>()
const { t } = useI18n()

const LABEL_KEY = {
  watchlist: "manga.detail.actions.wantToRead",
  watched: "manga.detail.actions.read",
} as const
</script>

<style scoped>
.msb { display: inline-flex; min-width: 0; padding: 0.125rem 0.5rem; font-size: var(--text-caption); font-weight: 600; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; color: var(--color-text-secondary); background: var(--color-surface-3); border: 1px solid var(--color-border); border-radius: 9999px; }
</style>
