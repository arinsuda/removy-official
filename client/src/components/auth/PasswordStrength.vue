<template>
  <div v-if="password" class="strength" :class="`strength--${strength.level}`">
    <div class="strength__bars" aria-hidden="true">
      <span
        v-for="i in 4"
        :key="i"
        class="strength__bar"
        :class="{ 'is-active': i <= strength.score }"
      />
    </div>
    <span class="strength__label">{{ strength.label }}</span>
  </div>
</template>

<script setup lang="ts">
import { computed } from "vue"

const props = defineProps<{ password: string }>()

const strength = computed(() => {
  const p = props.password
  let score = 0
  if (p.length >= 8) score++
  if (p.length >= 12) score++
  if (/[A-Z]/.test(p) && /[a-z]/.test(p)) score++
  if (/[0-9]/.test(p)) score++
  if (/[^A-Za-z0-9]/.test(p)) score++

  if (score <= 1) return { score: 1, level: "weak", label: "Weak" }
  if (score === 2) return { score: 2, level: "fair", label: "Fair" }
  if (score === 3) return { score: 3, level: "good", label: "Good" }
  return { score: 4, level: "strong", label: "Strong" }
})
</script>

<style scoped>
.strength {
  display: flex;
  align-items: center;
  gap: 8px;
}

.strength--weak {
  --strength-color: #ef4444;
}

.strength--fair {
  --strength-color: #f97316;
}

.strength--good {
  --strength-color: #eab308;
}

.strength--strong {
  --strength-color: #22c55e;
}

.strength__bars {
  display: flex;
  flex: 1;
  gap: 4px;
}

.strength__bar {
  flex: 1;
  height: 3px;
  background: rgb(255 255 255 / 10%);
  border-radius: 99px;
  transition: background var(--motion-base) var(--ease-out-soft);
}

.strength__bar.is-active {
  background: var(--strength-color);
}

.strength__label {
  min-width: 44px;
  font-size: 0.6875rem;
  font-weight: 600;
  letter-spacing: 0.5px;
  text-align: right;
  text-transform: uppercase;
  color: var(--strength-color);
}
</style>
