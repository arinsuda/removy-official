<template>
  <main class="auth-page">
    <div class="auth-bg" aria-hidden="true" />
    <div ref="cardRef" class="auth-card">
      <aside class="auth-brand">
        <img src="/popcorn-icon-192.png" alt="" class="auth-brand__icon" />
        <p class="auth-brand__name">POP<span>CORN</span></p>
        <p class="auth-brand__tagline">
          Welcome to Popcorn – Your Ultimate Destination
          for Movie Reviews and Recommendations!
        </p>
      </aside>
      <section ref="panelRef" class="auth-panel">
        <h1 class="auth-panel__title form-item">{{ title }}</h1>
        <slot />
      </section>
    </div>
  </main>
</template>

<script setup lang="ts">
import { onMounted, ref } from "vue"
import { gsap } from "gsap"

defineProps<{ title: string }>()

const cardRef = ref<HTMLElement | null>(null)
const panelRef = ref<HTMLElement | null>(null)

const prefersReducedMotion = () =>
  window.matchMedia?.("(prefers-reduced-motion: reduce)").matches ?? false

onMounted(() => {
  if (prefersReducedMotion() || !cardRef.value) return
  const items = cardRef.value.querySelectorAll(".form-item")
  gsap
    .timeline({ defaults: { ease: "power3.out" } })
    .from(cardRef.value, { opacity: 0, y: 24, duration: 0.55, clearProps: "all" })
    // clearProps hands transforms back to CSS so hover and active states work.
    .from(items, { opacity: 0, y: 12, duration: 0.4, stagger: 0.06, clearProps: "all" }, "-=0.3")
})

function shake() {
  gsap.fromTo(
    panelRef.value,
    { x: -8 },
    { x: 0, duration: 0.45, ease: "elastic.out(1, 0.3)", clearProps: "transform" },
  )
}

function playExit() {
  return new Promise<void>(resolve => {
    gsap.to(cardRef.value, {
      opacity: 0,
      y: -20,
      duration: 0.35,
      ease: "power2.in",
      onComplete: resolve,
    })
  })
}

defineExpose({ shake, playExit })
</script>

<style scoped>
.auth-page {
  position: relative;
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 100dvh;
  padding: 2rem;
  box-sizing: border-box;
  background: #0a0a0a;
}

.auth-bg {
  position: fixed;
  inset: 0;
  background:
    radial-gradient(ellipse at center, rgb(10 10 10 / 55%), rgb(10 10 10 / 92%)),
    url("@/assets/bg.png") center / cover no-repeat;
  filter: saturate(0.7);
}

.auth-card {
  position: relative;
  display: grid;
  grid-template-columns: minmax(0, 5fr) minmax(0, 6fr);
  width: 100%;
  max-width: 960px;
  overflow: hidden;
  background: rgb(14 14 14 / 82%);
  border: 1px solid rgb(255 255 255 / 9%);
  border-radius: 24px;
  box-shadow: 0 32px 80px rgb(0 0 0 / 55%);
  backdrop-filter: blur(24px);
}

.auth-brand {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 3rem 2.5rem;
  text-align: center;
  background:
    radial-gradient(circle at 15% 100%, rgb(245 67 44 / 30%), transparent 60%),
    radial-gradient(circle at 90% 5%, rgb(255 217 59 / 12%), transparent 50%);
  border-right: 1px solid rgb(255 255 255 / 7%);
}

.auth-brand__icon {
  width: 88px;
  height: 88px;
  margin-bottom: 1.25rem;
  object-fit: contain;
}

.auth-brand__name {
  margin: 0 0 1rem;
  font-size: clamp(2.25rem, 4vw, 3rem);
  font-weight: 900;
  line-height: 1;
  letter-spacing: -0.02em;
  color: #fff;
}

.auth-brand__name span {
  color: var(--color-brand);
}

.auth-brand__tagline {
  max-width: 280px;
  margin: 0;
  font-size: 0.875rem;
  line-height: 1.7;
  color: var(--color-text-secondary);
}

.auth-panel {
  display: flex;
  flex-direction: column;
  justify-content: center;
  padding: 2.75rem 3rem;
}

.auth-panel__title {
  margin: 0 0 1.75rem;
  font-size: 1.75rem;
  font-weight: 800;
  letter-spacing: 0.08em;
  color: #fff;
}

@media (max-width: 760px) {
  .auth-page {
    align-items: flex-start;
    padding: 1rem;
  }

  .auth-card {
    grid-template-columns: minmax(0, 1fr);
    border-radius: 20px;
  }

  .auth-brand {
    flex-direction: row;
    gap: 0.75rem;
    padding: 1.25rem 1.5rem;
    border-right: 0;
    border-bottom: 1px solid rgb(255 255 255 / 7%);
  }

  .auth-brand__icon {
    width: 40px;
    height: 40px;
    margin: 0;
  }

  .auth-brand__name {
    margin: 0;
    font-size: 1.75rem;
  }

  .auth-brand__tagline {
    display: none;
  }

  .auth-panel {
    padding: 1.75rem 1.5rem;
  }

  .auth-panel__title {
    margin-bottom: 1.25rem;
    font-size: 1.5rem;
  }
}
</style>
