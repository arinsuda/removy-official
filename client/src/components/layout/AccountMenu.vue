<template>
  <div ref="root" class="am">
    <button
      :id="triggerId"
      ref="trigger"
      type="button"
      class="am-trigger"
      :class="{ 'am-trigger--open': open }"
      aria-haspopup="menu"
      :aria-expanded="open"
      :aria-controls="menuId"
      :aria-label="t('navigation.header.accountMenu')"
      @click="toggle"
      @keydown="onTriggerKeydown"
    >
      <AppAvatar :src="user.avatar_url" :name="displayName" size="sm" decorative />
      <span class="am-trigger-name">{{ user.username }}</span>
      <ChevronDown :size="14" class="am-chevron" aria-hidden="true" />
    </button>

    <Transition name="am-pop">
      <div v-if="open" :id="menuId" class="am-panel" role="menu" :aria-labelledby="triggerId" @keydown="onMenuKeydown">
        <div class="am-card" role="none">
          <AppAvatar :src="user.avatar_url" :name="displayName" size="md" decorative />
          <div class="am-card-text">
            <p class="am-card-name">{{ displayName }}</p>
            <p class="am-card-user">@{{ user.username }}</p>
          </div>
          <span v-if="user.level > 0" class="am-level">{{ t("navigation.header.level", { level: user.level }) }}</span>
        </div>

        <p :id="`${menuId}-you`" class="am-label" role="none">{{ t("navigation.header.yourSpace") }}</p>
        <div role="group" :aria-labelledby="`${menuId}-you`">
          <RouterLink
            v-for="item in spaceItems"
            :key="item.key"
            :to="item.to"
            role="menuitem"
            tabindex="-1"
            class="am-item"
            @click="close"
          >
            <span class="am-tile"><component :is="item.icon" :size="16" aria-hidden="true" /></span>
            <span class="am-text">
              <span class="am-title">{{ t(item.title) }}</span>
              <span class="am-sub">{{ t(item.hint) }}</span>
            </span>
          </RouterLink>
        </div>

        <p :id="`${menuId}-community`" class="am-label" role="none">{{ t("navigation.header.community") }}</p>
        <div role="group" :aria-labelledby="`${menuId}-community`">
          <RouterLink :to="{ name: 'feed' }" role="menuitem" tabindex="-1" class="am-item" @click="close">
            <span class="am-tile"><Rss :size="16" aria-hidden="true" /></span>
            <span class="am-text">
              <span class="am-title">{{ t("navigation.feed") }}</span>
              <span class="am-sub">{{ t("navigation.header.feedHint") }}</span>
            </span>
          </RouterLink>
          <button v-if="isAdmin" type="button" role="menuitem" tabindex="-1" class="am-item" @click="emitAndClose('enterAdmin')">
            <span class="am-tile am-tile--admin"><Shield :size="16" aria-hidden="true" /></span>
            <span class="am-text">
              <span class="am-title">{{ t("navigation.adminMode") }}</span>
              <span class="am-sub">{{ t("navigation.header.adminHint") }}</span>
            </span>
          </button>
        </div>

        <div class="am-sep" role="separator" />

        <button type="button" role="menuitem" tabindex="-1" class="am-item am-item--platform" @click="emitAndClose('switchPlanzi')">
          <img src="/planzi-icon-192.png" alt="" class="am-platform-icon" width="32" height="32" />
          <span class="am-text">
            <span class="am-title">{{ t("navigation.switchToPlanzi") }}</span>
            <span class="am-sub">{{ t("navigation.header.planziHint") }}</span>
          </span>
          <ExternalLink :size="14" class="am-external" aria-hidden="true" />
        </button>

        <button type="button" role="menuitem" tabindex="-1" class="am-item am-item--danger" @click="emitAndClose('logout')">
          <span class="am-tile"><LogOut :size="16" aria-hidden="true" /></span>
          <span class="am-text"><span class="am-title">{{ t("navigation.logout") }}</span></span>
        </button>
      </div>
    </Transition>
  </div>
</template>

<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, ref, useId } from "vue"
import { useI18n } from "vue-i18n"
import { BookMarked, ChevronDown, ExternalLink, LogOut, Rss, Settings, Shield, Trophy, User as UserIcon } from "lucide-vue-next"
import AppAvatar from "@/components/ui/AppAvatar.vue"
import type { AuthUser } from "@/types/auth"

const props = defineProps<{ user: AuthUser; isAdmin: boolean }>()
const emit = defineEmits<{ switchPlanzi: []; logout: []; enterAdmin: [] }>()
const open = defineModel<boolean>("open", { default: false })

const { t } = useI18n()
const uid = useId()
const triggerId = `am-trigger-${uid}`
const menuId = `am-menu-${uid}`
const root = ref<HTMLElement | null>(null)
const trigger = ref<HTMLButtonElement | null>(null)

const displayName = computed(() => props.user.display_name?.trim() || props.user.username)
const spaceItems = computed(() => [
  { key: "profile", to: `/users/${props.user.id}`, icon: UserIcon, title: "navigation.profile", hint: "navigation.header.profileHint" },
  { key: "library", to: `/users/${props.user.id}/library`, icon: BookMarked, title: "navigation.myLibrary", hint: "navigation.header.libraryHint" },
  { key: "achievements", to: `/users/${props.user.id}/achievements`, icon: Trophy, title: "navigation.achievement", hint: "navigation.header.achievementHint" },
  { key: "settings", to: "/settings", icon: Settings, title: "navigation.settings", hint: "navigation.header.settingsHint" },
])

const items = () => Array.from(root.value?.querySelectorAll<HTMLElement>('[role="menuitem"]') ?? [])

async function openAndFocus(which: "first" | "last") {
  open.value = true
  await nextTick()
  const list = items()
  ;(which === "first" ? list[0] : list[list.length - 1])?.focus()
}

function toggle() {
  if (open.value) open.value = false
  else openAndFocus("first")
}
function close() {
  open.value = false
}
function emitAndClose(event: "switchPlanzi" | "logout" | "enterAdmin") {
  open.value = false
  if (event === "switchPlanzi") emit("switchPlanzi")
  else if (event === "logout") emit("logout")
  else emit("enterAdmin")
}

function onTriggerKeydown(e: KeyboardEvent) {
  if (open.value) return
  if (e.key === "ArrowDown") {
    e.preventDefault()
    openAndFocus("first")
  } else if (e.key === "ArrowUp") {
    e.preventDefault()
    openAndFocus("last")
  }
}

function onMenuKeydown(e: KeyboardEvent) {
  const list = items()
  const index = list.indexOf(document.activeElement as HTMLElement)
  const focusAt = (i: number) => {
    e.preventDefault()
    list[(i + list.length) % list.length]?.focus()
  }
  switch (e.key) {
    case "ArrowDown":
      return focusAt(index + 1)
    case "ArrowUp":
      return focusAt(index < 0 ? list.length - 1 : index - 1)
    case "Home":
      return focusAt(0)
    case "End":
      return focusAt(list.length - 1)
    case "Escape":
      e.preventDefault()
      e.stopPropagation()
      open.value = false
      trigger.value?.focus()
      return
    case "Tab":
      open.value = false
  }
}

function onDocumentClick(e: MouseEvent) {
  if (open.value && root.value && !root.value.contains(e.target as Node)) open.value = false
}
function onDocumentKeydown(e: KeyboardEvent) {
  // Escape with focus outside the menu (e.g. after a mouse open) still closes it.
  if (open.value && e.key === "Escape") {
    open.value = false
    trigger.value?.focus()
  }
}

onMounted(() => {
  document.addEventListener("click", onDocumentClick)
  document.addEventListener("keydown", onDocumentKeydown)
})
onBeforeUnmount(() => {
  document.removeEventListener("click", onDocumentClick)
  document.removeEventListener("keydown", onDocumentKeydown)
})
</script>

<style scoped>
.am {
  position: relative;
  flex-shrink: 0;
}
.am-trigger {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  height: 2.5rem;
  padding: 0 0.6rem 0 0.25rem;
  border: 1px solid var(--color-border);
  border-radius: 9999px;
  background: color-mix(in srgb, var(--color-text-primary) 6%, transparent);
  color: var(--color-text-primary);
  cursor: pointer;
  transition: background var(--motion-fast), border-color var(--motion-fast);
}
.am-trigger :deep(.app-avatar) {
  box-shadow: none;
}
@media (hover: hover) {
  .am-trigger:hover {
    background: color-mix(in srgb, var(--color-text-primary) 12%, transparent);
    border-color: var(--color-border-strong);
  }
}
.am-trigger--open {
  background: color-mix(in srgb, var(--color-text-primary) 12%, transparent);
  border-color: var(--color-border-strong);
}
.am-trigger-name {
  max-width: 7.5rem;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-size: 0.875rem;
  font-weight: 500;
}
.am-chevron {
  color: var(--color-text-muted);
  transition: transform var(--motion-fast) var(--ease-out-soft);
}
.am-trigger--open .am-chevron {
  transform: rotate(180deg);
}
@media (max-width: 1023.98px) {
  .am-trigger {
    padding-right: 0.25rem;
  }
  .am-trigger-name,
  .am-chevron {
    display: none;
  }
}

.am-panel {
  position: absolute;
  top: calc(100% + 0.5rem);
  right: 0;
  width: 18.5rem;
  max-width: calc(100vw - 1rem);
  max-height: calc(100vh - 5rem); max-height: calc(100dvh - 5rem);
  overflow-y: auto;
  padding: 0.5rem;
  border: 1px solid var(--color-border-strong);
  border-radius: 1rem;
  background: var(--color-surface-2);
  box-shadow: var(--shadow-popover);
  transform-origin: top right;
}
/* Phones: full-width panel under the header, like the notification dropdown. */
@media (max-width: 639.98px) {
  .am-panel {
    position: fixed;
    top: 4rem;
    right: max(0.75rem, env(safe-area-inset-right));
    left: max(0.75rem, env(safe-area-inset-left));
    width: auto;
    max-width: none;
    max-height: calc(100vh - 5rem - env(safe-area-inset-bottom));
    max-height: calc(100dvh - 5rem - env(safe-area-inset-bottom));
  }
}
.am-card {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 0.75rem;
  margin-bottom: 0.25rem;
  border-radius: 0.75rem;
  background: var(--color-surface-3);
}
.am-card-text {
  min-width: 0;
  flex: 1;
}
.am-card-name {
  margin: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-size: 0.9375rem;
  font-weight: 600;
  color: var(--color-text-primary);
}
.am-card-user {
  margin: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-size: 0.8125rem;
  color: var(--color-text-muted);
}
.am-level {
  flex-shrink: 0;
  padding: 0.15rem 0.5rem;
  border-radius: 9999px;
  background: color-mix(in srgb, var(--color-brand) 18%, transparent);
  color: var(--color-text-primary);
  font-size: 0.75rem;
  font-weight: 600;
}
.am-label {
  margin: 0.5rem 0.75rem 0.25rem;
  font-size: 0.6875rem;
  font-weight: 600;
  letter-spacing: 0.06em;
  text-transform: uppercase;
  color: var(--color-text-muted);
}
.am-item {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  width: 100%;
  padding: 0.5rem 0.75rem;
  border: none;
  border-radius: 0.625rem;
  background: none;
  color: var(--color-text-primary);
  text-align: left;
  text-decoration: none;
  cursor: pointer;
  transition: background var(--motion-fast);
}
.am-item:hover,
.am-item:focus-visible {
  background: color-mix(in srgb, var(--color-text-primary) 8%, transparent);
}
.am-item:focus-visible {
  outline-offset: -2px;
}
.am-tile {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 2rem;
  height: 2rem;
  flex-shrink: 0;
  border-radius: 0.5rem;
  background: var(--color-surface-3);
  color: var(--color-text-secondary);
}
.am-tile--admin {
  color: var(--color-warning);
}
.am-text {
  display: flex;
  min-width: 0;
  flex: 1;
  flex-direction: column;
}
.am-title {
  font-size: 0.875rem;
  font-weight: 500;
}
.am-sub {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-size: 0.75rem;
  color: var(--color-text-muted);
}
.am-sep {
  height: 1px;
  margin: 0.5rem 0.25rem;
  background: var(--color-border);
}
.am-item--platform {
  border: 1px solid var(--color-border);
  background: color-mix(in srgb, var(--color-info) 8%, transparent);
  margin-bottom: 0.25rem;
}
.am-platform-icon {
  flex-shrink: 0;
  border-radius: 0.5rem;
  object-fit: contain;
}
.am-external {
  flex-shrink: 0;
  color: var(--color-text-muted);
}
.am-item--danger {
  color: var(--color-danger);
}
.am-item--danger .am-tile {
  color: var(--color-danger);
  background: color-mix(in srgb, var(--color-danger) 12%, transparent);
}
.am-item--danger:hover,
.am-item--danger:focus-visible {
  background: color-mix(in srgb, var(--color-danger) 12%, transparent);
}

.am-pop-enter-active,
.am-pop-leave-active {
  transition: opacity var(--motion-fast) var(--ease-out-soft), transform var(--motion-fast) var(--ease-out-soft);
}
.am-pop-enter-from,
.am-pop-leave-to {
  opacity: 0;
  transform: scale(0.95) translateY(-4px);
}
@media (prefers-reduced-motion: reduce) {
  .am-pop-enter-active,
  .am-pop-leave-active,
  .am-chevron {
    transition: none;
  }
}
</style>
