<template>
  <div ref="root" class="feed-menu">
    <button
      type="button"
      class="feed-menu__trigger"
      :aria-label="t('feed.manage.open')"
      :title="t('feed.manage.current', { value: visibilityLabel(visibility) })"
      aria-haspopup="menu"
      :aria-expanded="open"
      @click="open = !open"
      @keydown.esc="open = false"
    >
      <component :is="ICONS[visibility]" :size="16" aria-hidden="true" />
    </button>

    <Transition name="feed-menu-pop">
      <div v-if="open" class="feed-menu__panel" role="menu" :aria-label="t('feed.manage.visibility')" @keydown.esc="open = false">
        <p class="feed-menu__heading">{{ t("feed.manage.visibility") }}</p>
        <button
          v-for="opt in VISIBILITY_OPTIONS"
          :key="opt"
          type="button"
          role="menuitemradio"
          :aria-checked="visibility === opt"
          class="feed-menu__item"
          :class="{ 'is-active': visibility === opt }"
          @click="choose(opt)"
        >
          <component :is="ICONS[opt]" :size="14" aria-hidden="true" />
          <span>{{ visibilityLabel(opt) }}</span>
          <Check v-if="visibility === opt" :size="14" class="feed-menu__check" aria-hidden="true" />
        </button>
        <hr class="feed-menu__rule" />
        <button type="button" role="menuitem" class="feed-menu__item feed-menu__item--danger" @click="askDelete">
          <Trash2 :size="14" aria-hidden="true" />
          <span>{{ t("feed.manage.delete") }}</span>
        </button>
      </div>
    </Transition>

    <AppModal v-model="confirming" role="alertdialog" size="sm" :show-close="false" :aria-labelledby="titleId" @close="confirming = false">
      <h3 :id="titleId" class="feed-menu__confirm-title">{{ t("feed.manage.deleteTitle") }}</h3>
      <p class="feed-menu__confirm-text">{{ t("feed.manage.deleteDescription") }}</p>
      <template #footer>
        <div class="feed-menu__confirm-actions">
          <AppButton variant="outline" @click="confirming = false">{{ t("common.cancel") }}</AppButton>
          <AppButton variant="danger" @click="confirmDelete">{{ t("common.delete") }}</AppButton>
        </div>
      </template>
    </AppModal>
  </div>
</template>

<script setup lang="ts">
import { ref, useId } from "vue"
import { useI18n } from "vue-i18n"
import { Check, Compass, Globe, Lock, Trash2, Users } from "lucide-vue-next"
import AppButton from "@/components/ui/AppButton.vue"
import AppModal from "@/components/ui/AppModal.vue"
import { useClickOutside } from "@/composables/useClickOutside"
import { VISIBILITY_OPTIONS, type Visibility } from "./feedConfig"

defineProps<{ visibility: Visibility }>()
const emit = defineEmits<{
  "update-visibility": [visibility: Visibility]
  "delete-activity": []
}>()

const { t } = useI18n()
const titleId = useId()
const open = ref(false)
const confirming = ref(false)
const root = ref<HTMLElement | null>(null)
useClickOutside(root, () => (open.value = false))

const ICONS = { default: Compass, public: Globe, followers: Users, private: Lock } as const

function visibilityLabel(v: Visibility): string {
  return v === "default" ? t("feed.manage.default") : t(`visibility.${v}`)
}

function choose(v: Visibility) {
  emit("update-visibility", v)
  open.value = false
}

function askDelete() {
  open.value = false
  confirming.value = true
}

function confirmDelete() {
  confirming.value = false
  emit("delete-activity")
}
</script>

<style scoped>
.feed-menu {
  position: relative;
  flex-shrink: 0;
}
.feed-menu__trigger {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 2.25rem;
  height: 2.25rem;
  border: 0;
  border-radius: var(--radius-btn);
  background: none;
  color: var(--color-text-muted);
  cursor: pointer;
  transition:
    color var(--motion-fast) ease,
    background-color var(--motion-fast) ease;
}
.feed-menu__trigger:hover {
  background: var(--color-surface-3);
  color: var(--color-text-primary);
}
.feed-menu__panel {
  position: absolute;
  top: 100%;
  right: 0;
  z-index: var(--z-popover);
  width: 14rem;
  margin-top: 0.25rem;
  padding: 0.375rem;
  border: 1px solid var(--color-border-strong);
  border-radius: var(--radius-card);
  background: var(--color-surface-2);
  box-shadow: var(--shadow-popover);
}
.feed-menu__heading {
  margin: 0;
  padding: 0.375rem 0.5rem;
  font-size: var(--text-caption);
  font-weight: 600;
  color: var(--color-text-muted);
}
.feed-menu__item {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  width: 100%;
  padding: 0.5rem;
  border: 0;
  border-radius: 0.5rem;
  background: none;
  font-size: var(--text-caption);
  text-align: left;
  color: var(--color-text-secondary);
  cursor: pointer;
}
.feed-menu__item:hover {
  background: var(--color-surface-3);
  color: var(--color-text-primary);
}
.feed-menu__item.is-active {
  color: var(--color-text-primary);
  font-weight: 600;
}
.feed-menu__check {
  margin-left: auto;
  color: var(--color-success);
}
.feed-menu__item--danger {
  color: var(--color-danger);
}
.feed-menu__rule {
  margin: 0.375rem 0;
  border: 0;
  border-top: 1px solid var(--color-border);
}
.feed-menu__confirm-title {
  margin: 0 0 0.5rem;
  font-size: var(--text-body);
  font-weight: 700;
  color: var(--color-text-primary);
}
.feed-menu__confirm-text {
  margin: 0;
  font-size: var(--text-small);
  color: var(--color-text-secondary);
}
.feed-menu__confirm-actions {
  display: flex;
  justify-content: flex-end;
  gap: 0.5rem;
}

.feed-menu-pop-enter-active,
.feed-menu-pop-leave-active {
  transition:
    opacity var(--motion-fast) ease,
    transform var(--motion-fast) var(--ease-out-soft);
}
.feed-menu-pop-enter-from,
.feed-menu-pop-leave-to {
  opacity: 0;
  transform: translateY(-4px);
}
@media (prefers-reduced-motion: reduce) {
  .feed-menu-pop-enter-from,
  .feed-menu-pop-leave-to {
    transform: none;
  }
}
</style>
