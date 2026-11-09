<template>
  <div class="wt">
    <section v-for="g in groups" :key="g.key" class="wt__group">
      <h3 v-if="g.label" class="wt__month">
        <span>{{ g.label }}</span>
        <span class="wt__month-n">{{ t("profile.tabContent.watched.monthTitles", { n: g.items.length }) }}</span>
      </h3>
      <ul class="wt__list">
        <li v-for="e in g.items" :key="e.media.key" class="wt__row">
          <RouterLink :to="detailRoute(e.media)" class="wt__poster" :aria-label="e.media.title">
            <img v-if="e.media.poster_path" :src="posterSrc(e.media.poster_path)" alt="" loading="lazy" decoding="async" />
            <Film v-else :size="18" aria-hidden="true" />
          </RouterLink>
          <div class="wt__body">
            <RouterLink :to="detailRoute(e.media)" class="wt__title">{{ e.media.title }}</RouterLink>
            <p class="wt__meta">
              <span class="wt__type">{{ e.media.type === "tv" ? t("profile.tabContent.common.typeTv") : t("profile.tabContent.common.typeMovie") }}</span>
              <time :datetime="e.at">{{ formatDate(e.at) }}</time>
            </p>
            <p class="wt__chips">
              <span v-if="e.count && e.count > 1" class="wt__chip wt__chip--count" :title="t('profile.tabContent.watched.watchedTimes', { n: e.count })">
                <Eye :size="12" aria-hidden="true" />{{ t("profile.tabContent.watched.times", { n: e.count }) }}
              </span>
              <span v-if="e.rating" class="wt__chip wt__chip--rating" :data-tier="popcornTier(e.rating, 1)" :aria-label="t('profile.tabContent.watched.yourRating', { rating: e.rating })">
                <Star :size="12" fill="currentColor" aria-hidden="true" />{{ e.rating }}
              </span>
            </p>
          </div>
          <div v-if="isOwner" class="wt__actions">
            <slot name="actions" :entry="e" />
          </div>
        </li>
      </ul>
    </section>
  </div>
</template>

<script setup lang="ts">
import { Eye, Film, Star } from "lucide-vue-next"
import { useI18n } from "vue-i18n"
import { RouterLink } from "vue-router"
import { detailRoute, posterSrc } from "@/components/home/homeMedia"
import { popcornTier } from "@/components/movie/popcornTier"
import type { MonthGroup, WatchedEntry } from "./watched"

defineProps<{ groups: MonthGroup[]; isOwner: boolean }>()
defineSlots<{ actions?: (p: { entry: WatchedEntry }) => unknown }>()

const { t, locale } = useI18n()
const formatDate = (iso: string) => new Intl.DateTimeFormat(locale.value, { dateStyle: "medium" }).format(new Date(iso))
</script>

<style scoped>
.wt { display: flex; flex-direction: column; gap: 1.5rem; }
/* Sticks under the sticky site header (100) so months stay readable while scrolling. */
.wt__month { position: sticky; top: 3.5rem; z-index: 2; display: flex; align-items: baseline; justify-content: space-between; gap: 0.5rem; margin: 0 0 0.5rem; padding: 0.5rem 0; font-size: var(--text-small); font-weight: 700; color: var(--color-text-primary); background: var(--color-surface); border-bottom: 1px solid var(--color-border); }
.wt__month-n { font-size: var(--text-caption); font-weight: 500; color: var(--color-text-muted); }
.wt__list { display: flex; flex-direction: column; gap: 0.5rem; margin: 0; padding: 0; list-style: none; }
.wt__row { display: flex; align-items: center; gap: 0.75rem; padding: 0.5rem; background: var(--color-surface-2); border: 1px solid var(--color-border); border-radius: var(--radius-card); transition: border-color var(--motion-fast) var(--ease-water); }
.wt__poster { display: grid; flex: none; place-items: center; width: 3.25rem; aspect-ratio: 2 / 3; overflow: hidden; color: var(--color-text-muted); background: var(--color-surface-3); border-radius: 0.5rem; }
.wt__poster img { width: 100%; height: 100%; object-fit: cover; }
.wt__body { display: flex; flex: 1; flex-direction: column; gap: 0.25rem; min-width: 0; }
.wt__title { overflow: hidden; font-size: var(--text-small); font-weight: 600; line-height: 1.3; color: var(--color-text-primary); text-decoration: none; text-overflow: ellipsis; white-space: nowrap; }
.wt__meta { display: flex; flex-wrap: wrap; align-items: center; gap: 0.25rem 0.5rem; margin: 0; font-size: var(--text-caption); color: var(--color-text-muted); }
.wt__type { padding: 0 0.4rem; font-size: 0.6875rem; font-weight: 600; color: var(--color-text-secondary); background: var(--color-surface-3); border-radius: 0.25rem; }
.wt__chips { display: flex; flex-wrap: wrap; gap: 0.375rem; margin: 0; }
.wt__chips:empty { display: none; }
.wt__chip { display: inline-flex; align-items: center; gap: 0.25rem; padding: 0.0625rem 0.5rem; font-size: var(--text-caption); font-weight: 700; border-radius: 9999px; }
.wt__chip--count { color: var(--color-text-secondary); background: var(--color-surface-3); }
.wt__chip--rating { color: var(--color-text-secondary); border: 1px solid var(--color-border-strong); }
.wt__chip--rating[data-tier="hot"] { color: var(--color-accent); border-color: var(--color-brand); background: color-mix(in srgb, var(--color-brand) 18%, transparent); }
.wt__chip--rating[data-tier="stale"] { color: var(--color-text-muted); }
.wt__actions { display: flex; flex: none; align-items: center; }
@media (min-width: 768px) { .wt__month { top: 4rem; } }
@media (hover: hover) {
  .wt__row:hover { border-color: var(--color-border-strong); }
  .wt__title:hover { color: var(--color-brand); }
}
@media (prefers-reduced-motion: reduce) { .wt__row { transition: none; } }
</style>
