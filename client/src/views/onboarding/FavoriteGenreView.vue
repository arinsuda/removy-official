<template>
  <OnboardingShell :reduced="reduced" :leaving="leaving">
    <OnboardingHeader :count="count" :min="MIN_PICKS" :genre-labels="genreLabels" />

    <PosterGrid v-if="status === 'ready'" :movies="movies" :selected-ids="selectedIds" :disabled="saving || leaving" @toggle="toggle" />

    <ul v-else-if="status === 'loading'" class="skeletons" :aria-label="t('common.loading')" aria-busy="true">
      <li v-for="i in TOTAL_MOVIES" :key="i"><SkeletonBlock aspect="2 / 3" radius="12px" /></li>
    </ul>

    <AppErrorState v-else :title="t('onboarding.loadError')" @retry="load" />

    <OnboardingActionBar
      :ready="canContinue"
      :busy="saving || leaving"
      :remaining="remaining"
      :failed="saveFailed"
      @continue="onContinue"
      @skip="onSkip"
    />
  </OnboardingShell>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from "vue"
import { useRouter } from "vue-router"
import { useI18n } from "vue-i18n"
import OnboardingShell from "@/components/onboarding/OnboardingShell.vue"
import OnboardingHeader from "@/components/onboarding/OnboardingHeader.vue"
import OnboardingActionBar from "@/components/onboarding/OnboardingActionBar.vue"
import PosterGrid from "@/components/onboarding/PosterGrid.vue"
import AppErrorState from "@/components/ui/AppErrorState.vue"
import SkeletonBlock from "@/components/ui/skeleton/SkeletonBlock.vue"
import { EXIT_MS, MIN_PICKS, TOTAL_MOVIES, genreLabelKey } from "@/components/onboarding/onboardingConfig"
import { useOnboardingPicks } from "@/composables/useOnboardingPicks"
import { useReducedMotion } from "@/composables/useReducedMotion"

const { t } = useI18n()
const router = useRouter()
const reduced = useReducedMotion()
const { movies, selectedIds, selectedGenreIds, status, saving, saveFailed, count, remaining, canContinue, load, toggle, confirm, skip } =
  useOnboardingPicks()

const leaving = ref(false)

const genreLabels = computed(() =>
  selectedGenreIds.value.flatMap((id) => {
    const key = genreLabelKey(id)
    return key ? [t(key)] : []
  }),
)

/** Plays the shell's exit transition, then moves on to the home page. */
async function leave() {
  leaving.value = true
  if (!reduced.value) await new Promise((resolve) => setTimeout(resolve, EXIT_MS))
  router.push({ name: "home" })
}

async function onContinue() {
  if (await confirm()) await leave()
}

async function onSkip() {
  await skip()
  await leave()
}

onMounted(load)
</script>

<style scoped>
.skeletons {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 0.75rem;
  margin: 0;
  padding: 1.25rem 0 2rem;
  list-style: none;
  flex: 1;
}

@media (min-width: 480px) {
  .skeletons {
    grid-template-columns: repeat(3, minmax(0, 1fr));
  }
}
@media (min-width: 768px) {
  .skeletons {
    grid-template-columns: repeat(4, minmax(0, 1fr));
    gap: 1rem;
  }
}
@media (min-width: 1024px) {
  .skeletons {
    grid-template-columns: repeat(6, minmax(0, 1fr));
  }
}
</style>
