<template>
  <main class="actor-page">
    <AppButton variant="ghost" size="sm" class="actor-page__back" @click="goBack">
      <ArrowLeft :size="16" aria-hidden="true" />
      {{ t("actor.back") }}
    </AppButton>

    <ActorSkeleton v-if="isLoading" />

    <AppEmptyState v-else-if="isNotFound" :icon="UserX" :title="t('actor.notFoundTitle')" :description="t('actor.notFoundDetail')">
      <template #action>
        <AppButton to="/" variant="outline" size="sm">{{ t("actor.backToHome") }}</AppButton>
      </template>
    </AppEmptyState>

    <AppErrorState v-else-if="isError" :title="t('actor.errorTitle')" :description="t('actor.errorDetail')" @retry="fetchData" />

    <div v-else-if="actor" class="actor-page__content">
      <ActorHero :actor="actor" />
      <ActorStats :stats="stats" />
      <ActorKnownFor :credits="knownForCredits" />
      <ActorFilmography v-model:role="role" :acting="acting" :crew="crew" />
    </div>
  </main>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, watch } from "vue"
import { useI18n } from "vue-i18n"
import { useRoute, useRouter } from "vue-router"
import { ArrowLeft, UserX } from "lucide-vue-next"
import { actorApi } from "@/api/api"
import ActorFilmography from "@/components/actor/ActorFilmography.vue"
import ActorHero from "@/components/actor/ActorHero.vue"
import ActorKnownFor from "@/components/actor/ActorKnownFor.vue"
import ActorSkeleton from "@/components/actor/ActorSkeleton.vue"
import ActorStats from "@/components/actor/ActorStats.vue"
import { KNOWN_FOR_COUNT, type CreditRole } from "@/components/actor/actorConfig"
import { actingCredits, creditStats, crewCredits, knownFor, type ActorCredit } from "@/components/actor/actorCredits"
import AppButton from "@/components/ui/AppButton.vue"
import AppEmptyState from "@/components/ui/AppEmptyState.vue"
import AppErrorState from "@/components/ui/AppErrorState.vue"
import type { PersonDetail } from "@/types"
import { provideHomePopup } from "@/composables/useHomePopup"
import { logApiError } from "@/utils/logApiError"

provideHomePopup()

const route = useRoute()
const router = useRouter()
const { t } = useI18n()

const actor = ref<PersonDetail | null>(null)
const acting = ref<ActorCredit[]>([])
const crew = ref<ActorCredit[]>([])
const role = ref<CreditRole>("acting")

const isLoading = ref(true)
const isError = ref(false)
const isNotFound = ref(false)

function goBack() {
  if (window.history.length > 1) router.back()
  else router.push("/")
}

// Ignores a slow response for an actor the user has already navigated away from.
let requestId = 0

async function fetchData() {
  const actorId = Number(route.params.id)
  if (!actorId || Number.isNaN(actorId)) {
    isNotFound.value = true
    isLoading.value = false
    return
  }

  const current = ++requestId
  isLoading.value = true
  isError.value = false
  isNotFound.value = false

  try {
    const [detail, movies, series] = await Promise.all([
      actorApi.getById(actorId),
      actorApi.getMovies(actorId),
      actorApi.getTVSeries(actorId),
    ])
    if (current !== requestId) return
    actor.value = detail.data
    acting.value = actingCredits(movies.data?.cast ?? [], series.data?.cast ?? [])
    crew.value = crewCredits(movies.data?.crew ?? [], series.data?.crew ?? [])
    // A director with no acting credits should land on the tab that has content.
    role.value = !acting.value.length && crew.value.length ? "crew" : "acting"
  } catch (err) {
    if (current !== requestId) return
    const status = (err as { response?: { status?: number } })?.response?.status
    if (status === 404) isNotFound.value = true
    else {
      logApiError("actor load failed:", err)
      isError.value = true
    }
  } finally {
    if (current === requestId) isLoading.value = false
  }
}

onMounted(fetchData)
watch(() => route.params.id, fetchData)

const stats = computed(() => creditStats(role.value === "acting" ? acting.value : crew.value))
const knownForCredits = computed(() => knownFor(acting.value.length ? acting.value : crew.value, KNOWN_FOR_COUNT))
</script>

<style scoped>
.actor-page {
  display: block;
  max-width: 1240px;
  margin: 0 auto;
  padding: 1rem 1rem 4rem;
  color: var(--color-text-primary);
}
.actor-page__back {
  margin-bottom: 1rem;
}
.actor-page__content {
  display: flex;
  flex-direction: column;
  gap: 2rem;
}

@media (min-width: 768px) {
  .actor-page {
    padding: 1.5rem 1.5rem 5rem;
  }
}
</style>
