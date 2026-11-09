<template>
  <section v-if="items.length" class="bmol" :aria-labelledby="titleId">
    <div class="bmol__head">
      <h2 :id="titleId" class="bmol__title">{{ t("profile.bmol.title") }}</h2>
      <RouterLink :to="{ query: { tab: 'bol' } }" class="bmol__more">{{ t("profile.bmol.viewAll") }}</RouterLink>
    </div>
    <ol class="bmol__list">
      <li v-for="item in items" :key="item.id" class="bmol__item">
        <span class="bmol__rank">{{ item.rank }}</span>
        <img v-if="item.media.poster_url" :src="item.media.poster_url" alt="" class="bmol__poster" loading="lazy" />
        <span class="bmol__name">{{ item.media.title || item.media.name }}</span>
      </li>
    </ol>
  </section>
</template>

<script setup lang="ts">
import { computed, useId } from "vue"
import { useI18n } from "vue-i18n"
import { useQuery } from "@tanstack/vue-query"
import { bmolApi } from "@/api/api"

const props = defineProps<{ userId: number }>()
const { t } = useI18n()
const titleId = useId()

// A 403 (private account) just hides the highlight; the tab shows the locked state.
const query = useQuery({
  queryKey: computed(() => ["bmol-highlight", props.userId]),
  queryFn: () => bmolApi.getUserBMOL(props.userId).then(res => res.data.items ?? []),
  retry: false,
})
const items = computed(() => [...(query.data.value ?? [])].sort((a, b) => a.rank - b.rank).slice(0, 3))
</script>

<style scoped>
.bmol { padding: 1.25rem; background: var(--color-surface-2); border: 1px solid var(--color-border); border-radius: var(--radius-card); }
.bmol__head { display: flex; align-items: baseline; justify-content: space-between; }
.bmol__title { font-size: var(--text-small); font-weight: 700; color: var(--color-text-primary); }
.bmol__more { font-size: var(--text-caption); color: var(--color-text-secondary); }
.bmol__more:hover { color: var(--color-text-primary); }
.bmol__list { display: grid; grid-template-columns: repeat(3, minmax(0, 1fr)); gap: 0.75rem; margin-top: 0.75rem; padding: 0; list-style: none; }
.bmol__item { position: relative; display: flex; flex-direction: column; gap: 0.375rem; min-width: 0; }
.bmol__rank { position: absolute; top: 0.25rem; left: 0.25rem; z-index: 1; display: grid; place-items: center; width: 1.375rem; height: 1.375rem; font-size: 0.75rem; font-weight: 700; color: var(--color-on-brand); background: var(--color-brand); border-radius: 999px; }
.bmol__poster { width: 100%; aspect-ratio: 2 / 3; object-fit: cover; background: var(--color-surface-3); border-radius: var(--radius-btn); }
.bmol__name { overflow: hidden; font-size: var(--text-caption); color: var(--color-text-secondary); text-overflow: ellipsis; white-space: nowrap; }
</style>
