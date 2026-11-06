<template>
  <component
    :is="tag"
    :to="to"
    v-bind="href ? { href } : {}"
    :type="isNativeButton ? type : undefined"
    :disabled="isNativeButton ? isInert : undefined"
    :aria-busy="loading || undefined"
    class="app-btn slide-fill"
    :class="[`app-btn--${variant}`, `app-btn--${size}`, { 'app-btn--block': block }]"
  >
    <AppSpinner v-if="loading" />
    <slot v-else />
  </component>
</template>

<script setup lang="ts">
import AppSpinner from "@/components/ui/AppSpinner.vue"
import { computed } from "vue"
import { RouterLink, type RouteLocationRaw } from "vue-router"

const props = withDefaults(
  defineProps<{
    variant?: "primary" | "outline"
    size?: "md" | "lg"
    type?: "button" | "submit" | "reset"
    to?: RouteLocationRaw
    href?: string
    disabled?: boolean
    loading?: boolean
    block?: boolean
  }>(),
  { variant: "primary", size: "md", type: "button" },
)

const tag = computed(() => (props.to ? RouterLink : props.href ? "a" : "button"))
const isNativeButton = computed(() => tag.value === "button")
const isInert = computed(() => props.disabled || props.loading)
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

.app-btn:disabled {
  opacity: 0.55;
  cursor: not-allowed;
}

.app-btn--block {
  display: flex;
  width: 100%;
}

.app-btn--lg {
  min-height: 48px;
  font-size: 15px;
}

.app-btn--primary {
  background: var(--color-brand);
  color: #fff;
}

.app-btn--outline {
  background: transparent;
  border-color: var(--color-border-strong);
  color: var(--color-text-primary);
}

.app-btn__spinner {
  width: 18px;
  height: 18px;
  border: 2px solid rgb(255 255 255 / 30%);
  border-top-color: #fff;
  border-radius: 50%;
  animation: app-btn-spin 0.7s linear infinite;
}

@keyframes app-btn-spin {
  to {
    transform: rotate(360deg);
  }
}
</style>
