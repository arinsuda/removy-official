<template>
  <section v-if="items.length" class="manga-similar" :aria-labelledby="headingId">
    <h2 :id="headingId" class="manga-similar__heading">{{ t("manga.detail.similar") }}</h2>
    <div class="manga-similar__grid">
      <MangaCard v-for="item in items" :key="item.id" :item="item" />
    </div>
  </section>
</template>

<script setup lang="ts">
import { useI18n } from "vue-i18n"
import MangaCard from "./MangaCard.vue"
import type { MangaItem } from "@/types/manga"

defineProps<{ items: MangaItem[] }>()

const { t } = useI18n()
const headingId = "manga-similar-heading"
</script>

<style scoped>
.manga-similar {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}
.manga-similar__heading {
  margin: 0;
  font-size: var(--text-body);
  font-weight: 600;
  color: var(--color-text-primary);
}
.manga-similar__grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 1rem 0.75rem;
}
@media (min-width: 768px) {
  .manga-similar__grid {
    grid-template-columns: repeat(6, minmax(0, 1fr));
  }
}
</style>
