<template>
  <span
    class="app-avatar"
    :class="`app-avatar--${size}`"
    :data-avatar-fallback="showImage ? undefined : ''"
    :role="decorative ? undefined : 'img'"
    :aria-label="decorative ? undefined : name"
    :aria-hidden="decorative || undefined"
  >
    <img v-if="showImage" :src="src ?? undefined" alt="" class="app-avatar__img" @error="failed = true" />
    <span v-else class="app-avatar__initials" aria-hidden="true">{{ initials }}</span>
  </span>
</template>

<script setup lang="ts">
import { computed, ref, watch } from "vue"

const props = withDefaults(
  defineProps<{
    src?: string | null
    name: string
    size?: "sm" | "md" | "lg" | "xl"
    /** Hide from assistive tech when a visible label already names the user. */
    decorative?: boolean
  }>(),
  { src: null, size: "md", decorative: false },
)

const failed = ref(false)
watch(() => props.src, () => (failed.value = false))

const showImage = computed(() => !!props.src && !failed.value)
const initials = computed(() => {
  const parts = props.name.trim().split(/\s+/).filter(Boolean)
  if (!parts.length) return "?"
  const letters = parts.length === 1 ? Array.from(parts[0]!).slice(0, 2) : parts.slice(0, 2).map((p) => Array.from(p)[0]!)
  return letters.join("").toUpperCase()
})
</script>

<style scoped>
.app-avatar {
  display: inline-flex;
  flex-shrink: 0;
  align-items: center;
  justify-content: center;
  overflow: hidden;
  border-radius: 9999px;
  background: var(--color-surface-3);
  color: var(--color-text-secondary);
  font-weight: 600;
  box-shadow: 0 0 0 2px var(--color-border);
}
.app-avatar--sm { width: 1.75rem; height: 1.75rem; font-size: 0.6875rem; }
.app-avatar--md { width: 2.5rem; height: 2.5rem; font-size: 0.875rem; }
.app-avatar--lg { width: 3.5rem; height: 3.5rem; font-size: 1.125rem; }
.app-avatar--xl { width: 6rem; height: 6rem; font-size: 1.75rem; }
.app-avatar__img { width: 100%; height: 100%; object-fit: cover; }
</style>
