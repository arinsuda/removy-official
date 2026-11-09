<template>
  <header class="lh">
    <RouterLink :to="{ name: 'user-profile', params: { userId } }" class="lh__back"><ChevronLeft :size="16" aria-hidden="true" />{{ t("navigation.profile") }}</RouterLink>
    <div class="lh__main">
      <AppAvatar :src="profile?.avatar_url" :name="displayName" size="lg" decorative />
      <div class="lh__text">
        <h1 class="lh__title">{{ isOwner ? t("library.title.own") : t("library.title.other", { name: displayName }) }}</h1>
        <p class="lh__sub">{{ isOwner ? t("library.subtitle.own") : t("library.subtitle.other", { name: displayName }) }}</p>
      </div>
    </div>
    <ul class="lh__chips" :aria-label="t('library.chips.label')">
      <li v-for="c in chips" :key="c.key">
        <RouterLink :to="listTo(c.key)" class="lh__chip">
          <component :is="LIST_ICONS[c.key]" :size="14" aria-hidden="true" />
          <span>{{ t(`library.lists.names.${c.key}`) }}</span>
          <span class="lh__chip-n">{{ n(c.count) }}</span>
        </RouterLink>
      </li>
    </ul>
  </header>
</template>

<script setup lang="ts">
import { computed } from "vue"
import { useI18n } from "vue-i18n"
import { RouterLink, type RouteLocationRaw } from "vue-router"
import { ChevronLeft } from "lucide-vue-next"
import AppAvatar from "@/components/ui/AppAvatar.vue"
import type { LibraryListKey } from "@/types"
import type { UserProfile } from "@/types/user"
import { LIBRARY_LIST_KEYS, LIST_ICONS } from "../config"

const props = defineProps<{
  userId: number
  profile: UserProfile | null
  isOwner: boolean
  counts: Record<LibraryListKey, number>
  listTo: (key: LibraryListKey) => RouteLocationRaw
}>()

const { t, n } = useI18n()
const displayName = computed(() => props.profile?.display_name || props.profile?.username || "")
const chips = computed(() => LIBRARY_LIST_KEYS.map((key) => ({ key, count: props.counts[key] })))
</script>

<style scoped>
.lh { display: flex; flex-direction: column; gap: 1rem; padding: 1rem 0 0; }
.lh__back { display: inline-flex; align-items: center; align-self: flex-start; gap: 0.25rem; min-height: 2.75rem; font-size: var(--text-small); color: var(--color-text-secondary); text-decoration: none; }
.lh__main { display: flex; align-items: center; gap: 1rem; min-width: 0; }
.lh__text { min-width: 0; }
.lh__title { margin: 0; font-size: var(--text-heading); font-weight: 700; line-height: 1.2; color: var(--color-text-primary); }
.lh__sub { margin: 0.25rem 0 0; font-size: var(--text-small); color: var(--color-text-secondary); }
.lh__chips { display: flex; gap: 0.5rem; margin: 0; padding: 0 0 0.25rem; overflow-x: auto; list-style: none; scrollbar-width: none; }
.lh__chips::-webkit-scrollbar { display: none; }
.lh__chip { display: inline-flex; align-items: center; gap: 0.5rem; min-height: 2.75rem; padding: 0 1rem; font-size: var(--text-caption); color: var(--color-text-secondary); text-decoration: none; white-space: nowrap; background: var(--color-surface-2); border: 1px solid var(--color-border); border-radius: var(--radius-btn); transition: border-color var(--motion-fast) var(--ease-water), color var(--motion-fast) var(--ease-water); }
.lh__chip-n { font-weight: 700; color: var(--color-text-primary); }
@media (hover: hover) {
  .lh__back:hover, .lh__chip:hover { color: var(--color-text-primary); }
  .lh__chip:hover { border-color: var(--color-border-strong); }
}
@media (min-width: 768px) {
  .lh__title { font-size: var(--text-page); }
}
@media (prefers-reduced-motion: reduce) {
  .lh__chip { transition: none; }
}
</style>
