<template>
  <section class="faq" aria-labelledby="donate-faq-title">
    <h2 id="donate-faq-title" class="faq__title">{{ t("donate.faq.title") }}</h2>
    <div class="faq__list">
      <div v-for="key in KEYS" :key="key" class="faq__item" :class="{ 'is-open': open === key }">
        <h3 class="faq__q">
          <button
            :id="`faq-${key}-btn`"
            type="button"
            class="faq__btn"
            :aria-expanded="open === key"
            :aria-controls="`faq-${key}-panel`"
            @click="open = open === key ? null : key"
          >
            <span>{{ t(`donate.faq.${key}.q`) }}</span>
            <ChevronDown :size="18" class="faq__chevron" aria-hidden="true" />
          </button>
        </h3>
        <div :id="`faq-${key}-panel`" class="faq__panel" role="region" :aria-labelledby="`faq-${key}-btn`">
          <div class="faq__panel-inner">
            <p class="faq__a">{{ t(`donate.faq.${key}.a`) }}</p>
          </div>
        </div>
      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
import { ref } from "vue"
import { useI18n } from "vue-i18n"
import { ChevronDown } from "lucide-vue-next"

const KEYS = ["required", "safe", "amount", "use"] as const
const { t } = useI18n()
const open = ref<(typeof KEYS)[number] | null>(null)
</script>

<style scoped>
.faq {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}
.faq__title {
  margin: 0;
  font-size: var(--text-heading);
  font-weight: 700;
  color: var(--color-text-primary);
}
.faq__list {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}
.faq__item {
  border: 1px solid var(--color-border);
  border-radius: var(--radius-card);
  background: var(--color-surface-2);
}
.faq__q {
  margin: 0;
}
.faq__btn {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
  width: 100%;
  min-height: 3rem;
  padding: 0.75rem 1rem;
  border: 0;
  border-radius: var(--radius-card);
  background: transparent;
  color: var(--color-text-primary);
  font: inherit;
  font-weight: 600;
  text-align: start;
  cursor: pointer;
}
.faq__chevron {
  flex: none;
  color: var(--color-text-muted);
  transition: transform var(--motion-base) var(--ease-out-soft);
}
.is-open .faq__chevron {
  transform: rotate(180deg);
}
/* grid-template-rows animates height without measuring the content */
.faq__panel {
  display: grid;
  grid-template-rows: 0fr;
  transition: grid-template-rows var(--motion-base) var(--ease-out-soft);
}
.is-open .faq__panel {
  grid-template-rows: 1fr;
}
.faq__panel-inner {
  min-height: 0;
  overflow: hidden;
}
.faq__a {
  margin: 0;
  padding: 0 1rem 1rem;
  font-size: var(--text-small);
  color: var(--color-text-secondary);
}
.faq__item:not(.is-open) .faq__panel-inner {
  visibility: hidden;
}
@media (hover: hover) {
  .faq__btn:hover {
    background: var(--color-surface-3);
  }
}
@media (prefers-reduced-motion: reduce) {
  .faq__chevron,
  .faq__panel {
    transition: none;
  }
}
</style>
