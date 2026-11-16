<template>
  <p class="manga-attribution">
    <a :href="MANGA_ANILIST_URL" target="_blank" rel="noopener noreferrer">{{ t("manga.attribution") }}</a>
  </p>
</template>

<script setup lang="ts">
import { useI18n } from "vue-i18n"
import { MANGA_ANILIST_URL } from "./mangaConfig"

const { t } = useI18n()
</script>

<style scoped>
.manga-attribution {
  margin: 0;
  font-size: var(--text-caption);
  color: var(--color-text-muted);
}
.manga-attribution a {
  color: inherit;
  text-decoration: underline;
}
</style>
