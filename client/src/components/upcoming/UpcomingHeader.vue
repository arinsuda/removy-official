<template>
  <header class="up-header">
    <div class="up-header__text">
      <h1 class="up-header__title">{{ t("upcoming.title") }}</h1>
      <p class="up-header__sub">{{ t("upcoming.subtitle") }}</p>
    </div>
    <div class="up-header__tools">
      <label class="up-header__genre">
        <span class="up-header__label">{{ t("upcoming.genre.label") }}</span>
        <select class="up-header__select" :value="genre ?? ''" @change="onChange">
          <option value="">{{ t("upcoming.genre.all") }}</option>
          <option v-for="g in genres" :key="g.id" :value="g.id">{{ g.name }}</option>
        </select>
      </label>
      <AppButton :to="{ name: 'movies', query: { tab: 'now' } }" variant="outline" size="sm">
        {{ t("upcoming.inTheaters") }}
        <ArrowRight :size="14" aria-hidden="true" />
      </AppButton>
    </div>
  </header>
</template>

<script setup lang="ts">
import { useI18n } from "vue-i18n"
import { ArrowRight } from "lucide-vue-next"
import AppButton from "@/components/ui/AppButton.vue"
import type { Genre } from "@/types"

defineProps<{ genre: number | null; genres: Genre[] }>()
const emit = defineEmits<{ "update:genre": [genre: number | null] }>()
const { t } = useI18n()

function onChange(e: Event) {
  const v = (e.target as HTMLSelectElement).value
  emit("update:genre", v ? Number(v) : null)
}
</script>

<style scoped>
.up-header {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}
.up-header__title {
  margin: 0;
  font-size: var(--text-title);
  font-weight: 700;
  color: var(--color-text-primary);
}
.up-header__sub {
  margin: 0.25rem 0 0;
  font-size: var(--text-small);
  color: var(--color-text-secondary);
}
.up-header__tools {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 0.75rem;
}
.up-header__genre {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  font-size: var(--text-caption);
  color: var(--color-text-muted);
}
.up-header__select {
  min-height: 2rem;
  max-width: 12rem;
  padding: 0 0.5rem;
  border-radius: var(--radius-btn);
  border: 1px solid var(--color-border-strong);
  background: var(--color-surface-2);
  color: var(--color-text-primary);
  font: inherit;
  font-size: var(--text-small);
}
@media (min-width: 768px) {
  .up-header {
    flex-direction: row;
    align-items: flex-end;
    justify-content: space-between;
  }
  .up-header__title {
    font-size: 1.75rem;
  }
}
</style>
