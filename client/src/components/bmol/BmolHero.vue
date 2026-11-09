<template>
  <section class="hero" :class="{ 'is-target': isTarget }" data-bmol-rank="1" :aria-labelledby="titleId">
    <header class="hero__head">
      <Crown :size="18" aria-hidden="true" />
      <h3 :id="titleId" class="hero__title">{{ t("library.bmol.hero.title") }}</h3>
      <span v-if="group.items.length > 1" class="hero__tie">{{ t("library.bmol.hero.tie", { count: group.items.length }) }}</span>
    </header>
    <div class="hero__cards" :class="{ 'hero__cards--multi': group.items.length > 1 }">
      <slot :items="group.items" />
    </div>
  </section>
</template>

<script setup lang="ts">
import { useId } from "vue"
import { useI18n } from "vue-i18n"
import { Crown } from "lucide-vue-next"
import type { RankGroup } from "./bmolRanks"

defineProps<{ group: RankGroup; isTarget?: boolean }>()
const { t } = useI18n()
const titleId = useId()
</script>

<style scoped>
.hero {
  padding: 0.75rem;
  background: linear-gradient(160deg, color-mix(in srgb, var(--color-accent) 12%, var(--color-surface-2)), var(--color-surface-2) 60%);
  border: 1px solid color-mix(in srgb, var(--color-accent) 35%, var(--color-border));
  border-radius: var(--radius-card);
}
.hero.is-target { outline: 2px dashed var(--color-accent); outline-offset: 3px; }
.hero__head { display: flex; align-items: center; gap: 0.5rem; margin-bottom: 0.875rem; color: var(--color-accent); }
.hero__title { font-size: var(--text-small); font-weight: 800; letter-spacing: 0.08em; text-transform: uppercase; }
.hero__tie { margin-left: auto; font-size: var(--text-caption); color: var(--color-text-secondary); }
.hero__cards { display: grid; gap: 1rem; max-width: 11rem; margin-inline: auto; }
.hero__cards--multi { grid-template-columns: repeat(2, minmax(0, 1fr)); max-width: none; }
/* Ties: title and note sit below each poster so long titles keep their width. */
.hero__cards--multi :deep(.bcard--hero.bcard--hero) { flex-direction: column; gap: 0.5rem; }
.hero__cards--multi :deep(.bcard--hero .bcard__poster) { width: 100%; max-width: 12rem; }
.hero__cards--multi :deep(.bcard--hero .bcard__body) { min-width: 0; padding-top: 0; }
@media (min-width: 768px) {
  .hero { padding: 1.5rem; }
  .hero__cards { max-width: none; margin-inline: 0; }
}
</style>
