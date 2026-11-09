<template>
  <li
    :id="id"
    role="option"
    :aria-selected="active"
    class="ssi"
    :class="{ 'ssi--active': active }"
    @mousedown.prevent
    @click="emit('select')"
    @mouseenter="emit('hover')"
  >
    <img v-if="item.poster_path && !failed" :src="item.poster_path" alt="" class="ssi-poster" loading="lazy" @error="failed = true" />
    <span v-else class="ssi-poster ssi-poster--empty" aria-hidden="true"><Film :size="16" /></span>

    <span class="ssi-body">
      <span class="ssi-title">
        <template v-for="(part, i) in parts" :key="i">
          <mark v-if="part.match" class="ssi-mark">{{ part.text }}</mark>
          <template v-else>{{ part.text }}</template>
        </template>
      </span>
      <span class="ssi-meta">
        <span class="ssi-chip" :data-type="item.type">{{ t(item.type === "movie" ? "navigation.movie" : "navigation.series") }}</span>
        <span v-if="item.year" class="ssi-year">{{ item.year }}</span>
        <span v-if="tmdb" class="ssi-tmdb" :aria-label="t('media.rating.tmdbScore', { score: tmdb })">
          <span aria-hidden="true">TMDB {{ tmdb }}</span> <!-- hardcode-ok: brand name -->
        </span>
        <span
          class="ssi-popcorn"
          :data-tier="tier"
          :aria-label="tier === 'none' ? t('navigation.search.noReviews') : t('media.rating.popcornScore', { score: popcornScore })"
        >
          <Popcorn :size="12" aria-hidden="true" />
          <span aria-hidden="true">{{ tier === "none" ? t("navigation.search.noReviews") : popcornScore }}</span>
        </span>
      </span>
    </span>
  </li>
</template>

<script setup lang="ts">
import { computed, ref } from "vue"
import { useI18n } from "vue-i18n"
import { Film, Popcorn } from "lucide-vue-next"
import { popcornTier, roundScore } from "@/components/movie/popcornTier"
import { highlightParts } from "./highlight"
import type { SearchSuggestion } from "./types"

const props = defineProps<{ item: SearchSuggestion; query: string; active: boolean; id: string }>()
const emit = defineEmits<{ select: []; hover: [] }>()

const { t } = useI18n()
const failed = ref(false)

const parts = computed(() => highlightParts(props.item.title, props.query))
const popcorn = computed(() => props.item.ratings?.popcorn)
const tier = computed(() => popcornTier(popcorn.value?.average, popcorn.value?.count))
const popcornScore = computed(() => roundScore(popcorn.value?.average).toFixed(1))
const tmdb = computed(() => {
  const src = props.item.ratings?.tmdb
  const v = src ? (src.available ? src.average : 0) : props.item.rating
  return v && v > 0 ? v.toFixed(1) : null
})
</script>

<style scoped>
.ssi {
  display: flex;
  align-items: center;
  gap: 12px;
  min-height: 72px;
  padding: 6px 12px;
  cursor: pointer;
  list-style: none;
  border-radius: 8px;
  transition: background-color 120ms var(--ease-out-soft, ease-out);
}
.ssi:hover,
.ssi--active {
  background: var(--color-surface-3);
}
.ssi--active {
  box-shadow: inset 2px 0 0 var(--color-brand);
}
.ssi-poster {
  width: 40px;
  height: 60px;
  flex-shrink: 0;
  object-fit: cover;
  border-radius: 4px;
  background: var(--color-surface-3);
}
.ssi-poster--empty {
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--color-text-muted);
  border: 1px solid var(--color-border);
}
.ssi-body {
  display: flex;
  flex-direction: column;
  gap: 4px;
  min-width: 0;
  flex: 1;
}
.ssi-title {
  overflow: hidden;
  white-space: nowrap;
  text-overflow: ellipsis;
  font-size: 0.875rem;
  font-weight: 600;
  line-height: 1.3;
  color: var(--color-text-primary);
}
.ssi-mark {
  background: none;
  color: var(--color-accent);
}
.ssi-meta {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 4px 8px;
  font-size: 0.75rem;
  line-height: 1;
  color: var(--color-text-secondary);
}
.ssi-chip {
  padding: 3px 6px;
  border-radius: 4px;
  border: 1px solid var(--color-border-strong);
  font-size: 0.6875rem;
  font-weight: 600;
  letter-spacing: 0.02em;
  color: var(--color-text-secondary);
}
.ssi-chip[data-type="tv"] {
  color: var(--color-info);
  border-color: color-mix(in srgb, var(--color-info) 40%, transparent);
}
.ssi-tmdb {
  font-weight: 600;
  letter-spacing: 0.02em;
  font-variant-numeric: tabular-nums;
}
.ssi-popcorn {
  display: inline-flex;
  align-items: center;
  gap: 3px;
  font-weight: 700;
  font-variant-numeric: tabular-nums;
}
.ssi-popcorn[data-tier="none"] {
  font-weight: 400;
  color: var(--color-text-muted);
}
.ssi-popcorn[data-tier="stale"] {
  color: var(--color-text-secondary);
}
.ssi-popcorn[data-tier="fresh"] {
  color: var(--color-warning);
}
.ssi-popcorn[data-tier="hot"] {
  color: var(--color-text-primary);
}
.ssi-popcorn[data-tier="hot"] svg {
  color: var(--color-brand);
}
@media (max-width: 767px) {
  .ssi {
    min-height: 76px;
  }
}
</style>
