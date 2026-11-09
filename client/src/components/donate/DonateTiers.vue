<template>
  <section class="tiers" aria-labelledby="donate-tiers-title">
    <h2 id="donate-tiers-title" class="tiers__title">{{ t("donate.tiers.title") }}</h2>
    <p class="tiers__hint">{{ t("donate.tiers.hint") }}</p>
    <div class="tiers__grid" role="group" :aria-label="t('donate.tiers.title')">
      <button
        v-for="tier in TIERS"
        :key="tier.key"
        type="button"
        class="tier"
        :class="{ 'is-active': model === tier.amount }"
        :aria-pressed="model === tier.amount"
        @click="model = model === tier.amount ? null : tier.amount"
      >
        <component :is="tier.icon" :size="26" class="tier__icon" aria-hidden="true" />
        <span class="tier__amount">{{ t("donate.tiers.amount", { amount: tier.amount.toLocaleString(locale) }) }}</span>
        <span class="tier__name">{{ t(`donate.tiers.${tier.key}.name`) }}</span>
        <span class="tier__desc">{{ t(`donate.tiers.${tier.key}.desc`) }}</span>
      </button>
    </div>
  </section>
</template>

<script setup lang="ts">
import { useI18n } from "vue-i18n"
import { Clapperboard, Coffee, Popcorn } from "lucide-vue-next"

/** Suggested amounts in THB; PromptPay and bank transfers accept any amount. */
const TIERS = [
  { key: "coffee", amount: 50, icon: Coffee },
  { key: "popcorn", amount: 100, icon: Popcorn },
  { key: "movie", amount: 300, icon: Clapperboard },
] as const

const model = defineModel<number | null>({ default: null })
const { t, locale } = useI18n()
</script>

<style scoped>
.tiers {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}
.tiers__title {
  margin: 0;
  font-size: var(--text-heading);
  font-weight: 700;
  color: var(--color-text-primary);
}
.tiers__hint {
  margin: 0 0 0.75rem;
  font-size: var(--text-small);
  color: var(--color-text-secondary);
}
.tiers__grid {
  display: grid;
  grid-template-columns: minmax(0, 1fr);
  gap: 0.75rem;
}
.tier {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  gap: 0.25rem;
  padding: 1.25rem;
  border: 1px solid var(--color-border);
  border-radius: var(--radius-card);
  background: var(--color-surface-2);
  color: var(--color-text-primary);
  font: inherit;
  text-align: start;
  cursor: pointer;
  transition:
    border-color var(--motion-fast) var(--ease-out-soft),
    background var(--motion-fast) var(--ease-out-soft),
    transform var(--motion-fast) var(--ease-out-soft);
}
.tier__icon {
  margin-bottom: 0.5rem;
  color: var(--color-accent);
}
.tier__amount {
  font-size: var(--text-heading);
  font-weight: 800;
  line-height: 1.2;
}
.tier__name {
  font-size: var(--text-small);
  font-weight: 600;
}
.tier__desc {
  font-size: var(--text-caption);
  color: var(--color-text-muted);
}
.tier.is-active {
  border-color: var(--color-brand);
  background: color-mix(in srgb, var(--color-brand) 12%, var(--color-surface-2));
}
@media (hover: hover) {
  .tier:hover {
    border-color: var(--color-border-strong);
    transform: translateY(-2px);
  }
  .tier.is-active:hover {
    border-color: var(--color-brand);
  }
}
@media (min-width: 640px) {
  .tiers__grid {
    grid-template-columns: repeat(3, minmax(0, 1fr));
  }
}
@media (prefers-reduced-motion: reduce) {
  .tier {
    transition: none;
  }
  .tier:hover {
    transform: none;
  }
}
</style>
