<template>
  <main class="auth-page">
    <aside class="auth-brand">
      <div class="auth-brand__inner">
        <RouterLink to="/" class="auth-brand__logo" :aria-label="t('auth.shell.home')">
          <img src="/popcorn-icon-192.png" alt="" class="auth-brand__icon" />
          <!-- hardcode-ok: brand name -->
          <span class="auth-brand__name">POP<span>CORN</span></span>
        </RouterLink>
        <div class="auth-brand__copy">
          <p class="auth-brand__tagline">{{ t("auth.shell.tagline") }}</p>
          <ul class="auth-brand__features">
            <li v-for="f in FEATURES" :key="f.key">
              <component :is="f.icon" :size="16" aria-hidden="true" />
              {{ t(`auth.shell.features.${f.key}`) }}
            </li>
          </ul>
        </div>
      </div>
    </aside>

    <div ref="cardRef" class="auth-main">
      <section ref="panelRef" class="auth-panel">
        <h1 class="auth-panel__title form-item">{{ title }}</h1>
        <p v-if="subtitle" class="auth-panel__subtitle form-item">{{ subtitle }}</p>
        <slot />
      </section>
    </div>
  </main>
</template>

<script setup lang="ts">
import { onMounted, ref } from "vue"
import { useI18n } from "vue-i18n"
import { gsap } from "gsap"
import { Popcorn, Star, Users } from "lucide-vue-next"

defineProps<{ title: string; subtitle?: string }>()

const { t } = useI18n()
const FEATURES = [
  { key: "rate", icon: Star },
  { key: "track", icon: Popcorn },
  { key: "follow", icon: Users },
]

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
  if (prefersReducedMotion()) return
  gsap.fromTo(
    panelRef.value,
    { x: -8 },
    { x: 0, duration: 0.45, ease: "elastic.out(1, 0.3)", clearProps: "transform" },
  )
}

function playExit() {
  if (prefersReducedMotion()) return Promise.resolve()
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
  display: grid;
  grid-template-columns: minmax(0, 1fr);
  grid-template-rows: auto 1fr;
  min-height: 100dvh;
  background: var(--color-surface);
}

/* Poster collage behind a scrim; a short band on phones, a full column from 900px. */
.auth-brand {
  position: relative;
  display: flex;
  align-items: flex-end;
  min-height: 9rem;
  overflow: hidden;
  background:
    linear-gradient(to top, var(--color-surface), color-mix(in srgb, var(--color-surface) 55%, transparent)),
    url("@/assets/bg.png") center / cover no-repeat;
  border-bottom: 1px solid var(--color-border);
}

.auth-brand__inner {
  position: relative;
  display: flex;
  flex-direction: column;
  gap: 2rem;
  width: 100%;
  padding: 1.25rem 1.25rem 1rem;
  box-sizing: border-box;
}

.auth-brand__logo {
  display: inline-flex;
  align-items: center;
  gap: 0.625rem;
  width: fit-content;
  color: var(--color-text-primary);
  text-decoration: none;
}

.auth-brand__icon {
  width: 2.5rem;
  height: 2.5rem;
  object-fit: contain;
}

.auth-brand__name {
  font-size: 1.75rem;
  font-weight: 900;
  line-height: 1;
  letter-spacing: -0.02em;
}

.auth-brand__name span {
  color: var(--color-brand);
}

.auth-brand__copy {
  display: none;
}

.auth-main {
  display: flex;
  align-items: flex-start;
  justify-content: center;
  padding: 1.5rem 1.25rem 3rem;
}

.auth-panel {
  width: 100%;
  max-width: 26rem;
}

.auth-panel__title {
  margin: 0 0 0.375rem;
  font-size: var(--text-heading);
  font-weight: 800;
  line-height: 1.3;
  color: var(--color-text-primary);
}

.auth-panel__subtitle {
  margin: 0 0 1.5rem;
  font-size: var(--text-small);
  color: var(--color-text-secondary);
}

/* Shared form furniture for every auth view rendered inside the panel. */
.auth-panel :slotted(.auth-form) {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.auth-panel :slotted(.auth-footer) {
  display: flex;
  flex-wrap: wrap;
  justify-content: center;
  gap: 0.375rem;
  margin: 0;
  font-size: var(--text-small);
  color: var(--color-text-muted);
}

.auth-panel :slotted(.auth-link) {
  font-size: var(--text-small);
  color: var(--color-text-secondary);
  text-decoration: underline;
  text-underline-offset: 3px;
  background: none;
  border: 0;
  padding: 0;
  cursor: pointer;
}

@media (hover: hover) {
  .auth-panel :slotted(.auth-link:hover) {
    color: var(--color-text-primary);
  }
}

@media (min-width: 900px) {
  .auth-page {
    grid-template-columns: minmax(0, 5fr) minmax(0, 6fr);
    grid-template-rows: minmax(0, 1fr);
  }

  .auth-brand {
    position: sticky;
    top: 0;
    height: 100dvh;
    border-right: 1px solid var(--color-border);
    border-bottom: 0;
    background:
      linear-gradient(to top, var(--color-surface) 8%, color-mix(in srgb, var(--color-surface) 35%, transparent) 70%),
      radial-gradient(circle at 10% 100%, color-mix(in srgb, var(--color-brand) 35%, transparent), transparent 60%),
      url("@/assets/bg.png") center / cover no-repeat;
  }

  .auth-brand__inner {
    height: 100%;
    justify-content: space-between;
    padding: 2.5rem;
  }

  .auth-brand__icon {
    width: 3rem;
    height: 3rem;
  }

  .auth-brand__name {
    font-size: 2rem;
  }

  .auth-brand__copy {
    display: block;
  }

  .auth-brand__tagline {
    max-width: 24rem;
    margin: 0 0 1.25rem;
    font-size: var(--text-title);
    font-weight: 700;
    line-height: 1.45;
    color: var(--color-text-primary);
  }

  .auth-brand__features {
    display: flex;
    flex-direction: column;
    gap: 0.625rem;
    margin: 0;
    padding: 0;
    list-style: none;
    font-size: var(--text-small);
    color: var(--color-text-secondary);
  }

  .auth-brand__features li {
    display: flex;
    align-items: center;
    gap: 0.625rem;
  }

  .auth-brand__features svg {
    color: var(--color-brand);
  }

  .auth-main {
    align-items: center;
    padding: 3rem 2rem;
  }

  .auth-panel__title {
    font-size: var(--text-page);
  }
}
</style>
