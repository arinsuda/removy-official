<template>
  <RouterLink :to="{ name: 'actor-detail', params: { id: person.id } }" class="person">
    <span class="person__photo">
      <img v-if="person.profile_path && !failed" :src="photo" alt="" loading="lazy" decoding="async" @error="failed = true" />
      <User v-else :size="32" aria-hidden="true" />
    </span>
    <span class="person__name">
      <template v-for="(part, i) in parts" :key="i">
        <mark v-if="part.match" class="person__mark">{{ part.text }}</mark>
        <template v-else>{{ part.text }}</template>
      </template>
    </span>
    <span v-if="department" class="person__dept">{{ department }}</span>
  </RouterLink>
</template>

<script setup lang="ts">
import { computed, ref } from "vue"
import { useI18n } from "vue-i18n"
import { User } from "lucide-vue-next"
import { KNOWN_DEPARTMENTS } from "@/components/actor/actorConfig"
import type { PersonSearchResult } from "@/types"
import { highlightParts } from "./highlight"

const props = defineProps<{ person: PersonSearchResult; query: string }>()
const { t, te } = useI18n()
const failed = ref(false)

const photo = computed(() => `https://image.tmdb.org/t/p/w185${props.person.profile_path}`)
const parts = computed(() => highlightParts(props.person.name, props.query))
const department = computed(() => {
  const d = props.person.known_for_department
  return KNOWN_DEPARTMENTS.includes(d) && te(`actor.departments.${d}`) ? t(`actor.departments.${d}`) : d
})
</script>

<style scoped>
.person {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0.375rem;
  min-width: 0;
  padding: 1rem 0.75rem;
  border: 1px solid var(--color-border);
  border-radius: var(--radius-card);
  background: var(--color-surface-2);
  color: inherit;
  text-align: center;
  text-decoration: none;
  transition:
    border-color var(--motion-fast) var(--ease-out-soft),
    background var(--motion-fast) var(--ease-out-soft),
    transform var(--motion-fast) var(--ease-out-soft);
}
.person__photo {
  display: grid;
  place-items: center;
  width: 5.5rem;
  aspect-ratio: 1;
  overflow: hidden;
  border-radius: 50%;
  background: var(--color-surface-3);
  color: var(--color-text-muted);
}
.person__photo img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}
.person__name {
  max-width: 100%;
  overflow: hidden;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  font-size: var(--text-small);
  font-weight: 600;
  line-height: 1.3;
  color: var(--color-text-primary);
}
.person__mark {
  background: color-mix(in srgb, var(--color-brand) 28%, transparent);
  border-radius: 2px;
  color: inherit;
}
.person__dept {
  font-size: var(--text-caption);
  color: var(--color-text-muted);
}
@media (hover: hover) {
  .person:hover {
    border-color: var(--color-border-strong);
    background: var(--color-surface-3);
    transform: translateY(-2px);
  }
}
@media (prefers-reduced-motion: reduce) {
  .person {
    transition: none;
  }
  .person:hover {
    transform: none;
  }
}
</style>
