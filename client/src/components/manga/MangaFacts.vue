<template>
  <dl class="manga-facts">
    <div v-for="fact in facts" :key="fact.label" class="manga-facts__item">
      <dt>{{ fact.label }}</dt>
      <dd>{{ fact.value }}</dd>
    </div>
  </dl>
</template>

<script setup lang="ts">
import { computed } from "vue"
import { useI18n } from "vue-i18n"
import { knownNumber, statusKey } from "./mangaFormat"
import { anilistScore, type MangaDetail } from "@/types/manga"

const props = defineProps<{ manga: MangaDetail }>()

const { t } = useI18n()

/** Unknown counts and years read as "Unknown" instead of 0. */
const orUnknown = (value: number | null) => {
  const known = knownNumber(value)
  return known === null ? t("manga.facts.unknown") : String(known)
}

const facts = computed(() => {
  const m = props.manga
  const score = anilistScore(m.ratings)
  return [
    { label: t("manga.facts.kind"), value: t(`manga.kind.${m.kind}`) },
    { label: t("manga.facts.format"), value: t(`manga.format.${m.format}`) },
    { label: t("manga.facts.status"), value: t(`manga.status.${statusKey(m.status)}`) },
    { label: t("manga.facts.startYear"), value: orUnknown(m.start_year) },
    { label: t("manga.facts.chapters"), value: orUnknown(m.chapters) },
    { label: t("manga.facts.volumes"), value: orUnknown(m.volumes) },
    { label: t("manga.facts.score"), value: score === null ? t("manga.facts.unknown") : t("manga.scoreShort", { score }) },
  ]
})
</script>

<style scoped>
.manga-facts {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 0.75rem 1rem;
  margin: 0;
}
.manga-facts__item {
  display: flex;
  flex-direction: column;
  gap: 0.125rem;
  min-width: 0;
}
.manga-facts dt {
  font-size: var(--text-caption);
  color: var(--color-text-muted);
}
.manga-facts dd {
  margin: 0;
  font-size: var(--text-small);
  font-weight: 600;
  color: var(--color-text-primary);
  overflow-wrap: anywhere;
}
@media (min-width: 768px) {
  .manga-facts {
    grid-template-columns: repeat(4, minmax(0, 1fr));
  }
}
</style>
