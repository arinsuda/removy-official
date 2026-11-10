<template>
  <article class="sc-card sc-card--story ach-st" :class="`sc-theme-${theme}`" data-template="achievement-story">
    <header class="ach-st__head">
      <span class="sc-chip">{{ t("achievements.share.badge") }}</span>
      <div v-if="username" class="ach-st__user" data-test="share-user">
        <img v-if="avatar" :src="avatar" :alt="username" class="ach-st__avatar" />
        <span v-else class="ach-st__initials" aria-hidden="true">{{ initials }}</span>
        <span class="sc-meta ach-st__handle">@{{ username }}</span>
      </div>
    </header>

    <section class="ach-st__hero">
      <div class="ach-st__halo">
        <div class="ach-st__emblem">
          <component :is="icon" :size="164" :stroke-width="1.75" aria-hidden="true" />
        </div>
      </div>

      <div class="ach-st__tags">
        <span class="ach-st__tag">{{ t(`achievements.categories.${category}`) }}</span>
        <span v-if="tierLabel" class="ach-st__tag ach-st__tag--tier" data-test="share-tier">{{ tierLabel }}</span>
        <span v-if="achievement.achievement.exp_reward" class="ach-st__tag ach-st__tag--xp">
          {{ t("achievements.xp", { n: achievement.achievement.exp_reward }) }}
        </span>
      </div>

      <h1 class="sc-title ach-st__title" :class="titleSizeClass(achName(achievement.achievement))">
        {{ achName(achievement.achievement) }}
      </h1>

      <p class="ach-st__desc" data-test="share-desc">{{ localizedDesc }}</p>
    </section>

    <div v-if="unlockedDate" class="sc-stat ach-st__date">
      <span class="sc-label">{{ t("achievements.filters.unlocked") }}</span>
      <span class="sc-stat__value sc-stat__value--sm" data-test="share-date">{{ unlockedDate }}</span>
    </div>

    <ShareCardFooter :url="shortUrl" class="ach-st__footer" />
  </article>
</template>

<script setup lang="ts">
import { computed } from "vue"
import { useI18n } from "vue-i18n"
import { TIER_NUMERALS, categoryOf } from "@/components/achievement/config"
import { useAchievementText } from "@/components/achievement/text"
import type { UserAchievement } from "@/types/achievement"
import type { ShareTheme } from "@/types/share"
import { getAchievementIcon } from "@/utils/achievementIcons"
import { safeCardImageUrl, titleSizeClass } from "../shareCards"
import ShareCardFooter from "./ShareCardFooter.vue"
import "../shareCard.css"

const props = withDefaults(
  defineProps<{
    achievement: UserAchievement
    /** All tiers of the family in ascending order; single badges have length <= 1. */
    tiers?: UserAchievement[]
    username?: string
    /** Data or blob URL from the share image proxy; raw remote URLs are rejected. */
    avatarUrl?: string | null
    theme?: ShareTheme
    shortUrl?: string
  }>(),
  {
    tiers: () => [],
    username: "",
    avatarUrl: null,
    theme: "dark",
    shortUrl: "",
  },
)

const { t, locale } = useI18n()
const { description, name: achName } = useAchievementText()

const icon = computed(() => getAchievementIcon(props.achievement.achievement.action_type))
const category = computed(() => categoryOf(props.achievement.achievement.action_type))
const localizedDesc = computed(() => description(props.achievement.achievement))
const avatar = computed(() => safeCardImageUrl(props.avatarUrl))

const initials = computed(() => {
  const clean = props.username.trim()
  return clean ? Array.from(clean).slice(0, 2).join("").toUpperCase() : "?"
})

const tierLabel = computed(() => {
  if (props.tiers.length <= 1) return null
  const idx = props.tiers.findIndex((item) => item.achievement_id === props.achievement.achievement_id)
  if (idx < 0) return null
  return t("achievements.family.tier", { tier: TIER_NUMERALS[idx] ?? idx + 1 })
})

const unlockedDate = computed(() => {
  const iso = props.achievement.unlocked_at
  if (!iso) return null
  const d = new Date(iso)
  if (Number.isNaN(d.getTime())) return null
  return new Intl.DateTimeFormat(locale.value, { dateStyle: "long" }).format(d)
})
</script>

<style scoped>
.ach-st { gap: 48px; padding: 80px 72px 64px; }
.ach-st__head { display: flex; flex-direction: column; align-items: flex-start; gap: 28px; }
.ach-st__user { display: inline-flex; align-items: center; gap: 24px; min-width: 0; }
.ach-st__avatar,
.ach-st__initials {
  flex: none;
  width: 84px;
  height: 84px;
  border-radius: 999px;
  border: 3px solid var(--sc-line);
}
.ach-st__avatar { object-fit: cover; background: var(--sc-panel); }
.ach-st__initials {
  display: grid;
  place-items: center;
  font-size: 32px;
  font-weight: 800;
  color: var(--sc-fg);
  background: var(--sc-panel);
}
.ach-st__handle { font-size: 42px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.ach-st__hero {
  display: flex;
  flex: 1;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 44px;
  text-align: center;
}
.ach-st__halo {
  display: grid;
  place-items: center;
  width: 420px;
  height: 420px;
  background: var(--sc-panel);
  border: 3px solid var(--sc-line);
  border-radius: 999px;
}
.ach-st__emblem {
  display: grid;
  place-items: center;
  width: 336px;
  height: 336px;
  color: var(--color-accent);
  background: var(--color-brand);
  border: 6px solid var(--sc-line);
  border-radius: 999px;
}
.sc-theme-brand .ach-st__emblem { background: var(--sc-panel); }
.ach-st__tags { display: flex; flex-wrap: wrap; justify-content: center; gap: 20px; }
.ach-st__tag {
  padding: 14px 36px;
  font-size: 34px;
  font-weight: 700;
  color: var(--sc-fg);
  background: var(--sc-panel);
  border: 3px solid var(--sc-line);
  border-radius: 999px;
}
.ach-st__tag--tier {
  color: var(--sc-bg);
  background: var(--sc-fg);
  border-color: var(--sc-fg);
}
.ach-st__tag--xp { color: var(--color-accent); }
.sc-theme-light .ach-st__tag--xp { color: var(--color-brand); }
.ach-st__title { font-size: 112px; line-height: 1.06; }
.ach-st__title.sc-title--sm,
.ach-st__title.sc-title--xs { font-size: 84px; }
.ach-st__desc {
  max-width: 860px;
  margin: 0;
  font-size: 46px;
  font-weight: 600;
  line-height: 1.42;
  color: var(--sc-muted);
  display: -webkit-box;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 4;
  line-clamp: 4;
  overflow: hidden;
}
.ach-st__date { align-items: center; text-align: center; padding: 36px 48px; }
.ach-st__footer { margin-top: auto; }
</style>
