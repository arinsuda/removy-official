<template>
  <div class="tiers">
    <div class="tiers__chips" role="group" :aria-label="t('browse.tier.label')">
      <button
        v-for="tier in BROWSE_TIERS"
        :key="tier"
        type="button"
        class="tier-chip"
        :class="[`tier-chip--${tier}`, { 'is-active': modelValue === tier }]"
        :aria-pressed="modelValue === tier"
        @click="emit('update:modelValue', tier)"
      >
        {{ t(`browse.tier.${tier}`) }}
        <span class="tier-chip__count">{{ countOf(tier) }}</span>
      </button>
    </div>
    <p class="tiers__caption">
      {{ t("browse.tier.caption", { n: total }) }}
      <span v-if="searching" role="status">{{ t("browse.tier.loading") }}</span>
    </p>
  </div>
</template>

<script setup lang="ts">
import { computed } from "vue"
import { useI18n } from "vue-i18n"
import type { PopcornTier } from "@/components/movie/popcornTier"
import { BROWSE_TIERS, type BrowseTierFilter } from "./browseQuery"

const props = defineProps<{
  modelValue: BrowseTierFilter
  counts: Record<PopcornTier, number>
  /** True while extra pages are being pulled in to find matches. */
  searching: boolean
}>()
const emit = defineEmits<{ "update:modelValue": [tier: BrowseTierFilter] }>()

const { t } = useI18n()
const total = computed(() => Object.values(props.counts).reduce((a, b) => a + b, 0))
const countOf = (tier: BrowseTierFilter) => (tier === "all" ? total.value : props.counts[tier])
</script>

<style scoped>
.tiers {
  display: flex;
  flex-direction: column;
  gap: 0.375rem;
}
.tiers__chips {
  display: flex;
  gap: 0.5rem;
  overflow-x: auto;
  scrollbar-width: none;
}
.tiers__chips::-webkit-scrollbar {
  display: none;
}
.tier-chip {
  display: inline-flex;
  flex: none;
  align-items: center;
  gap: 0.375rem;
  min-height: 2rem;
  padding: 0 0.75rem;
  border-radius: 9999px;
  border: 1px solid var(--color-border-strong);
  background: var(--color-surface-2);
  color: var(--color-text-secondary);
  font: inherit;
  font-size: var(--text-caption);
  white-space: nowrap;
  cursor: pointer;
}
.tier-chip:hover {
  background: var(--color-surface-3);
}
.tier-chip--hot {
  border-color: color-mix(in srgb, var(--color-brand) 60%, transparent);
}
.tier-chip--none {
  border-style: dashed;
}
.tier-chip.is-active {
  background: var(--color-text-primary);
  border-color: var(--color-text-primary);
  color: var(--color-surface);
}
.tier-chip__count {
  font-variant-numeric: tabular-nums;
  opacity: 0.7;
}
.tiers__caption {
  margin: 0;
  font-size: var(--text-caption);
  color: var(--color-text-muted);
}
</style>
