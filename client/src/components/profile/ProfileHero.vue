<template>
  <header class="hero">
    <AppAvatar :src="user.avatar_url" :name="displayName" size="xl" />
    <div class="hero__body">
      <div class="hero__title">
        <h1 class="hero__name">{{ displayName }}</h1>
        <span class="hero__chip">{{ t("profile.hero.level", { level: user.level ?? 1 }) }}</span>
        <span v-if="user.is_private" class="hero__chip hero__chip--lock">
          <Lock :size="12" aria-hidden="true" />{{ t("profile.hero.private") }}
        </span>
      </div>
      <p class="hero__handle">@{{ user.username }}</p>
      <p v-if="user.bio" class="hero__bio">{{ user.bio }}</p>
      <p class="hero__joined">{{ t("profile.hero.joined", { date: joined }) }}</p>
      <div class="hero__counts">
        <div class="hero__count">
          <span class="hero__num">{{ user.review_count ?? 0 }}</span>
          <span class="hero__lbl">{{ t("profile.hero.reviews") }}</span>
        </div>
        <button type="button" class="hero__count hero__count--btn" @click="emit('openList', 'followers')">
          <span class="hero__num">{{ user.follower_count ?? 0 }}</span>
          <span class="hero__lbl">{{ t("profile.hero.followers") }}</span>
        </button>
        <button type="button" class="hero__count hero__count--btn" @click="emit('openList', 'following')">
          <span class="hero__num">{{ user.following_count ?? 0 }}</span>
          <span class="hero__lbl">{{ t("profile.hero.following") }}</span>
        </button>
      </div>
    </div>
    <div class="hero__actions">
      <AppButton v-if="isOwner" variant="outline" to="/settings">
        <Settings :size="14" aria-hidden="true" />{{ t("profile.hero.editProfile") }}
      </AppButton>
      <template v-else-if="canFollow">
        <AppButton v-if="followState === 'accepted'" variant="outline" :loading="busy" @click="emit('unfollow')">
          <UserCheck :size="14" aria-hidden="true" />{{ t("profile.follow.following") }}
        </AppButton>
        <AppButton v-else-if="followState === 'pending'" variant="outline" :loading="busy" @click="emit('cancel')">
          <Clock :size="14" aria-hidden="true" />{{ t("profile.follow.requested") }}
          <span class="sr-only">{{ t("profile.follow.cancelRequest") }}</span>
        </AppButton>
        <AppButton v-else variant="primary" :loading="busy" @click="emit('follow')">
          <UserPlus :size="14" aria-hidden="true" />{{ t("profile.follow.follow") }}
        </AppButton>
      </template>
    </div>
  </header>
</template>

<script setup lang="ts">
import { computed } from "vue"
import { useI18n } from "vue-i18n"
import { Clock, Lock, Settings, UserCheck, UserPlus } from "lucide-vue-next"
import AppAvatar from "@/components/ui/AppAvatar.vue"
import AppButton from "@/components/ui/AppButton.vue"
import type { UserProfile } from "@/types/user"

const props = defineProps<{
  user: UserProfile
  isOwner: boolean
  canFollow: boolean
  followState?: "pending" | "accepted"
  busy?: boolean
}>()
const emit = defineEmits<{
  follow: []
  unfollow: []
  cancel: []
  openList: [kind: "followers" | "following"]
}>()

const { t, locale } = useI18n()
const displayName = computed(() => props.user.display_name || props.user.username)
const joined = computed(() => new Date(props.user.created_at).toLocaleDateString(locale.value, { year: "numeric", month: "short" }))
</script>

<style scoped>
.hero { display: flex; flex-direction: column; align-items: flex-start; gap: 1rem; padding: 1.5rem 0; }
.hero__body { min-width: 0; flex: 1; }
.hero__title { display: flex; flex-wrap: wrap; align-items: center; gap: 0.5rem; }
.hero__name { font-size: 1.5rem; font-weight: 700; color: var(--color-text-primary); overflow-wrap: anywhere; }
.hero__chip { display: inline-flex; align-items: center; gap: 0.25rem; padding: 0.125rem 0.5rem; font-size: 0.75rem; font-weight: 600; color: var(--color-text-secondary); background: var(--color-surface-3); border-radius: 999px; }
.hero__handle, .hero__joined { font-size: 0.8125rem; color: var(--color-text-muted); }
.hero__bio { margin: 0.5rem 0; max-width: 40rem; color: var(--color-text-secondary); white-space: pre-line; overflow-wrap: anywhere; }
.hero__counts { display: flex; gap: 1.5rem; margin-top: 0.75rem; }
.hero__count { display: flex; flex-direction: column; align-items: flex-start; padding: 0; color: inherit; text-align: left; background: none; border: 0; }
.hero__num { font-size: 1.125rem; font-weight: 700; color: var(--color-text-primary); }
.hero__lbl { font-size: 0.75rem; color: var(--color-text-muted); }
.hero__count--btn { cursor: pointer; }
.hero__count--btn:hover .hero__lbl { color: var(--color-text-primary); }
.sr-only { position: absolute; width: 1px; height: 1px; overflow: hidden; clip-path: inset(50%); white-space: nowrap; }
@media (min-width: 768px) {
  .hero { flex-direction: row; align-items: flex-start; gap: 1.5rem; padding: 2.5rem 0 1.5rem; }
}
</style>
