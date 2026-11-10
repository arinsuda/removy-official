<template>
  <div
    class="star-rating"
    :class="{ 'star-rating--readonly': readonly }"
    :role="readonly ? 'img' : 'slider'"
    :tabindex="readonly ? undefined : 0"
    :aria-label="ariaLabel"
    :aria-valuemin="readonly ? undefined : STEP"
    :aria-valuemax="readonly ? undefined : max"
    :aria-valuenow="readonly ? undefined : modelValue"
    :aria-valuetext="`${modelValue} / ${max}`"
    @keydown="onKeydown"
  >
    <div
      v-for="n in max"
      :key="n"
      class="star-slot"
      :style="{ width: `${size}px`, height: `${size}px` }"
      @mouseleave="!readonly && (hoverValue = 0)"
    >
      <template v-if="!readonly">
        <div
          class="half-zone left"
          @click="select(n - 0.5)"
          @mouseenter="hoverValue = n - 0.5"
        />
        <div
          class="half-zone right"
          @click="select(n)"
          @mouseenter="hoverValue = n"
        />
      </template>

      <svg
        class="star-base"
        viewBox="0 0 24 24"
        xmlns="http://www.w3.org/2000/svg"
      >
        <path
          :d="STAR_PATH"
          stroke="currentColor"
          stroke-width="1.5"
          stroke-linejoin="round"
          fill="none"
        />
      </svg>
      <svg
        class="star-fill"
        viewBox="0 0 24 24"
        xmlns="http://www.w3.org/2000/svg"
      >
        <defs>
          <clipPath :id="`${uid}-${n}-half`">
            <rect x="0" y="0" width="12" height="24" />
          </clipPath>
          <clipPath :id="`${uid}-${n}-full`">
            <rect x="0" y="0" width="24" height="24" />
          </clipPath>
        </defs>
        <path
          :d="STAR_PATH"
          fill="currentColor"
          :clip-path="clipPathFor(n)"
          :class="n - 0.5 <= displayValue ? 'fill--on' : 'fill--off'"
        />
      </svg>
    </div>

    <span v-if="showLabel" class="rating-label"
      >{{ displayValue }} / {{ max }}</span
    >
  </div>
</template>

<script setup lang="ts">
  import { computed, getCurrentInstance, ref } from "vue"

  const STAR_PATH =
    "M12 2l3.09 6.26L22 9.27l-5 4.87 1.18 6.88L12 17.77l-6.18 3.25L7 14.14 2 9.27l6.91-1.01L12 2z"

  const props = withDefaults(
    defineProps<{
      modelValue: number
      max?: number
      readonly?: boolean
      showLabel?: boolean
      size?: number
      ariaLabel?: string
    }>(),
    { max: 5, readonly: false, showLabel: true, size: 22, ariaLabel: "Rating" },
  )

  const emit = defineEmits<{
    (e: "update:modelValue", value: number): void
  }>()

  // Unique per-instance id so clipPath ids never collide when several
  // StarRatingInput components render on the same page at once.
  const uid = `star-rating-${getCurrentInstance()?.uid ?? Math.random().toString(36).slice(2)}`

  const hoverValue = ref(0)
  const displayValue = computed(() => hoverValue.value || props.modelValue)

  const STEP = 0.5

  function select(value: number) {
    emit("update:modelValue", value)
  }

  function onKeydown(e: KeyboardEvent) {
    if (props.readonly) return
    const keys: Record<string, number> = {
      ArrowRight: Math.min(props.max, props.modelValue + STEP),
      ArrowUp: Math.min(props.max, props.modelValue + STEP),
      ArrowLeft: Math.max(STEP, props.modelValue - STEP),
      ArrowDown: Math.max(STEP, props.modelValue - STEP),
      Home: STEP,
      End: props.max,
    }
    const next = keys[e.key]
    if (next === undefined) return
    e.preventDefault()
    select(next)
  }

  function clipPathFor(n: number) {
    const value = displayValue.value
    if (value >= n) return `url(#${uid}-${n}-full)`
    if (value >= n - 0.5) return `url(#${uid}-${n}-half)`
    return "none"
  }
</script>

<style scoped>
  .star-rating {
    display: flex;
    align-items: center;
    gap: 0.35rem;
    font-size: 0.85rem;
    color: var(--color-text-muted);
  }

  .stars-wrap,
  .star-rating {
    display: flex;
    gap: 3px;
  }

  .star-slot {
    position: relative;
    flex-shrink: 0;
  }
  .star-rating:not(.star-rating--readonly) .star-slot {
    cursor: pointer;
    transition: transform 0.12s ease;
  }
  .star-rating:not(.star-rating--readonly) .star-slot:hover {
    transform: scale(1.18);
  }

  .star-base,
  .star-fill {
    position: absolute;
    inset: 0;
    width: 100%;
    height: 100%;
    pointer-events: none;
  }
  .star-base {
    color: var(--color-border-strong);
  }
  .fill--on {
    color: var(--color-rating);
  }
  .fill--off {
    color: transparent;
  }

  .half-zone {
    position: absolute;
    top: 0;
    height: 100%;
    width: 50%;
    cursor: pointer;
    z-index: 1;
  }
  .half-zone.left {
    left: 0;
  }
  .half-zone.right {
    right: 0;
  }

  .rating-label {
    font-size: 0.85rem;
    color: var(--color-text-muted);
    min-width: 40px;
  }
</style>
