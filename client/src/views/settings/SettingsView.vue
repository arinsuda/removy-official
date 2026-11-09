<template>
  <div class="settings">
    <div class="settings__inner">
      <header class="settings__head">
        <h1 class="settings__title">{{ t("settings.title") }}</h1>
        <RouterLink v-if="auth.user" :to="`/users/${auth.user.id}`" class="settings__back">
          <ArrowLeft :size="14" aria-hidden="true" />{{ t("settings.backToProfile") }}
        </RouterLink>
      </header>

      <SkeletonProfile v-if="profile.isPending.value" />
      <AppErrorState v-else-if="!user" :title="t('profile.states.errorTitle')" :description="t('profile.states.errorBody')" @retry="profile.refetch()" />

      <div v-else class="settings__layout">
        <nav class="settings__nav" :aria-label="t('settings.title')">
          <a v-for="s in sections" :key="s" :href="`#${s}`" class="settings__link" :class="{ 'is-active': active === s }" :aria-current="active === s ? 'location' : undefined" @click.prevent="go(s)">
            {{ t(`settings.sections.${s}`) }}
          </a>
        </nav>

        <div class="settings__content">
          <SettingsProfileSection :user="user" :form="form" :avatar-preview="avatarPreview" @avatar="setAvatar" />
          <SettingsPreferencesSection v-model="form.favorite_genres" />
          <SettingsPrivacySection v-model="form.is_private" />
          <SettingsActivitySection />

          <section id="security" class="settings__group" aria-labelledby="security-heading">
            <h2 id="security-heading" class="settings__group-title">{{ t("settings.sections.security") }}</h2>
            <SettingsEmailSection :user="user" />
            <SettingsPasswordSection :user="user" />
          </section>

          <SettingsConnectionsSection />
          <SettingsDangerSection :user="user" />

          <SettingsSaveBar :dirty="isDirty" :saving="saving" :error="error" @save="onSave" @discard="reset" />
        </div>
      </div>
    </div>

    <AppModal :model-value="leaveOpen" size="sm" :title="t('settings.saveBar.label')" @close="decideLeave(false)">
      <p class="settings__modal-text">{{ t("settings.unsavedConfirm") }}</p>
      <template #footer>
        <AppButton variant="outline" @click="decideLeave(false)">{{ t("common.cancel") }}</AppButton>
        <AppButton variant="danger" @click="decideLeave(true)">{{ t("settings.saveBar.discard") }}</AppButton>
      </template>
    </AppModal>
  </div>
</template>

<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from "vue"
import { onBeforeRouteLeave, useRoute } from "vue-router"
import { useI18n } from "vue-i18n"
import { useQuery } from "@tanstack/vue-query"
import { ArrowLeft } from "lucide-vue-next"
import { userApi } from "@/api/api"
import { useAuthStore } from "@/stores/auth"
import { useUIStore } from "@/stores/ui"
import { useProfileForm } from "@/components/settings/useProfileForm"
import SkeletonProfile from "@/components/ui/skeleton/SkeletonProfile.vue"
import AppButton from "@/components/ui/AppButton.vue"
import AppErrorState from "@/components/ui/AppErrorState.vue"
import AppModal from "@/components/ui/AppModal.vue"
import SettingsProfileSection from "@/components/settings/SettingsProfileSection.vue"
import SettingsPreferencesSection from "@/components/settings/SettingsPreferencesSection.vue"
import SettingsPrivacySection from "@/components/settings/SettingsPrivacySection.vue"
import SettingsEmailSection from "@/components/settings/SettingsEmailSection.vue"
import SettingsPasswordSection from "@/components/settings/SettingsPasswordSection.vue"
import SettingsConnectionsSection from "@/components/settings/SettingsConnectionsSection.vue"
import { RETURN_FLAG } from "@/components/settings/connectionsReturn"
import SettingsActivitySection from "@/components/settings/SettingsActivitySection.vue"
import SettingsDangerSection from "@/components/settings/SettingsDangerSection.vue"
import SettingsSaveBar from "@/components/settings/SettingsSaveBar.vue"

const SECTIONS = ["profile", "preferences", "privacy", "activity", "security", "connections", "danger"] as const
type SectionId = (typeof SECTIONS)[number]
const sections = SECTIONS

const { t } = useI18n()
const route = useRoute()
const auth = useAuthStore()
const ui = useUIStore()

const userId = computed(() => auth.user?.id ?? 0)
const profile = useQuery({
  queryKey: computed(() => ["user-profile", userId.value]),
  queryFn: () => userApi.getProfile(userId.value).then(res => res.data.user),
  enabled: computed(() => userId.value > 0),
  retry: false,
})
const user = computed(() => profile.data.value ?? null)

const { form, avatarPreview, saving, error, isDirty, reset, setAvatar, save } = useProfileForm(user)

async function onSave() {
  if (await save()) ui.toastSuccess(t("settings.saved"))
  else ui.toastError(t("settings.saveFailed"))
}

// ── section navigation ──
const active = ref<SectionId>("profile")
function go(id: SectionId) {
  active.value = id
  document.getElementById(id)?.scrollIntoView({ behavior: matchMedia("(prefers-reduced-motion: reduce)").matches ? "auto" : "smooth", block: "start" })
  history.replaceState(history.state, "", `#${id}`)
}
function scrollToHash() {
  const returned = sessionStorage.getItem(RETURN_FLAG)
  sessionStorage.removeItem(RETURN_FLAG)
  const id = (returned ? "connections" : route.hash.slice(1)) as SectionId
  if (SECTIONS.includes(id)) nextTick(() => go(id))
}
watch(user, (next, prev) => !prev && next && scrollToHash())
onMounted(() => user.value && scrollToHash())

// ── unsaved changes guard ──
const leaveOpen = ref(false)
let resolveLeave: ((allow: boolean) => void) | null = null
function decideLeave(allow: boolean) {
  leaveOpen.value = false
  resolveLeave?.(allow)
  resolveLeave = null
}
function onBeforeUnload(e: BeforeUnloadEvent) {
  if (!isDirty.value) return
  e.preventDefault()
  e.returnValue = ""
}
window.addEventListener("beforeunload", onBeforeUnload)
onBeforeUnmount(() => {
  window.removeEventListener("beforeunload", onBeforeUnload)
  resolveLeave?.(false)
})
onBeforeRouteLeave(() => {
  if (!isDirty.value || saving.value) return true
  resolveLeave?.(false)
  leaveOpen.value = true
  return new Promise<boolean>((resolve) => {
    resolveLeave = resolve
  })
})
</script>

<style scoped>
.settings { min-height: 100vh; color: var(--color-text-primary); background: var(--color-surface); }
.settings__inner { max-width: 75rem; margin: 0 auto; padding: 1.5rem 1rem 5rem; }
.settings__head { display: flex; flex-wrap: wrap; align-items: baseline; justify-content: space-between; gap: 0.5rem; margin-bottom: 1rem; }
.settings__title { font-size: 1.5rem; font-weight: 700; }
.settings__back { display: inline-flex; align-items: center; gap: 0.375rem; font-size: 0.8125rem; color: var(--color-text-secondary); }
.settings__back:hover { color: var(--color-text-primary); }
.settings__layout { display: block; }
.settings__nav { position: sticky; top: 3.5rem; z-index: var(--z-sticky); display: flex; gap: 0.5rem; padding: 0.5rem 0; margin-bottom: 1rem; overflow-x: auto; background: var(--color-surface); scrollbar-width: none; }
.settings__nav::-webkit-scrollbar { display: none; }
.settings__link { flex: none; padding: 0.375rem 0.875rem; font-size: 0.8125rem; color: var(--color-text-secondary); white-space: nowrap; border: 1px solid var(--color-border); border-radius: 999px; }
.settings__link.is-active { color: var(--color-surface); background: var(--color-text-primary); border-color: var(--color-text-primary); }
.settings__content { display: flex; flex-direction: column; gap: 1.25rem; min-width: 0; }
.settings__group { display: flex; flex-direction: column; gap: 1.25rem; scroll-margin-top: 6rem; }
.settings__group-title { font-size: 1.0625rem; font-weight: 700; }
.settings__modal-text { margin: 0; font-size: var(--text-small); line-height: 1.6; color: var(--color-text-secondary); }
@media (min-width: 768px) {
  .settings__inner { padding: 2rem 1.5rem 5rem; }
}
@media (min-width: 1024px) {
  .settings__layout { display: grid; grid-template-columns: 13rem minmax(0, 40rem); gap: 2rem; }
  .settings__nav { top: 5rem; flex-direction: column; align-self: start; padding: 0; margin: 0; overflow: visible; background: none; }
  .settings__link { padding: 0.5rem 0.75rem; border: 0; border-radius: var(--radius-btn); }
  .settings__link:hover { color: var(--color-text-primary); }
}
</style>
