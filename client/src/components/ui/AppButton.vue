<template>
  <component
    :is="tag"
    :to="isRouterLink ? to : undefined"
    v-bind="linkHref ? { href: linkHref } : {}"
    :role="!isNativeButton && isInert ? 'link' : undefined"
    :type="isNativeButton ? type : undefined"
    :disabled="isNativeButton ? isInert : undefined"
    :aria-disabled="!isNativeButton && isInert ? 'true' : undefined"
    :tabindex="!isNativeButton && isInert ? -1 : undefined"
    :aria-busy="loading || undefined"
    class="app-btn slide-fill"
    :class="[
      `app-btn--${variant}`,
      `app-btn--${size}`,
      { 'app-btn--block': block, 'app-btn--icon': iconOnly },
    ]"
    @click.capture="blockInertLink"
  >
    <AppSpinner v-if="loading" />
    <slot v-else />
  </component>
</template>

<script setup lang="ts">
import AppSpinner from "@/components/ui/AppSpinner.vue"
import { computed, onMounted, useAttrs } from "vue"
import { RouterLink, type RouteLocationRaw } from "vue-router"

const props = withDefaults(
  defineProps<{
    variant?: "primary" | "outline" | "ghost" | "danger"
    size?: "sm" | "md" | "lg"
    type?: "button" | "submit" | "reset"
    to?: RouteLocationRaw
    href?: string
    disabled?: boolean
    loading?: boolean
    block?: boolean
    /** Square button with an icon only; needs an aria-label. */
    iconOnly?: boolean
  }>(),
  { variant: "primary", size: "md", type: "button" },
)

const isLink = computed(() => Boolean(props.to || props.href))
const isInert = computed(() => props.disabled || props.loading)
// An inert link drops its href (RouterLink would always render one), so it is
// no longer a navigable target for middle-click, drag, or "open in new tab".
const isRouterLink = computed(() => Boolean(props.to) && !isInert.value)
const linkHref = computed(() => (isInert.value ? undefined : props.href))
const tag = computed(() => (isRouterLink.value ? RouterLink : isLink.value ? "a" : "button"))
const isNativeButton = computed(() => !isLink.value)
const attrs = useAttrs()

// Links have no disabled attribute, so block activation by hand.
function blockInertLink(event: Event) {
  if (isNativeButton.value || !isInert.value) return
  event.preventDefault()
  event.stopImmediatePropagation()
}

onMounted(() => {
  if (import.meta.env.DEV && props.iconOnly && !attrs["aria-label"] && !attrs["aria-labelledby"]) {
    console.warn("[AppButton] icon-only buttons need an aria-label.")
  }
})
</script>

<style scoped>
.app-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  box-sizing: border-box;
  min-height: 44px;
  padding: 10px 20px;
  border: 1px solid transparent;
  border-radius: var(--radius-btn);
  font-size: 14px;
  font-weight: 700;
  text-decoration: none;
  cursor: pointer;
}

.app-btn:disabled,
.app-btn[aria-disabled="true"] {
  opacity: 0.55;
  cursor: not-allowed;
}

.app-btn--block {
  display: flex;
  width: 100%;
}

.app-btn--sm {
  min-height: 36px;
  padding: 6px 14px;
  font-size: 13px;
}

/* Touch targets stay at 44px. */
@media (pointer: coarse) {
  .app-btn--sm {
    min-height: 44px;
  }
}

.app-btn--lg {
  min-height: 48px;
  font-size: 15px;
}

.app-btn--icon {
  min-width: 44px;
  padding: 0;
}

.app-btn--icon.app-btn--sm {
  min-width: 36px;
}

.app-btn--icon.app-btn--lg {
  min-width: 48px;
}

@media (pointer: coarse) {
  .app-btn--icon.app-btn--sm {
    min-width: 44px;
  }
}

.app-btn--primary {
  --slide-fill: var(--color-accent);
  background: var(--color-brand);
  color: var(--color-on-brand);
}

.app-btn--outline {
  --slide-fill: var(--color-accent);
  background: transparent;
  border-color: var(--color-border-strong);
  color: var(--color-text-primary);
}

.app-btn--ghost {
  --slide-fill: var(--color-surface-3);
  --slide-fill-text: var(--color-text-primary);
  background: transparent;
  color: var(--color-text-secondary);
}

.app-btn--danger {
  --slide-fill: var(--color-accent);
  background: var(--color-danger);
  color: var(--color-on-brand);
}
</style>
