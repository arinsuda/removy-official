<template>
  <section class="facts" :aria-labelledby="titleId">
    <h2 :id="titleId" class="facts__title">{{ t("media.detail.facts.title") }}</h2>
    <dl class="facts__list">
      <div v-for="f in facts" :key="f.key" class="facts__row">
        <dt>{{ t(`media.detail.facts.${f.key}`) }}</dt>
        <dd>{{ display(f) }}</dd>
      </div>
      <div v-if="safeHomepage" class="facts__row">
        <dt>{{ t("media.detail.facts.homepage") }}</dt>
        <dd>
          <a :href="safeHomepage" target="_blank" rel="noopener noreferrer" class="facts__link">{{ host }}</a>
        </dd>
      </div>
    </dl>
  </section>
</template>

<script setup lang="ts">
import { computed, useId } from "vue"
import { useI18n } from "vue-i18n"
import { MONEY_CURRENCY, statusKey, type DetailFact } from "./detailConfig"

const props = defineProps<{ facts: DetailFact[]; homepage?: string }>()
const { t, te, locale } = useI18n()
const titleId = useId()

/** Only http(s) links are rendered; anything else (javascript:, data:, junk) is dropped. */
const safeHomepage = computed(() => {
  try {
    const u = new URL(props.homepage ?? "")
    return u.protocol === "http:" || u.protocol === "https:" ? u.href : ""
  } catch {
    return ""
  }
})

const host = computed(() => (safeHomepage.value ? new URL(safeHomepage.value).hostname.replace(/^www\./, "") : ""))

function display(f: DetailFact): string {
  if (f.kind === "status") return te(`media.detail.status.${statusKey(f.value)}`) ? t(`media.detail.status.${statusKey(f.value)}`) : f.value
  if (f.kind === "date") {
    const d = new Date(`${f.value.slice(0, 10)}T00:00:00`)
    return Number.isNaN(d.getTime()) ? f.value : new Intl.DateTimeFormat(locale.value, { dateStyle: "long" }).format(d)
  }
  if (f.kind === "money") {
    return new Intl.NumberFormat(locale.value, { style: "currency", currency: MONEY_CURRENCY, maximumFractionDigits: 0 }).format(Number(f.value))
  }
  return f.value
}
</script>

<style scoped>
.facts { padding: 1.25rem; background: var(--color-surface-2); border: 1px solid var(--color-border); border-radius: var(--radius-card); }
.facts__title { font-size: var(--text-small); font-weight: 700; color: var(--color-text-primary); }
.facts__list { display: grid; grid-template-columns: minmax(0, 1fr); gap: 0.75rem 1.5rem; margin-top: 0.75rem; }
.facts__row { display: flex; flex-direction: column; gap: 0.1rem; min-width: 0; }
.facts__row dt { font-size: var(--text-caption); font-weight: 700; letter-spacing: 0.04em; text-transform: uppercase; color: var(--color-text-muted); }
.facts__row dd { margin: 0; font-size: var(--text-small); color: var(--color-text-primary); overflow-wrap: anywhere; }
.facts__link { color: var(--color-info); text-decoration: underline; text-underline-offset: 3px; }
@media (min-width: 768px) {
  .facts__list { grid-template-columns: repeat(2, minmax(0, 1fr)); }
}
@media (min-width: 1024px) {
  .facts__list { grid-template-columns: minmax(0, 1fr); }
}
</style>
