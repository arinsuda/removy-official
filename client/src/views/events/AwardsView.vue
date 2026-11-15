<template>
  <section class="awards-page">
    <header class="awards-page__header">
      <div class="awards-page__heading">
        <h1 class="awards-page__title">{{ t("awards.title") }}</h1>
        <p class="awards-page__subtitle">{{ t("awards.subtitle") }}</p>
      </div>
      <AwardsFilters v-model:media-type="mediaType" v-model:year="year" :years="years" />
    </header>

    <div v-if="query.isPending.value" class="awards-page__loading">
      <AppSpinner :size="32" />
    </div>

    <AppErrorState
      v-else-if="query.isError.value"
      :title="t('common.error')"
      :description="t('awards.loadErrorDesc')"
      @retry="query.refetch()"
    />

    <AppEmptyState
      v-else-if="groups.length === 0"
      :title="t('awards.empty.title')"
      :description="t('awards.empty.description')"
      :icon="Trophy"
    />

    <div v-else class="awards-page__groups">
      <section
        v-for="group in groups"
        :key="`${group.kind}:${group.period}`"
        class="awards-group"
      >
        <h2 class="awards-group__title">
          {{ t(`awards.group.${group.kind}`, { period: formatAwardPeriod(group.period, locale, 'long') }) }}
        </h2>
        <div class="awards-group__grid">
          <AwardWinnerCard
            v-for="award in group.awards"
            :key="`${award.event_id}:${award.media_id}`"
            :award="award"
            :co-winner="group.awards.length > 1"
          />
        </div>
      </section>
    </div>
  </section>
</template>

<script setup lang="ts">
import { useI18n } from "vue-i18n"
import { Trophy } from "lucide-vue-next"
import AppEmptyState from "@/components/ui/AppEmptyState.vue"
import AppErrorState from "@/components/ui/AppErrorState.vue"
import AppSpinner from "@/components/ui/AppSpinner.vue"
import AwardsFilters from "@/components/events/awards/AwardsFilters.vue"
import AwardWinnerCard from "@/components/events/awards/AwardWinnerCard.vue"
import { useEventAwards } from "@/composables/useEventAwards"
import { formatAwardPeriod } from "@/utils/events"

const { t, locale } = useI18n()
const { mediaType, year, query, years, groups } = useEventAwards()
</script>

<style scoped>
.awards-page {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
  width: 100%;
  max-width: var(--container-page);
  margin: 0 auto;
  padding: 1.5rem 1rem 3rem;
}

.awards-page__header {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

@media (min-width: 768px) {
  .awards-page__header {
    flex-direction: row;
    align-items: flex-end;
    justify-content: space-between;
  }
}

.awards-page__heading {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.awards-page__title {
  font-size: var(--text-page);
  font-weight: 700;
  color: var(--color-text-primary);
  line-height: 1.2;
}

.awards-page__subtitle {
  font-size: var(--text-small);
  color: var(--color-text-muted);
}

.awards-page__loading {
  display: flex;
  justify-content: center;
  align-items: center;
  padding: 6rem 0;
}

.awards-page__groups {
  display: flex;
  flex-direction: column;
  gap: 2rem;
}

.awards-group {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.awards-group__title {
  font-size: var(--text-title);
  font-weight: 600;
  color: var(--color-text-primary);
}

.awards-group__grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 1rem;
}

@media (min-width: 640px) {
  .awards-group__grid {
    grid-template-columns: repeat(3, 1fr);
  }
}

@media (min-width: 1024px) {
  .awards-group__grid {
    grid-template-columns: repeat(5, 1fr);
  }
}
</style>
