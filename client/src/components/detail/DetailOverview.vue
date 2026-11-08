<template>
  <section class="overview" :aria-labelledby="titleId">
    <h2 :id="titleId" class="detail-heading">{{ t("media.detail.overview") }}</h2>
    <template v-if="text">
      <p ref="body" class="overview__text" :class="{ 'overview__text--clamped': !expanded }">{{ text }}</p>
      <button v-if="expanded || overflowing" type="button" class="overview__toggle" :aria-expanded="expanded" @click="expanded = !expanded">
        {{ t(expanded ? "media.detail.showLess" : "media.detail.readMore") }}
      </button>
    </template>
    <p v-else class="overview__empty">{{ t("media.detail.noOverview") }}</p>
  </section>
</template>

<script setup lang="ts">
import { nextTick, onMounted, ref, useId, useTemplateRef, watch } from "vue"
import { useI18n } from "vue-i18n"

const props = defineProps<{ text: string }>()
const { t } = useI18n()
const titleId = useId()
const body = useTemplateRef<HTMLElement>("body")
const expanded = ref(false)
const overflowing = ref(false)

function measure() {
  const el = body.value
  overflowing.value = Boolean(el && el.scrollHeight > el.clientHeight + 1)
}

onMounted(measure)
watch(() => props.text, async () => {
  expanded.value = false
  await nextTick()
  measure()
})
</script>

<style scoped>
.overview__text { margin-top: 0.75rem; font-size: var(--text-body); line-height: 1.7; color: var(--color-text-secondary); white-space: pre-line; }
.overview__text--clamped { display: -webkit-box; -webkit-box-orient: vertical; -webkit-line-clamp: 5; line-clamp: 5; overflow: hidden; }
.overview__empty { margin-top: 0.75rem; font-size: var(--text-small); color: var(--color-text-muted); }
.overview__toggle { margin-top: 0.4rem; padding: 0.5rem 0; font-size: var(--text-small); font-weight: 600; color: var(--color-brand); background: none; border: 0; cursor: pointer; }
.overview__toggle:hover { text-decoration: underline; text-underline-offset: 3px; }
</style>
