<template>
  <section class="filmo" :aria-labelledby="titleId">
    <header class="filmo__head">
      <h2 :id="titleId" class="filmo__title">{{ t("actor.filmography.title") }}</h2>
      <p class="filmo__count" role="status">{{ t("actor.filmography.count", { n: visible.length }, visible.length) }}</p>
    </header>

    <div class="filmo__toolbar">
      <AppTabs v-model="role" :tabs="roleTabs" :label="t('actor.filmography.roleLabel')" :controls="false">
        <template #tab="{ tab }">
          {{ tab.label }} <span class="filmo__badge">{{ n(counts[tab.key as CreditRole]) }}</span>
        </template>
      </AppTabs>

      <div class="filmo__options">
        <AppTabs v-model="filter" :tabs="filterTabs" :label="t('actor.filmography.filterLabel')" :controls="false" />
        <label class="filmo__sort">
          <span class="sr-only">{{ t("actor.filmography.sortLabel") }}</span>
          <select v-model="sort" class="filmo__select" data-focus-custom>
            <option v-for="s in CREDIT_SORTS" :key="s" :value="s">{{ t(`actor.filmography.sort.${s}`) }}</option>
          </select>
          <ChevronDown :size="14" class="filmo__select-icon" aria-hidden="true" />
        </label>
        <AppTabs v-model="view" :tabs="viewTabs" :label="t('actor.filmography.viewLabel')" :controls="false" />
      </div>
    </div>

    <AppEmptyState
      v-if="!visible.length"
      :icon="Film"
      :title="t(roleCredits.length ? 'actor.filmography.empty.filtered' : `actor.filmography.empty.${role}`)"
    />
    <template v-else-if="view === 'grid'">
      <ActorCreditGrid :credits="visible.slice(0, shown)" />
      <div v-if="visible.length > shown" class="filmo__more">
        <AppButton variant="outline" @click="shown += GRID_PAGE_SIZE">{{ t("actor.filmography.showMore") }}</AppButton>
      </div>
    </template>
    <ActorTimeline v-else :credits="visible" :ascending="sort === 'oldest'" />
  </section>
</template>

<script setup lang="ts">
import { computed, ref, useId, watch } from "vue"
import { useI18n } from "vue-i18n"
import { ChevronDown, Film } from "lucide-vue-next"
import AppButton from "@/components/ui/AppButton.vue"
import AppEmptyState from "@/components/ui/AppEmptyState.vue"
import AppTabs from "@/components/ui/AppTabs.vue"
import ActorCreditGrid from "./ActorCreditGrid.vue"
import ActorTimeline from "./ActorTimeline.vue"
import {
  CREDIT_FILTERS,
  CREDIT_ROLES,
  CREDIT_SORTS,
  CREDIT_VIEWS,
  GRID_PAGE_SIZE,
  type CreditFilter,
  type CreditRole,
  type CreditSort,
  type CreditView,
} from "./actorConfig"
import { filterCredits, sortCredits, type ActorCredit } from "./actorCredits"

const props = defineProps<{ acting: ActorCredit[]; crew: ActorCredit[] }>()
const role = defineModel<CreditRole>("role", { required: true })

const { t, n } = useI18n()
const titleId = useId()

const filter = ref<CreditFilter>("all")
const sort = ref<CreditSort>("popular")
const view = ref<CreditView>("grid")
const shown = ref(GRID_PAGE_SIZE)

const counts = computed(() => ({ acting: props.acting.length, crew: props.crew.length }))
const roleCredits = computed(() => (role.value === "acting" ? props.acting : props.crew))
const visible = computed(() => sortCredits(filterCredits(roleCredits.value, filter.value), sort.value))

watch([role, filter, sort], () => (shown.value = GRID_PAGE_SIZE))

const roleTabs = computed(() => CREDIT_ROLES.map((key) => ({ key, label: t(`actor.filmography.roles.${key}`) })))
const filterTabs = computed(() => CREDIT_FILTERS.map((key) => ({ key, label: t(`actor.filmography.filter.${key}`) })))
const viewTabs = computed(() => CREDIT_VIEWS.map((key) => ({ key, label: t(`actor.filmography.view.${key}`) })))
</script>

<style scoped>
.filmo__head {
  display: flex;
  align-items: baseline;
  justify-content: space-between;
  gap: 1rem;
  margin-bottom: 1rem;
}
.filmo__title {
  margin: 0;
  font-size: var(--text-title);
  font-weight: 700;
  color: var(--color-text-primary);
}
.filmo__count {
  margin: 0;
  font-size: var(--text-caption);
  color: var(--color-text-muted);
}
.filmo__toolbar {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
  margin-bottom: 1.25rem;
}
.filmo__options {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 0.75rem;
}
.filmo__badge {
  margin-left: 0.25rem;
  font-variant-numeric: tabular-nums;
  opacity: 0.7;
}
.filmo__sort {
  position: relative;
  display: inline-flex;
  align-items: center;
}
.filmo__select {
  appearance: none;
  padding: 0.375rem 2rem 0.375rem 0.875rem;
  border: 1px solid var(--color-border);
  border-radius: var(--radius-btn);
  background: var(--color-surface-2);
  font-family: var(--font-sans);
  font-size: var(--text-caption);
  color: var(--color-text-secondary);
  cursor: pointer;
}
.filmo__select:hover {
  border-color: var(--color-border-strong);
  color: var(--color-text-primary);
}
.filmo__select:focus-visible {
  outline: 2px solid var(--color-focus);
  outline-offset: 2px;
}
.filmo__select-icon {
  position: absolute;
  right: 0.75rem;
  color: var(--color-text-muted);
  pointer-events: none;
}
.filmo__more {
  display: flex;
  justify-content: center;
  margin-top: 1.5rem;
}

@media (min-width: 1024px) {
  .filmo__toolbar {
    flex-direction: row;
    align-items: center;
    justify-content: space-between;
  }
}
</style>
