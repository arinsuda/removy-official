<template>
  <ul ref="gridRef" class="grid" :aria-label="t('onboarding.gridLabel')" @keydown="onKeydown">
    <PosterTile
      v-for="(movie, i) in movies"
      :key="movie.id"
      :movie="movie"
      :index="i"
      :selected="selectedIds.has(movie.id)"
      :disabled="disabled"
      @toggle="emit('toggle', movie)"
    />
  </ul>
</template>

<script setup lang="ts">
import { ref } from "vue"
import { useI18n } from "vue-i18n"
import PosterTile from "@/components/onboarding/PosterTile.vue"
import type { Movie } from "@/types"

defineProps<{ movies: Movie[]; selectedIds: Set<number>; disabled?: boolean }>()
const emit = defineEmits<{ toggle: [movie: Movie] }>()
const { t } = useI18n()
const gridRef = ref<HTMLElement | null>(null)

/** Arrow keys move focus across the grid; Tab and Space/Enter keep working as usual. */
function onKeydown(e: KeyboardEvent) {
  const root = gridRef.value
  const key = e.key
  if (!root || !["ArrowLeft", "ArrowRight", "ArrowUp", "ArrowDown"].includes(key)) return
  const buttons = [...root.querySelectorAll<HTMLButtonElement>(".tile__btn")]
  const at = buttons.findIndex((b) => b === document.activeElement)
  if (at < 0) return
  const cols = getComputedStyle(root).gridTemplateColumns.split(" ").length || 1
  const step = key === "ArrowLeft" ? -1 : key === "ArrowRight" ? 1 : key === "ArrowUp" ? -cols : cols
  const next = buttons[at + step]
  if (!next) return
  e.preventDefault()
  next.focus()
}
</script>

<style scoped>
.grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 0.75rem;
  margin: 0;
  padding: 1.25rem 0 2rem;
  list-style: none;
}

@media (min-width: 480px) {
  .grid {
    grid-template-columns: repeat(3, minmax(0, 1fr));
  }
}
@media (min-width: 768px) {
  .grid {
    grid-template-columns: repeat(4, minmax(0, 1fr));
    gap: 1rem;
  }
}
@media (min-width: 1024px) {
  .grid {
    grid-template-columns: repeat(6, minmax(0, 1fr));
  }
}
</style>
