<template>
  <header class="onb-head">
    <p class="onb-head__brand">{{ t("onboarding.brand") }}</p>
    <h1 class="onb-head__title">{{ t("onboarding.heading") }}</h1>
    <p class="onb-head__sub">{{ t("onboarding.subheading", { min }) }}</p>

    <div
      class="onb-head__track"
      role="progressbar"
      :aria-valuemin="0"
      :aria-valuemax="min"
      :aria-valuenow="Math.min(count, min)"
      :aria-label="t('onboarding.progressLabel')"
    >
      <span class="onb-head__fill" :class="{ 'onb-head__fill--done': done }" :style="{ transform: `scaleX(${progress})` }" />
    </div>

    <i18n-t keypath="onboarding.progress" tag="p" class="onb-head__count" :class="{ 'onb-head__count--done': done }" aria-hidden="true">
      <template #count>
        <span class="roll">
          <Transition name="roll" mode="out-in">
            <span :key="count" class="roll__digit">{{ count }}</span>
          </Transition>
        </span>
      </template>
      <template #min>{{ min }}</template>
    </i18n-t>
    <p class="sr-only" role="status" aria-live="polite">{{ t("onboarding.announce", { count, min }) }}</p>

    <ul class="onb-head__chips" :aria-label="t('onboarding.genresLabel')">
      <TransitionGroup name="chip">
        <li v-for="label in genreLabels" :key="label" class="onb-head__chip">{{ label }}</li>
      </TransitionGroup>
    </ul>
  </header>
</template>

<script setup lang="ts">
import { computed } from "vue"
import { useI18n } from "vue-i18n"

const props = defineProps<{ count: number; min: number; genreLabels: string[] }>()
const { t } = useI18n()

const progress = computed(() => Math.min(props.count / props.min, 1))
const done = computed(() => props.count >= props.min)
</script>

<style scoped>
.onb-head {
  text-align: center;
  animation: onb-rise var(--motion-wave) var(--ease-out-soft) both;
}

.onb-head__brand {
  margin: 0 0 0.75rem;
  font-size: 0.8125rem;
  font-weight: 800;
  letter-spacing: 0.32em;
  color: var(--color-brand);
}

.onb-head__title {
  margin: 0 0 0.5rem;
  font-size: 1.625rem;
  font-weight: 800;
  line-height: 1.2;
  color: var(--color-text-primary);
}

.onb-head__sub {
  max-width: 34rem;
  margin: 0 auto 1.25rem;
  font-size: 0.9375rem;
  line-height: 1.6;
  color: var(--color-text-secondary);
}

.onb-head__track {
  width: 100%;
  max-width: 20rem;
  height: 4px;
  margin: 0 auto 0.6rem;
  overflow: hidden;
  border-radius: 9999px;
  background: var(--color-border-strong);
}

/* Filled by scaleX so the bar never triggers layout. */
.onb-head__fill {
  display: block;
  height: 100%;
  border-radius: inherit;
  background: var(--color-brand);
  transform-origin: left center;
  transition:
    transform var(--motion-base) var(--ease-spring),
    background-color var(--motion-base) var(--ease-water);
}
.onb-head__fill--done {
  background: var(--color-success);
}

.onb-head__count {
  margin: 0;
  font-size: 0.875rem;
  color: var(--color-text-muted);
  transition: color var(--motion-base) var(--ease-water);
}
.onb-head__count--done {
  color: var(--color-success);
}

/* Counter roll: the old digit slides up and out, the new one rises in. */
.roll {
  display: inline-block;
  overflow: hidden;
  vertical-align: bottom;
  font-weight: 800;
  font-variant-numeric: tabular-nums;
  min-width: 1ch;
}
.roll__digit {
  display: inline-block;
}
.roll-enter-active,
.roll-leave-active {
  transition:
    transform var(--motion-fast) var(--ease-out-soft),
    opacity var(--motion-fast) var(--ease-out-soft);
}
.roll-enter-from {
  opacity: 0;
  transform: translateY(70%);
}
.roll-leave-to {
  opacity: 0;
  transform: translateY(-70%);
}

.onb-head__chips {
  display: flex;
  flex-wrap: wrap;
  justify-content: center;
  gap: 0.4rem;
  min-height: 1.75rem;
  margin: 0.75rem 0 0;
  padding: 0;
  list-style: none;
}
.onb-head__chip {
  padding: 0.2rem 0.7rem;
  border: 1px solid color-mix(in srgb, var(--color-brand) 45%, transparent);
  border-radius: 9999px;
  background: color-mix(in srgb, var(--color-brand) 14%, transparent);
  font-size: 0.75rem;
  font-weight: 600;
  color: var(--color-text-primary);
}
.chip-enter-active {
  transition:
    transform var(--motion-base) var(--ease-spring),
    opacity var(--motion-base) var(--ease-out-soft);
}
.chip-leave-active {
  position: absolute;
  transition: opacity var(--motion-fast) var(--ease-water);
}
.chip-move {
  transition: transform var(--motion-base) var(--ease-out-soft);
}
.chip-enter-from {
  opacity: 0;
  transform: scale(0.6) translateY(4px);
}
.chip-leave-to {
  opacity: 0;
}

@keyframes onb-rise {
  from {
    opacity: 0;
    transform: translateY(-16px);
  }
}

@media (min-width: 768px) {
  .onb-head__title {
    font-size: 2.25rem;
  }
}

@media (prefers-reduced-motion: reduce) {
  .onb-head,
  .onb-head__fill,
  .onb-head__count,
  .roll-enter-active,
  .roll-leave-active,
  .chip-enter-active,
  .chip-leave-active,
  .chip-move {
    animation: none;
    transition-duration: 1ms;
  }
}
</style>
