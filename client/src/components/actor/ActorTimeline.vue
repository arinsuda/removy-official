<template>
  <ol class="timeline">
    <li v-for="group in groups" :key="group.year ?? 'unknown'" class="timeline__year">
      <h3 class="timeline__label">{{ group.year ?? t("actor.filmography.unknownYear") }}</h3>
      <ul class="timeline__list">
        <li v-for="credit in group.credits" :key="credit.key">
          <RouterLink :to="detailRoute(credit.media)" class="timeline__row">
            <img
              v-if="credit.media.poster_path"
              :src="posterSrc(credit.media.poster_path)"
              alt=""
              class="timeline__poster"
              loading="lazy"
              decoding="async"
            />
            <span v-else class="timeline__poster timeline__poster--empty" aria-hidden="true"><Film :size="16" /></span>
            <span class="timeline__text">
              <span class="timeline__title">{{ credit.media.title }}</span>
              <span class="timeline__sub">
                {{ t(`actor.credit.${credit.media.type}`) }}
                <template v-if="credit.character"> · {{ t("actor.credit.asCharacter", { character: credit.character }) }}</template>
              </span>
            </span>
          </RouterLink>
        </li>
      </ul>
    </li>
  </ol>
</template>

<script setup lang="ts">
import { computed } from "vue"
import { useI18n } from "vue-i18n"
import { Film } from "lucide-vue-next"
import { detailRoute, posterSrc } from "@/components/home/homeMedia"
import { groupByYear, type ActorCredit } from "./actorCredits"

const props = defineProps<{
  /** Already sorted. */
  credits: ActorCredit[]
  ascending: boolean
}>()

const { t } = useI18n()
const groups = computed(() => groupByYear(props.credits, props.ascending))
</script>

<style scoped>
.timeline {
  position: relative;
  margin: 0;
  padding: 0 0 0 1.25rem;
  list-style: none;
  border-left: 2px solid var(--color-border-strong);
}
.timeline__year {
  position: relative;
  padding-bottom: 1.5rem;
}
.timeline__year::before {
  content: "";
  position: absolute;
  top: 0.4rem;
  left: calc(-1.25rem - 6px);
  width: 10px;
  height: 10px;
  border: 2px solid var(--color-surface);
  border-radius: 9999px;
  background: var(--color-brand);
}
.timeline__label {
  margin: 0 0 0.5rem;
  font-size: var(--text-body);
  font-weight: 800;
  color: var(--color-text-primary);
  font-variant-numeric: tabular-nums;
}
.timeline__list {
  display: grid;
  gap: 0.5rem;
  margin: 0;
  padding: 0;
  list-style: none;
}
.timeline__row {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 0.5rem;
  border: 1px solid var(--color-border);
  border-radius: var(--radius-card);
  background: var(--color-surface-2);
  color: inherit;
  text-decoration: none;
  transition:
    border-color var(--motion-fast) ease,
    background-color var(--motion-fast) ease;
}
.timeline__row:hover {
  border-color: var(--color-border-strong);
  background: var(--color-surface-3);
}
.timeline__poster {
  flex-shrink: 0;
  width: 2.5rem;
  aspect-ratio: 2 / 3;
  border-radius: 0.375rem;
  object-fit: cover;
  background: var(--color-surface-3);
}
.timeline__poster--empty {
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--color-text-muted);
}
.timeline__text {
  display: flex;
  min-width: 0;
  flex-direction: column;
  gap: 0.125rem;
}
.timeline__title {
  overflow: hidden;
  font-size: var(--text-small);
  font-weight: 600;
  color: var(--color-text-primary);
  text-overflow: ellipsis;
  white-space: nowrap;
}
.timeline__sub {
  overflow: hidden;
  font-size: var(--text-caption);
  color: var(--color-text-muted);
  text-overflow: ellipsis;
  white-space: nowrap;
}

@media (min-width: 768px) {
  .timeline__list {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}
</style>
