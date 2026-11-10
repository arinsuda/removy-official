<template>
  <article class="sc-card sc-card--square ach-sq" :class="`sc-theme-${theme}`" data-template="achievement-square">
    <header class="ach-sq__head">
      <span class="sc-chip">{{ t("achievements.share.badge") }}</span>
      <div v-if="username" class="ach-sq__user" data-test="share-user">
        <img v-if="avatar" :src="avatar" :alt="username" class="ach-sq__avatar" />
        <span v-else class="ach-sq__initials" aria-hidden="true">{{ initials }}</span>
        <span class="sc-meta ach-sq__handle">@{{ username }}</span>
      </div>
    </header>

    <section class="ach-sq__main">
      <div class="ach-sq__emblem">
        <component :is="icon" :size="108" :stroke-width="1.75" aria-hidden="true" />
      </div>

      <div class="ach-sq__body">
        <div class="ach-sq__tags">
          <span class="ach-sq__tag">{{ t(`achievements.categories.${category}`) }}</span>
          <span v-if="tierLabel" class="ach-sq__tag ach-sq__tag--tier" data-test="share-tier">{{ tierLabel }}</span>
          <span v-if="achievement.achievement.exp_reward" class="ach-sq__tag ach-sq__tag--xp">
            {{ t("achievements.xp", { n: achievement.achievement.exp_reward }) }}
          </span>
        </div>

        <h1 class="sc-title ach-sq__title" :class="titleSizeClass(achievement.achievement.name)">
          {{ achievement.achievement.name }}
        </h1>
        <p class="ach-sq__desc" data-test="share-desc">{{ localizedDesc }}</p>
      </div>
    </section>

    <div v-if="unlockedDate" class="sc-stat ach-sq__date">
      <span class="sc-label">{{ t("achievements.filters.unlocked") }}</span>
      <span class="ach-sq__date-val" data-test="share-date">{{ unlockedDate }}</span>
    </div>

    <ShareCardFooter :url="shortUrl" class="ach-sq__footer" />
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
const { description } = useAchievementText()

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
.ach-sq { gap: 40px; padding: 64px 72px 56px; }
.ach-sq__head { display: flex; align-items: center; justify-content: space-between; gap: 24px; }
.ach-sq__user { display: inline-flex; align-items: center; gap: 20px; min-width: 0; }
.ach-sq__avatar,
.ach-sq__initials {
  flex: none;
  width: 72px;
  height: 72px;
  border-radius: 999px;
  border: 3px solid var(--sc-line);
}
.ach-sq__avatar { object-fit: cover; background: var(--sc-panel); }
.ach-sq__initials {
  display: grid;
  place-items: center;
  font-size: 28px;
  font-weight: 800;
  color: var(--sc-fg);
  background: var(--sc-panel);
}
.ach-sq__handle { font-size: 36px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.ach-sq__main {
  display: flex;
  flex: 1;
  align-items: center;
  gap: 56px;
  min-height: 0;
}
.ach-sq__emblem {
  flex: none;
  display: grid;
  place-items: center;
  width: 240px;
  height: 240px;
  color: var(--color-accent);
  background: var(--color-brand);
  border: 6px solid var(--sc-line);
  border-radius: 999px;
}
.sc-theme-brand .ach-sq__emblem { background: var(--sc-panel); }
.ach-sq__body { display: flex; flex: 1; flex-direction: column; gap: 24px; min-width: 0; }
.ach-sq__tags { display: flex; flex-wrap: wrap; gap: 16px; }
.ach-sq__tag {
  padding: 10px 28px;
  font-size: 28px;
  font-weight: 700;
  color: var(--sc-fg);
  background: var(--sc-panel);
  border: 2px solid var(--sc-line);
  border-radius: 999px;
}
.ach-sq__tag--tier {
  color: var(--sc-bg);
  background: var(--sc-fg);
  border-color: var(--sc-fg);
}
.ach-sq__tag--xp { color: var(--color-accent); }
.sc-theme-light .ach-sq__tag--xp { color: var(--color-brand); }
.ach-sq__title { font-size: 84px; line-height: 1.08; }
.ach-sq__title.sc-title--sm,
.ach-sq__title.sc-title--xs { font-size: 64px; }
.ach-sq__desc {
  margin: 0;
  font-size: 38px;
  font-weight: 600;
  line-height: 1.38;
  color: var(--sc-muted);
  display: -webkit-box;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 3;
  line-clamp: 3;
  overflow: hidden;
}
.ach-sq__date {
  flex-direction: row;
  align-items: center;
  justify-content: space-between;
  padding: 28px 40px;
}
.ach-sq__date-val { font-size: 38px; font-weight: 800; color: var(--sc-fg); }
.ach-sq__footer { margin-top: auto; }
</style>
