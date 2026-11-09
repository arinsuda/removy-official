<template>
  <RouterLink :to="to" class="cl" data-custom-list>
    <div class="cl__cover" :class="{ 'cl__cover--single': !!coverUrl }" aria-hidden="true">
      <img v-if="coverUrl" :src="coverUrl" alt="" class="cl__poster" loading="lazy" decoding="async" />
      <template v-else-if="list.collage_posters.length">
        <img v-for="(p, i) in list.collage_posters" :key="`${p}-${i}`" :src="posterUrl(p)" alt="" class="cl__poster" loading="lazy" decoding="async" />
      </template>
      <ListIcon v-else :size="40" class="cl__empty-icon" />
    </div>
    <div class="cl__body">
      <h3 class="cl__name">{{ list.name }}</h3>
      <p v-if="list.description" class="cl__desc">{{ list.description }}</p>
      <p class="cl__meta">
        <span class="cl__count">{{ t("library.lists.count", { n: list.item_count }, list.item_count) }}</span>
        <span class="cl__vis" data-visibility :data-value="list.visibility">
          <component :is="LIST_VISIBILITY_ICONS[list.visibility]" :size="12" aria-hidden="true" />{{ t(`visibility.${list.visibility}`) }}
        </span>
      </p>
    </div>
  </RouterLink>
</template>

<script setup lang="ts">
import { computed } from "vue"
import { RouterLink, type RouteLocationRaw } from "vue-router"
import { useI18n } from "vue-i18n"
import { List as ListIcon } from "lucide-vue-next"
import { getTmdbImageUrl } from "@/utils/image"
import type { ListSummary } from "@/types"
import { COLLAGE_POSTER_SIZE, COVER_POSTER_SIZE, LIST_VISIBILITY_ICONS } from "../config"

const props = defineProps<{ list: ListSummary; to: RouteLocationRaw }>()
const { t } = useI18n()
const posterUrl = (path: string) => getTmdbImageUrl(path, COLLAGE_POSTER_SIZE) ?? ""
const coverUrl = computed(() => (props.list.cover_poster ? (getTmdbImageUrl(props.list.cover_poster, COVER_POSTER_SIZE) ?? "") : ""))
</script>

<style scoped>
.cl { display: flex; flex-direction: column; min-width: 0; overflow: hidden; color: inherit; text-decoration: none; background: var(--color-surface-2); border: 1px solid var(--color-border); border-radius: var(--radius-card); transition: border-color var(--motion-fast) var(--ease-water), transform var(--motion-base) var(--ease-out-soft); }
.cl__cover { display: grid; grid-template-columns: repeat(4, minmax(0, 1fr)); aspect-ratio: 2 / 1; background: var(--color-surface-3); }
.cl__cover { position: relative; overflow: hidden; }
.cl__cover--single { grid-template-columns: minmax(0, 1fr); }
.cl__cover--single .cl__poster { position: absolute; inset: 0; }
.cl__poster { width: 100%; height: 100%; object-fit: cover; }
.cl__empty-icon { grid-column: 1 / -1; align-self: center; justify-self: center; color: var(--color-text-muted); }
.cl__body { display: flex; flex-direction: column; gap: 0.25rem; padding: 1rem; }
.cl__name { margin: 0; overflow: hidden; font-size: var(--text-body); font-weight: 700; color: var(--color-text-primary); text-overflow: ellipsis; white-space: nowrap; }
.cl__desc { display: -webkit-box; margin: 0; overflow: hidden; font-size: var(--text-caption); color: var(--color-text-muted); -webkit-line-clamp: 2; -webkit-box-orient: vertical; }
.cl__meta { display: flex; flex-wrap: wrap; justify-content: space-between; gap: 0.25rem 0.75rem; margin: 0.5rem 0 0; font-size: var(--text-caption); color: var(--color-text-secondary); }
.cl__count { font-weight: 600; color: var(--color-text-primary); }
.cl__vis { display: inline-flex; align-items: center; gap: 0.25rem; }
@media (hover: hover) {
  .cl:hover { border-color: var(--color-border-strong); transform: translateY(-2px); }
}
@media (prefers-reduced-motion: reduce) {
  .cl { transition: none; }
  .cl:hover { transform: none; }
}
</style>
