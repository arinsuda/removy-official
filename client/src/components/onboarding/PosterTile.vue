<template>
  <li class="tile" :style="{ '--i': index }">
    <button
      type="button"
      class="tile__btn"
      :class="{ 'tile__btn--on': selected }"
      :aria-pressed="selected"
      :disabled="disabled"
      @click="emit('toggle')"
    >
      <span class="tile__frame">
        <img
          v-if="movie.poster_path && !failed"
          class="tile__img"
          :src="`https://image.tmdb.org/t/p/w342${movie.poster_path}`"
          alt=""
          loading="lazy"
          @error="failed = true"
        />
        <span v-else class="tile__fallback" aria-hidden="true"><Film :size="28" /></span>

        <span v-if="genre" class="tile__genre">{{ genre }}</span>
        <span class="tile__title">{{ movie.title }}</span>

        <span class="tile__check" aria-hidden="true">
          <svg viewBox="0 0 24 24" width="22" height="22" fill="none" stroke="currentColor" stroke-width="3" stroke-linecap="round" stroke-linejoin="round">
            <path d="M5 12.5l4.5 4.5L19 7.5" pathLength="1" />
          </svg>
        </span>
      </span>
    </button>
  </li>
</template>

<script setup lang="ts">
import { computed, ref } from "vue"
import { useI18n } from "vue-i18n"
import { Film } from "lucide-vue-next"
import { genreLabelKey } from "@/components/onboarding/onboardingConfig"
import type { Movie } from "@/types"

const props = defineProps<{ movie: Movie; index: number; selected: boolean; disabled?: boolean }>()
const emit = defineEmits<{ toggle: [] }>()
const { t } = useI18n()

const failed = ref(false)
const genre = computed(() => {
  const key = genreLabelKey(props.movie.genre_ids?.[0])
  return key ? t(key) : ""
})
</script>

<style scoped>
/* The list item owns the entrance; the button owns select and hover motion. */
.tile {
  list-style: none;
  animation: tile-in 520ms var(--ease-out-soft) both;
  animation-delay: calc(var(--i) * 32ms + 120ms);
}

.tile__btn {
  position: relative;
  display: block;
  width: 100%;
  min-height: 44px;
  padding: 0;
  border: 0;
  border-radius: 12px;
  background: none;
  cursor: pointer;
  color: var(--color-text-primary);
  -webkit-tap-highlight-color: transparent;
}
.tile__btn:disabled {
  cursor: default;
}

.tile__btn--on {
  animation: tile-pop 460ms var(--ease-spring);
}

.tile__frame {
  position: relative;
  display: block;
  aspect-ratio: 2 / 3;
  overflow: hidden;
  border-radius: inherit;
  background: var(--color-surface-2);
  box-shadow: inset 0 0 0 1px var(--color-border);
  transition: transform var(--motion-base) var(--ease-out-soft);
}

/* Selection ring and glow sit on a pseudo-element so only opacity animates. */
.tile__btn::after {
  content: "";
  position: absolute;
  inset: 0;
  border-radius: inherit;
  box-shadow:
    0 0 0 2px var(--color-brand),
    0 10px 30px color-mix(in srgb, var(--color-brand) 45%, transparent);
  opacity: 0;
  pointer-events: none;
  transition: opacity var(--motion-base) var(--ease-out-soft);
}
.tile__btn--on::after {
  opacity: 1;
}

.tile__img {
  display: block;
  width: 100%;
  height: 100%;
  object-fit: cover;
  transition:
    transform var(--motion-base) var(--ease-out-soft),
    filter var(--motion-base) var(--ease-out-soft);
}
.tile__btn--on .tile__img {
  transform: scale(1.06);
  filter: brightness(0.7) saturate(1.1);
}

.tile__fallback {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 100%;
  height: 100%;
  color: var(--color-text-muted);
}

.tile__genre {
  position: absolute;
  top: 6px;
  left: 6px;
  padding: 0.15rem 0.5rem;
  border: 1px solid var(--color-border);
  border-radius: 9999px;
  background: color-mix(in srgb, var(--color-surface) 78%, transparent);
  font-size: 0.6875rem;
  font-weight: 700;
  color: var(--color-text-primary);
}

.tile__title {
  position: absolute;
  inset: auto 0 0;
  display: -webkit-box;
  padding: 1.75rem 0.5rem 0.5rem;
  overflow: hidden;
  background: linear-gradient(to top, color-mix(in srgb, var(--color-surface) 92%, transparent), transparent);
  font-size: 0.75rem;
  font-weight: 600;
  line-height: 1.3;
  text-align: left;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
}

.tile__check {
  position: absolute;
  top: 50%;
  left: 50%;
  display: grid;
  place-items: center;
  width: 44px;
  height: 44px;
  margin: -22px 0 0 -22px;
  border-radius: 50%;
  background: var(--color-brand);
  color: var(--color-on-brand);
  opacity: 0;
  transform: scale(0.4);
  transition:
    opacity var(--motion-fast) var(--ease-out-soft),
    transform var(--motion-base) var(--ease-spring);
}
.tile__check path {
  stroke-dasharray: 1;
  stroke-dashoffset: 1;
  transition: stroke-dashoffset 360ms var(--ease-out-soft) 120ms;
}
.tile__btn--on .tile__check {
  opacity: 1;
  transform: scale(1);
}
.tile__btn--on .tile__check path {
  stroke-dashoffset: 0;
}

@media (hover: hover) {
  .tile__btn:not(:disabled):hover .tile__frame {
    transform: translateY(-4px);
  }
}
.tile__btn:not(:disabled):active .tile__frame {
  transform: scale(0.97);
}

@keyframes tile-in {
  from {
    opacity: 0;
    transform: translateY(24px) scale(0.94);
  }
}

@keyframes tile-pop {
  0% {
    transform: scale(0.94);
  }
  55% {
    transform: scale(1.04);
  }
  100% {
    transform: scale(1);
  }
}

@media (prefers-reduced-motion: reduce) {
  .tile,
  .tile__btn--on {
    animation: none;
  }
  .tile__frame,
  .tile__img,
  .tile__check,
  .tile__check path,
  .tile__btn::after {
    transition-duration: 1ms;
    transition-delay: 0ms;
  }
  .tile__btn:hover .tile__frame,
  .tile__btn:active .tile__frame {
    transform: none;
  }
}
</style>
