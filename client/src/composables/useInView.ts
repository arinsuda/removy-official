import { onBeforeUnmount, onMounted, ref, type Ref } from "vue"

/**
 * True once `target` has come within `rootMargin` of the viewport (never
 * resets). Without IntersectionObserver it is true immediately.
 */
export function useInView(target: Ref<HTMLElement | null>, rootMargin = "300px") {
  const seen = ref(false)
  let observer: IntersectionObserver | null = null

  onMounted(() => {
    if (typeof IntersectionObserver === "undefined" || !target.value) {
      seen.value = true
      return
    }
    observer = new IntersectionObserver(
      (entries) => {
        if (entries.some((e) => e.isIntersecting)) {
          seen.value = true
          observer?.disconnect()
        }
      },
      { rootMargin },
    )
    observer.observe(target.value)
  })
  onBeforeUnmount(() => observer?.disconnect())

  return seen
}
