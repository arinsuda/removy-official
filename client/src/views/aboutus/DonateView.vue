<template>
  <main class="donate">
    <section class="hero">
      <span class="hero__icon" aria-hidden="true"><Popcorn :size="36" /></span>
      <p class="hero__eyebrow">{{ t("donate.eyebrow") }}</p>
      <h1 class="hero__title">{{ t("donate.title") }}</h1>
      <p class="hero__subtitle">{{ t("donate.subtitle") }}</p>
      <AppButton size="lg" class="hero__cta" @click="scrollToPay">{{ t("donate.cta") }}</AppButton>
    </section>

    <section class="why" aria-labelledby="donate-why-title">
      <h2 id="donate-why-title" class="section-title">{{ t("donate.impact.title") }}</h2>
      <ul class="why__grid">
        <li v-for="item in IMPACT" :key="item.key" class="why__card">
          <span class="why__icon" aria-hidden="true"><component :is="item.icon" :size="22" /></span>
          <h3 class="why__name">{{ t(`donate.impact.${item.key}.title`) }}</h3>
          <p class="why__desc">{{ t(`donate.impact.${item.key}.desc`) }}</p>
        </li>
      </ul>
    </section>

    <DonateTiers v-model="amount" />
    <div ref="pay"><DonatePayment :amount="amount" /></div>
    <DonateFaq />

    <footer class="thanks">
      <Heart :size="28" class="thanks__icon" aria-hidden="true" />
      <h2 class="thanks__title">{{ t("donate.thanks.title") }}</h2>
      <p class="thanks__text">{{ t("donate.thanks.subtitle") }}</p>
      <AppButton variant="outline" @click="share">
        <Share2 :size="16" aria-hidden="true" />
        {{ t("donate.share.button") }}
      </AppButton>
    </footer>
  </main>
</template>

<script setup lang="ts">
import { ref, useTemplateRef } from "vue"
import { useI18n } from "vue-i18n"
import { Heart, Popcorn, Server, Share2, ShieldCheck, Zap } from "lucide-vue-next"
import DonateFaq from "@/components/donate/DonateFaq.vue"
import DonatePayment from "@/components/donate/DonatePayment.vue"
import DonateTiers from "@/components/donate/DonateTiers.vue"
import AppButton from "@/components/ui/AppButton.vue"
import { useUIStore } from "@/stores/ui"

const IMPACT = [
  { key: "server", icon: Server },
  { key: "features", icon: Zap },
  { key: "maintenance", icon: ShieldCheck },
] as const

const { t } = useI18n()
const ui = useUIStore()
const amount = ref<number | null>(null)
const pay = useTemplateRef<HTMLElement>("pay")

function scrollToPay() {
  const reduce = window.matchMedia?.("(prefers-reduced-motion: reduce)").matches
  pay.value?.scrollIntoView({ behavior: reduce ? "auto" : "smooth", block: "start" })
}

/** Native share sheet where available, otherwise copy the link. */
async function share() {
  const url = `${window.location.origin}/donate`
  try {
    if (navigator.share) {
      await navigator.share({ title: t("donate.title"), text: t("donate.subtitle"), url })
      return
    }
    await navigator.clipboard.writeText(url)
    ui.toastSuccess(t("donate.share.copied"))
  } catch {
    // The person closed the share sheet; nothing to report.
  }
}
</script>

<style scoped>
.donate {
  display: flex;
  flex-direction: column;
  gap: 3rem;
  width: 100%;
  max-width: 56rem;
  margin: 0 auto;
  padding: 1.5rem 1rem 4rem;
  box-sizing: border-box;
}
.section-title {
  margin: 0 0 1rem;
  font-size: var(--text-heading);
  font-weight: 700;
  color: var(--color-text-primary);
}

.hero {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0.75rem;
  padding: 2.5rem 1.25rem;
  border: 1px solid var(--color-border);
  border-radius: var(--radius-card);
  background:
    radial-gradient(circle at 50% 0%, color-mix(in srgb, var(--color-brand) 28%, transparent), transparent 65%),
    var(--color-surface-2);
  text-align: center;
}
.hero__icon {
  display: grid;
  place-items: center;
  width: 4.5rem;
  height: 4.5rem;
  border-radius: 50%;
  background: color-mix(in srgb, var(--color-brand) 18%, transparent);
  color: var(--color-accent);
}
.hero__eyebrow {
  margin: 0;
  font-size: var(--text-caption);
  font-weight: 700;
  letter-spacing: 0.12em;
  text-transform: uppercase;
  color: var(--color-text-muted);
}
.hero__title {
  margin: 0;
  font-size: var(--text-heading);
  font-weight: 800;
  line-height: 1.25;
  color: var(--color-text-primary);
}
.hero__subtitle {
  max-width: 36rem;
  margin: 0;
  color: var(--color-text-secondary);
}
.hero__cta {
  margin-top: 0.75rem;
}

.why__grid {
  display: grid;
  grid-template-columns: minmax(0, 1fr);
  gap: 0.75rem;
  margin: 0;
  padding: 0;
  list-style: none;
}
.why__card {
  padding: 1.25rem;
  border: 1px solid var(--color-border);
  border-radius: var(--radius-card);
  background: var(--color-surface-2);
}
.why__icon {
  display: grid;
  place-items: center;
  width: 2.5rem;
  height: 2.5rem;
  margin-bottom: 0.75rem;
  border-radius: var(--radius-card);
  background: var(--color-surface-3);
  color: var(--color-info);
}
.why__name {
  margin: 0 0 0.25rem;
  font-size: var(--text-body);
  font-weight: 700;
  color: var(--color-text-primary);
}
.why__desc {
  margin: 0;
  font-size: var(--text-small);
  color: var(--color-text-secondary);
}

.thanks {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0.5rem;
  padding: 2rem 1rem 0;
  text-align: center;
}
.thanks__icon {
  color: var(--color-brand);
}
.thanks__title {
  margin: 0;
  font-size: var(--text-title);
  font-weight: 700;
  color: var(--color-text-primary);
}
.thanks__text {
  max-width: 30rem;
  margin: 0 0 0.75rem;
  color: var(--color-text-secondary);
}

@media (min-width: 640px) {
  .why__grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
  .why__card:last-child:nth-child(odd) {
    grid-column: 1 / -1;
  }
  .hero__title {
    font-size: var(--text-page);
  }
}
@media (min-width: 1024px) {
  .why__grid {
    grid-template-columns: repeat(3, minmax(0, 1fr));
  }
  .why__card:last-child:nth-child(odd) {
    grid-column: auto;
  }
}
@media (min-width: 768px) {
  .donate {
    padding: 2.5rem 1.5rem 5rem;
  }
  .hero {
    padding: 3.5rem 2rem;
  }
  .hero__title {
    font-size: var(--text-display);
  }
}
</style>
