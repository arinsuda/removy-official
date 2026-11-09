<template>
  <section class="actor-hero" :aria-labelledby="nameId">
    <div class="actor-hero__photo">
      <img v-if="actor.profile_path && !photoFailed" :src="`${ACTOR_PROFILE_BASE}${actor.profile_path}`" :alt="t('actor.photo', { name: actor.name })" @error="photoFailed = true" />
      <span v-else class="actor-hero__photo-empty" role="img" :aria-label="t('actor.photo', { name: actor.name })"><User :size="48" aria-hidden="true" /></span>
    </div>

    <div class="actor-hero__info">
      <h1 :id="nameId" class="actor-hero__name">{{ actor.name }}</h1>

      <ul class="actor-hero__facts">
        <li v-if="actor.known_for_department" class="actor-hero__fact">
          <Clapperboard :size="14" aria-hidden="true" />{{ department }}
        </li>
        <li v-if="born" class="actor-hero__fact"><CalendarDays :size="14" aria-hidden="true" />{{ born }}</li>
        <li v-if="death" class="actor-hero__fact">{{ death }}</li>
        <li v-if="alive" class="actor-hero__fact">{{ alive }}</li>
        <li v-if="actor.place_of_birth" class="actor-hero__fact" :title="t('actor.birthplace')">
          <MapPin :size="14" aria-hidden="true" /><span class="sr-only">{{ t("actor.birthplace") }}: </span>{{ actor.place_of_birth }}
        </li>
        <li v-if="actor.imdb_id">
          <a :href="`${IMDB_NAME_URL}${actor.imdb_id}`" target="_blank" rel="noopener noreferrer" class="actor-hero__fact actor-hero__fact--link">
            <ExternalLink :size="14" aria-hidden="true" />{{ t("actor.imdb") }}
          </a>
        </li>
      </ul>

      <div class="actor-hero__bio">
        <h2 class="actor-hero__bio-title">{{ t("actor.biography") }}</h2>
        <template v-if="actor.biography">
          <p :id="bioId" ref="bio" class="actor-hero__bio-text" :class="{ 'is-clamped': !expanded }">{{ actor.biography }}</p>
          <button v-if="overflowing || expanded" type="button" class="actor-hero__bio-toggle" :aria-expanded="expanded" :aria-controls="bioId" @click="expanded = !expanded">
            {{ expanded ? t("actor.showLess") : t("actor.showMore") }}
            <component :is="expanded ? ChevronUp : ChevronDown" :size="14" aria-hidden="true" />
          </button>
        </template>
        <p v-else class="actor-hero__bio-empty">{{ t("actor.noBiography") }}</p>
      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed, nextTick, onMounted, ref, useId, useTemplateRef, watch } from "vue"
import { useI18n } from "vue-i18n"
import { CalendarDays, ChevronDown, ChevronUp, Clapperboard, ExternalLink, MapPin, User } from "lucide-vue-next"
import type { PersonDetail } from "@/types"
import { ACTOR_PROFILE_BASE, IMDB_NAME_URL, KNOWN_DEPARTMENTS } from "./actorConfig"
import { ageOf, parseDay } from "./actorCredits"

const props = defineProps<{ actor: PersonDetail }>()

const { t, te, locale, n } = useI18n()
const nameId = useId()
const bioId = useId()

const photoFailed = ref(false)
watch(() => props.actor.profile_path, () => (photoFailed.value = false))

const department = computed(() => {
  const d = props.actor.known_for_department
  return KNOWN_DEPARTMENTS.includes(d) && te(`actor.departments.${d}`) ? t(`actor.departments.${d}`) : d
})

// Gregorian on purpose: film years across the app are Gregorian, and Thai defaults to the Buddhist era.
const formatDay = (value: string | null) => {
  const date = parseDay(value)
  return date ? new Intl.DateTimeFormat(locale.value, { dateStyle: "long", calendar: "gregory" }).format(date) : null
}
const age = computed(() => ageOf(props.actor.birthday, props.actor.deathday))
const born = computed(() => {
  const date = formatDay(props.actor.birthday)
  return date ? t("actor.bornOn", { date }) : null
})
const alive = computed(() => (age.value && !age.value.deceased ? t("actor.age", { n: n(age.value.age) }, age.value.age) : null))
const death = computed(() => {
  const date = formatDay(props.actor.deathday)
  if (!date) return null
  const base = t("actor.diedOn", { date })
  return age.value?.deceased ? `${base} (${t("actor.agedAtDeath", { n: n(age.value.age) })})` : base
})

// Biography clamps to a few lines; the toggle appears only when the text really overflows.
const expanded = ref(false)
const overflowing = ref(false)
const bio = useTemplateRef<HTMLElement>("bio")
async function measure() {
  await nextTick()
  const el = bio.value
  overflowing.value = !!el && !expanded.value && el.scrollHeight > el.clientHeight + 1
}
onMounted(measure)
watch(() => props.actor.biography, measure)
watch(locale, measure)
</script>

<style scoped>
.actor-hero {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 1.5rem;
  padding: 1.25rem;
  border: 1px solid var(--color-border);
  border-radius: var(--radius-card);
  background: var(--color-surface-2);
}
.actor-hero__photo {
  flex-shrink: 0;
  width: 9.5rem;
  aspect-ratio: 2 / 3;
  overflow: hidden;
  border: 1px solid var(--color-border-strong);
  border-radius: var(--radius-card);
  background: var(--color-surface-3);
}
.actor-hero__photo img {
  display: block;
  width: 100%;
  height: 100%;
  object-fit: cover;
}
.actor-hero__photo-empty {
  display: flex;
  width: 100%;
  height: 100%;
  align-items: center;
  justify-content: center;
  color: var(--color-text-muted);
}
.actor-hero__info {
  display: flex;
  min-width: 0;
  flex: 1;
  flex-direction: column;
  align-items: center;
  text-align: center;
}
.actor-hero__name {
  margin: 0 0 0.75rem;
  font-size: var(--text-heading);
  font-weight: 800;
  line-height: 1.2;
  color: var(--color-text-primary);
}
.actor-hero__facts {
  display: flex;
  flex-wrap: wrap;
  justify-content: center;
  gap: 0.5rem;
  margin: 0 0 1.25rem;
  padding: 0;
  list-style: none;
}
.actor-hero__fact {
  display: inline-flex;
  align-items: center;
  gap: 0.375rem;
  padding: 0.25rem 0.75rem;
  border: 1px solid var(--color-border);
  border-radius: var(--radius-btn);
  background: var(--color-surface);
  font-size: var(--text-caption);
  color: var(--color-text-secondary);
  text-decoration: none;
}
.actor-hero__fact svg {
  flex-shrink: 0;
  color: var(--color-text-muted);
}
.actor-hero__fact--link:hover {
  border-color: var(--color-border-strong);
  color: var(--color-text-primary);
}
.actor-hero__bio {
  width: 100%;
  max-width: 48rem;
  text-align: left;
}
.actor-hero__bio-title {
  margin: 0 0 0.5rem;
  font-size: var(--text-small);
  font-weight: 700;
  color: var(--color-text-primary);
}
.actor-hero__bio-text {
  margin: 0;
  font-size: var(--text-small);
  line-height: 1.7;
  color: var(--color-text-secondary);
  white-space: pre-line;
  overflow-wrap: anywhere;
}
.actor-hero__bio-text.is-clamped {
  display: -webkit-box;
  -webkit-line-clamp: 5;
  -webkit-box-orient: vertical;
  overflow: hidden;
}
.actor-hero__bio-empty {
  margin: 0;
  font-size: var(--text-small);
  font-style: italic;
  color: var(--color-text-muted);
}
.actor-hero__bio-toggle {
  display: inline-flex;
  align-items: center;
  gap: 0.25rem;
  margin-top: 0.5rem;
  padding: 0;
  border: 0;
  background: none;
  font-size: var(--text-caption);
  font-weight: 600;
  color: var(--color-text-primary);
  cursor: pointer;
}
.actor-hero__bio-toggle:hover {
  color: var(--color-brand);
}

@media (min-width: 768px) {
  .actor-hero {
    flex-direction: row;
    align-items: flex-start;
    gap: 2rem;
    padding: 1.75rem;
  }
  .actor-hero__photo {
    width: 14rem;
  }
  .actor-hero__info {
    align-items: flex-start;
    text-align: left;
  }
  .actor-hero__name {
    font-size: var(--text-page);
  }
  .actor-hero__facts {
    justify-content: flex-start;
  }
}
@media (min-width: 1280px) {
  .actor-hero__name {
    font-size: var(--text-display);
  }
}
</style>
