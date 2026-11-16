<template>
  <section class="manga-read" :aria-labelledby="headingId">
    <h2 :id="headingId" class="manga-read__heading">{{ t("manga.detail.readTitle") }}</h2>
    <p v-if="!links.length" class="manga-read__empty">{{ t("manga.detail.readEmpty") }}</p>
    <ul v-else class="manga-read__list">
      <li v-for="link in links" :key="link.url">
        <a
          class="manga-read__link"
          :href="link.url"
          target="_blank"
          rel="noopener noreferrer"
          :aria-label="t('manga.detail.readOpens', { site: link.site })"
        >
          <span class="manga-read__site">{{ link.site }}</span>
          <span v-if="link.language" class="manga-read__lang">{{ link.language }}</span>
        </a>
      </li>
    </ul>
  </section>
</template>

<script setup lang="ts">
import { computed } from "vue"
import { useI18n } from "vue-i18n"
import { isWebUrl } from "./mangaFormat"
import type { MangaReadLink } from "@/types/manga"

const props = defineProps<{ links: MangaReadLink[] }>()

const { t } = useI18n()
const headingId = "manga-read-heading"
/** Only web links are offered; anything else from the API is dropped. */
const links = computed(() => props.links.filter((link) => isWebUrl(link.url)))
</script>

<style scoped>
.manga-read {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}
.manga-read__heading {
  margin: 0;
  font-size: var(--text-body);
  font-weight: 600;
  color: var(--color-text-primary);
}
.manga-read__empty {
  margin: 0;
  font-size: var(--text-small);
  color: var(--color-text-muted);
}
.manga-read__list {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
  margin: 0;
  padding: 0;
  list-style: none;
}
.manga-read__link {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.75rem;
  min-height: 2.5rem;
  padding: 0.5rem 0.875rem;
  border-radius: var(--radius-btn);
  border: 1px solid var(--color-border-strong);
  color: var(--color-text-primary);
  text-decoration: none;
}
.manga-read__link:hover {
  background: var(--color-surface-3);
}
.manga-read__site {
  font-size: var(--text-small);
  font-weight: 600;
}
.manga-read__lang {
  font-size: var(--text-caption);
  color: var(--color-text-muted);
}
</style>
