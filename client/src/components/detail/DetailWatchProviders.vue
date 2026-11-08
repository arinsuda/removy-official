<template>
  <section class="providers" :aria-labelledby="titleId" :aria-busy="query.isLoading.value">
    <div class="providers__head">
      <h2 :id="titleId" class="providers__title">{{ t("media.detail.providers.title") }}</h2>
      <label class="providers__region">
        <span class="sr-only">{{ t("media.detail.providers.region") }}</span>
        <select :value="region" class="providers__select" @change="onRegion">
          <option v-for="r in PROVIDER_REGIONS" :key="r" :value="r">{{ r }}</option>
        </select>
      </label>
    </div>

    <div v-if="query.isLoading.value" class="providers__skeleton" data-test="providers-loading">
      <SkeletonBlock v-for="n in 4" :key="n" width="40px" height="40px" radius="8px" />
    </div>

    <AppErrorState
      v-else-if="query.isError.value"
      class="providers__error"
      :title="t('media.detail.providers.error')"
      @retry="query.refetch()"
    />

    <p v-else-if="groups.length === 0" class="providers__empty" data-test="providers-empty">
      {{ t("media.detail.providers.empty", { region }) }}
    </p>

    <template v-else>
      <component
        :is="href ? 'a' : 'div'"
        v-for="g in groups"
        :key="g.key"
        class="providers__group"
        data-test="providers-group"
        v-bind="href ? { href, target: '_blank', rel: 'noopener noreferrer', 'aria-label': `${g.label}: ${t('media.detail.providers.open')}` } : {}"
      >
        <h3 class="providers__label">{{ g.label }}</h3>
        <ul class="providers__list" :aria-label="g.label">
          <li v-for="p in g.items" :key="p.provider_id" class="providers__item" :title="p.provider_name">
            <img :src="tmdbImage(p.logo_path, 'w92')" :alt="p.provider_name" width="40" height="40" loading="lazy" class="providers__logo" />
          </li>
        </ul>
      </component>
    </template>

    <p class="providers__attribution">{{ t("media.detail.providers.attribution") }}</p>
  </section>
</template>

<script setup lang="ts">
import { computed, useId } from "vue"
import { useI18n } from "vue-i18n"
import AppErrorState from "@/components/ui/AppErrorState.vue"
import SkeletonBlock from "@/components/ui/skeleton/SkeletonBlock.vue"
import { useImageUrl } from "@/composables/useImageUrl"
import { PROVIDER_REGIONS, setWatchRegion, useWatchProviders } from "@/composables/useWatchProviders"
import type { WatchProvider } from "@/types/movie"

const props = defineProps<{ type: "movie" | "tv"; id: number }>()
const { t } = useI18n()
const { tmdbImage } = useImageUrl()
const titleId = useId()

const { region, query } = useWatchProviders(props.type, computed(() => props.id))

const groups = computed(() => {
  const d = query.data.value
  if (!d) return []
  const defs: { key: string; items: WatchProvider[] }[] = [
    { key: "stream", items: d.flatrate },
    { key: "free", items: [...d.free, ...d.ads] },
    { key: "rent", items: d.rent },
    { key: "buy", items: d.buy },
  ]
  return defs
    .filter(g => g.items.length > 0)
    .map(g => ({ ...g, label: t(`media.detail.providers.${g.key}`) }))
})

/** Only https links to the provider page are used. */
const href = computed(() => {
  try {
    const u = new URL(query.data.value?.link ?? "")
    return u.protocol === "https:" ? u.href : ""
  } catch {
    return ""
  }
})

function onRegion(e: Event) {
  setWatchRegion((e.target as HTMLSelectElement).value)
}
</script>

<style scoped>
.providers { padding: 1.25rem; background: var(--color-surface-2); border: 1px solid var(--color-border); border-radius: var(--radius-card); }
.providers__head { display: flex; align-items: center; justify-content: space-between; gap: 0.75rem; }
.providers__title { font-size: var(--text-small); font-weight: 700; color: var(--color-text-primary); }
.providers__select { padding: 0.25rem 0.5rem; font-size: var(--text-caption); color: var(--color-text-primary); background: var(--color-surface-3); border: 1px solid var(--color-border-strong); border-radius: var(--radius-btn); }
.providers__skeleton { display: flex; gap: 0.5rem; margin-top: 0.75rem; }
.providers__error :deep(.app-state) { padding: 1rem 0; }
.providers__empty { margin-top: 0.75rem; font-size: var(--text-small); color: var(--color-text-secondary); }
.providers__group { display: block; margin-top: 0.75rem; color: inherit; text-decoration: none; border-radius: var(--radius-btn); }
a.providers__group:hover .providers__logo { border-color: var(--color-border-strong); }
.providers__label { font-size: var(--text-caption); font-weight: 700; letter-spacing: 0.04em; text-transform: uppercase; color: var(--color-text-muted); }
.providers__list { display: flex; flex-wrap: wrap; gap: 0.5rem; margin: 0.4rem 0 0; padding: 0; list-style: none; }
.providers__logo { width: 40px; height: 40px; object-fit: cover; border-radius: 8px; border: 1px solid var(--color-border); }
.providers__attribution { margin-top: 1rem; font-size: var(--text-caption); color: var(--color-text-muted); }
</style>
