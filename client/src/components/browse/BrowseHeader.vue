<template>
  <header class="browse-header">
    <h1 class="browse-title">{{ t(`browse.title.${type}`) }}</h1>
    <div class="browse-header__row">
      <div class="type-switch" role="group" :aria-label="t('browse.type.label')">
        <AppButton
          v-for="option in TYPES"
          :key="option"
          size="sm"
          :variant="option === type ? 'primary' : 'ghost'"
          :aria-pressed="option === type"
          @click="emit('switch-type', option)"
        >
          {{ t(`browse.type.${option}`) }}
        </AppButton>
      </div>
      <div class="search">
        <label class="search__label" :for="inputId">{{ t("browse.search.label") }}</label>
        <Search :size="16" class="search__icon" aria-hidden="true" />
        <input
          :id="inputId"
          v-model="text"
          class="search__input"
          type="search"
          autocomplete="off"
          :placeholder="t(`browse.search.placeholder.${type}`)"
          :aria-describedby="short ? hintId : undefined"
          @keydown.enter="flush"
        />
        <button v-if="text" type="button" class="search__clear" :aria-label="t('browse.search.clear')" @click="clear">
          <X :size="14" aria-hidden="true" />
        </button>
        <p v-if="short" :id="hintId" class="search__hint">{{ t("browse.search.hint", { n: MIN_SEARCH_CHARS }) }}</p>
      </div>
    </div>
  </header>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, ref, useId, watch } from "vue"
import { useI18n } from "vue-i18n"
import { Search, X } from "lucide-vue-next"
import AppButton from "@/components/ui/AppButton.vue"
import type { HomeMediaType } from "@/components/home/homeMedia"
import { MIN_SEARCH_CHARS } from "./browseConfig"

const props = defineProps<{ type: HomeMediaType; query: string }>()
const emit = defineEmits<{ "switch-type": [type: HomeMediaType]; search: [q: string] }>()

const TYPES: HomeMediaType[] = ["movie", "tv"]
const DEBOUNCE_MS = 350

const { t } = useI18n()
const inputId = useId()
const hintId = useId()
const text = ref(props.query)
let timer: ReturnType<typeof setTimeout> | null = null

const short = computed(() => text.value.trim().length > 0 && text.value.trim().length < MIN_SEARCH_CHARS)

// Back/forward or a reset elsewhere changes the URL; mirror it into the box.
watch(
  () => props.query,
  (q) => {
    if (q !== text.value.trim()) text.value = q
  },
)

function flush() {
  if (timer) clearTimeout(timer)
  timer = null
  const q = text.value.trim()
  // One character is not searched; emit only empty or searchable text.
  if (q === "" || q.length >= MIN_SEARCH_CHARS) {
    if (q !== props.query) emit("search", q)
  }
}

watch(text, () => {
  if (timer) clearTimeout(timer)
  timer = setTimeout(flush, DEBOUNCE_MS)
})

function clear() {
  text.value = ""
  flush()
}

onBeforeUnmount(() => timer && clearTimeout(timer))
</script>

<style scoped>
.browse-header {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}
.browse-title {
  margin: 0;
  font-size: var(--text-title);
  font-weight: 700;
  color: var(--color-text-primary);
}
.browse-header__row {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}
.type-switch {
  display: flex;
  gap: 0.25rem;
}
.search {
  position: relative;
  flex: 1;
  min-width: 0;
}
.search__label {
  position: absolute;
  width: 1px;
  height: 1px;
  overflow: hidden;
  clip-path: inset(50%);
  white-space: nowrap;
}
.search__icon {
  position: absolute;
  top: 0.75rem;
  left: 0.75rem;
  color: var(--color-text-muted);
  pointer-events: none;
}
.search__input {
  width: 100%;
  height: 2.5rem;
  padding: 0 2.25rem;
  border-radius: var(--radius-btn);
  border: 1px solid var(--color-border-strong);
  background: var(--color-surface-2);
  color: var(--color-text-primary);
  font: inherit;
  font-size: var(--text-small);
}
.search__input::placeholder {
  color: var(--color-text-muted);
}
.search__clear {
  position: absolute;
  top: 0.4rem;
  right: 0.4rem;
  display: grid;
  place-items: center;
  width: 1.75rem;
  height: 1.75rem;
  border: 0;
  border-radius: 9999px;
  background: transparent;
  color: var(--color-text-secondary);
  cursor: pointer;
}
.search__clear:hover {
  background: var(--color-surface-3);
}
.search__hint {
  margin: 0.25rem 0 0;
  font-size: var(--text-caption);
  color: var(--color-text-muted);
}
@media (min-width: 640px) {
  .browse-header__row {
    flex-direction: row;
    align-items: flex-start;
    justify-content: space-between;
  }
  .search {
    max-width: 22rem;
  }
}
@media (min-width: 768px) {
  .browse-title {
    font-size: 1.75rem;
  }
}
</style>
