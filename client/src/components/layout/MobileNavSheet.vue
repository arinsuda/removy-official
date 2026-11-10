<template>
  <Teleport to="body">
    <Transition name="mns">
      <div v-if="modelValue" class="mns" @mousedown.self="close">
        <div
          ref="panel"
          class="mns-panel"
          role="dialog"
          aria-modal="true"
          :aria-label="t('navigation.header.mainNav')"
          tabindex="-1"
          @keydown="onKeydown"
        >
          <div class="mns-top">
            <RouterLink to="/" class="mns-logo" @click="close">
              <img src="/popcorn-icon-192.png" alt="" width="28" height="28" /><span>POP</span><span class="mns-logo-brand">CORN</span> <!-- hardcode-ok: brand name -->
            </RouterLink>
            <button type="button" class="mns-close" :aria-label="t('navigation.header.closeMenu')" @click="close">
              <X :size="20" aria-hidden="true" />
            </button>
          </div>

          <div class="mns-body">
            <RouterLink v-if="user" :to="`/users/${user.id}`" class="mns-card" @click="close">
              <AppAvatar :src="user.avatar_url" :name="displayName" size="md" decorative />
              <span class="mns-card-text">
                <span class="mns-card-name">{{ displayName }}</span>
                <span class="mns-card-user">@{{ user.username }}</span>
              </span>
              <span v-if="user.level > 0" class="mns-level">{{ t("navigation.header.level", { level: user.level }) }}</span>
            </RouterLink>

            <nav :aria-label="t('navigation.header.mainNav')" class="mns-list">
              <RouterLink
                v-for="item in NAV_ITEMS"
                :key="item.key"
                :to="item.to"
                class="mns-row"
                :class="{ 'mns-row--active': item.isActive(route) }"
                :aria-current="item.isActive(route) ? 'page' : undefined"
                @click="close"
              >
                <component :is="item.icon" :size="20" aria-hidden="true" />{{ t(item.labelKey) }}
              </RouterLink>
            </nav>

            <template v-if="user">
              <p class="mns-label">{{ t("navigation.header.account") }}</p>
              <div class="mns-list">
                <RouterLink :to="`/users/${user.id}/library`" class="mns-row" @click="close">
                  <BookMarked :size="20" aria-hidden="true" />{{ t("navigation.myLibrary") }}
                </RouterLink>
                <RouterLink :to="`/users/${user.id}/achievements`" class="mns-row" @click="close">
                  <Trophy :size="20" aria-hidden="true" />{{ t("navigation.achievement") }}
                </RouterLink>
                <RouterLink :to="{ name: 'feed' }" class="mns-row" @click="close">
                  <Rss :size="20" aria-hidden="true" />{{ t("navigation.feed") }}
                </RouterLink>
                <RouterLink to="/settings" class="mns-row" @click="close">
                  <Settings :size="20" aria-hidden="true" />{{ t("navigation.settings") }}
                </RouterLink>
                <button v-if="isAdmin" type="button" class="mns-row" @click="emitAndClose('enterAdmin')">
                  <Shield :size="20" aria-hidden="true" />{{ t("navigation.adminMode") }}
                </button>
              </div>
            </template>

            <p class="mns-label">{{ t("navigation.header.language") }}</p>
            <LanguageToggle variant="segmented" />

            <div v-if="user" class="mns-list mns-list--end">
              <button type="button" class="mns-row mns-row--platform" @click="emitAndClose('switchPlanzi')">
                <img src="/planzi-icon-192.png" alt="" width="24" height="24" class="mns-platform-icon" />
                <span class="mns-row-text">
                  {{ t("navigation.switchToPlanzi") }}
                  <small>{{ t("navigation.header.planziHint") }}</small>
                </span>
                <ExternalLink :size="16" aria-hidden="true" />
              </button>
              <button type="button" class="mns-row mns-row--danger" @click="emitAndClose('logout')">
                <LogOut :size="20" aria-hidden="true" />{{ t("navigation.logout") }}
              </button>
            </div>
            <div v-else class="mns-guest">
              <AppButton to="/login" variant="outline" block size="lg" @click="close">{{ t("navigation.logIn") }}</AppButton>
              <AppButton to="/register" block size="lg" @click="close">{{ t("navigation.signUp") }}</AppButton>
            </div>
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, ref, watch } from "vue"
import { useRoute } from "vue-router"
import { useI18n } from "vue-i18n"
import { BookMarked, ExternalLink, LogOut, Rss, Settings, Shield, Trophy, X } from "lucide-vue-next"
import AppAvatar from "@/components/ui/AppAvatar.vue"
import AppButton from "@/components/ui/AppButton.vue"
import LanguageToggle from "./LanguageToggle.vue"
import { NAV_ITEMS } from "./navItems"
import { lockScroll, unlockScroll } from "@/composables/scrollLock"
import { useDialogFocus } from "@/composables/useDialogFocus"
import type { AuthUser } from "@/types/auth"

const props = defineProps<{ modelValue: boolean; user: AuthUser | null; isAdmin: boolean }>()
const emit = defineEmits<{ "update:modelValue": [boolean]; switchPlanzi: []; logout: []; enterAdmin: [] }>()

const { t } = useI18n()
const route = useRoute()
const panel = ref<HTMLElement | null>(null)
const active = computed(() => props.modelValue)
const { onKeydown: onKeydownTrap } = useDialogFocus(panel, active, ref('.mns-close'))

const displayName = computed(() => props.user?.display_name?.trim() || props.user?.username || "")

function close() {
  emit("update:modelValue", false)
}
function emitAndClose(event: "switchPlanzi" | "logout" | "enterAdmin") {
  close()
  if (event === "switchPlanzi") emit("switchPlanzi")
  else if (event === "logout") emit("logout")
  else emit("enterAdmin")
}
function onKeydown(e: KeyboardEvent) {
  if (e.key === "Escape") {
    e.stopPropagation()
    close()
    return
  }
  onKeydownTrap(e)
}

let locked = false
watch(
  active,
  (open) => {
    if (open && !locked) {
      lockScroll()
      locked = true
    } else if (!open && locked) {
      unlockScroll()
      locked = false
    }
  },
  { immediate: true },
)
watch(() => route.fullPath, close)
onBeforeUnmount(() => {
  if (locked) unlockScroll()
})
</script>

<style scoped>
.mns {
  position: fixed;
  inset: 0;
  z-index: var(--z-modal-base);
  display: flex;
  justify-content: flex-end;
  background: var(--color-scrim);
  backdrop-filter: blur(2px);
}
.mns-panel {
  display: flex;
  flex-direction: column;
  width: min(22rem, 100%);
  height: 100%;
  max-height: 100dvh;
  padding-top: env(safe-area-inset-top);
  padding-right: env(safe-area-inset-right);
  padding-bottom: env(safe-area-inset-bottom);
  background: var(--color-surface-2);
  border-left: 1px solid var(--color-border-strong);
  box-shadow: var(--shadow-sheet);
  outline: none;
}
.mns-top {
  display: flex;
  align-items: center;
  justify-content: space-between;
  height: 3.5rem;
  padding: 0 0.75rem 0 1rem;
  flex-shrink: 0;
  border-bottom: 1px solid var(--color-border);
}
.mns-logo {
  display: flex;
  align-items: center;
  gap: 0.375rem;
  font-size: 1.2rem;
  font-weight: 900;
  letter-spacing: -0.5px;
  color: var(--color-text-primary);
  text-decoration: none;
}
.mns-logo-brand {
  margin-left: -0.375rem;
  color: var(--color-brand);
}
.mns-close {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 2.75rem;
  height: 2.75rem;
  border: none;
  border-radius: 9999px;
  background: none;
  color: var(--color-text-secondary);
  cursor: pointer;
}
.mns-close:active {
  background: var(--color-surface-3);
}
.mns-body {
  flex: 1;
  min-height: 0;
  overflow-y: auto;
  overscroll-behavior: contain;
  padding: 1rem;
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}
.mns-card {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 0.75rem;
  margin-bottom: 0.5rem;
  border-radius: 0.875rem;
  background: var(--color-surface-3);
  color: var(--color-text-primary);
  text-decoration: none;
}
.mns-card-text {
  display: flex;
  min-width: 0;
  flex: 1;
  flex-direction: column;
}
.mns-card-name {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-weight: 600;
}
.mns-card-user {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-size: 0.8125rem;
  color: var(--color-text-muted);
}
.mns-level {
  padding: 0.15rem 0.5rem;
  border-radius: 9999px;
  background: color-mix(in srgb, var(--color-brand) 18%, transparent);
  font-size: 0.75rem;
  font-weight: 600;
}
.mns-label {
  margin: 1rem 0.5rem 0.25rem;
  font-size: 0.6875rem;
  font-weight: 600;
  letter-spacing: 0.06em;
  text-transform: uppercase;
  color: var(--color-text-muted);
}
.mns-list {
  display: flex;
  flex-direction: column;
  gap: 0.125rem;
}
.mns-list--end {
  margin-top: 1rem;
  padding-top: 0.75rem;
  border-top: 1px solid var(--color-border);
}
.mns-row {
  display: flex;
  align-items: center;
  gap: 0.875rem;
  width: 100%;
  min-height: 2.75rem;
  padding: 0.5rem 0.75rem;
  border: none;
  border-radius: 0.75rem;
  background: none;
  color: var(--color-text-secondary);
  font-size: 0.9375rem;
  font-weight: 500;
  text-align: left;
  text-decoration: none;
  cursor: pointer;
  transition: background var(--motion-fast), color var(--motion-fast);
}
.mns-row:active {
  background: var(--color-surface-3);
}
.mns-row--active {
  color: var(--color-text-primary);
  background: color-mix(in srgb, var(--color-brand) 14%, transparent);
  box-shadow: inset 3px 0 0 var(--color-brand);
}
.mns-row--platform {
  color: var(--color-text-primary);
  border: 1px solid var(--color-border);
  background: color-mix(in srgb, var(--color-info) 8%, transparent);
}
.mns-platform-icon {
  border-radius: 0.375rem;
  object-fit: contain;
}
.mns-row-text {
  display: flex;
  flex: 1;
  flex-direction: column;
}
.mns-row-text small {
  font-size: 0.75rem;
  color: var(--color-text-muted);
}
.mns-row--danger {
  color: var(--color-danger);
}
.mns-guest {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
  margin-top: 1.25rem;
}

.mns-enter-active,
.mns-leave-active {
  transition: opacity var(--motion-base) var(--ease-out-soft);
}
.mns-enter-active .mns-panel,
.mns-leave-active .mns-panel {
  transition: transform var(--motion-base) var(--ease-out-soft);
}
.mns-enter-from,
.mns-leave-to {
  opacity: 0;
}
.mns-enter-from .mns-panel,
.mns-leave-to .mns-panel {
  transform: translateX(100%);
}
@media (prefers-reduced-motion: reduce) {
  .mns-enter-active,
  .mns-leave-active,
  .mns-enter-active .mns-panel,
  .mns-leave-active .mns-panel {
    transition: none;
  }
}
</style>
