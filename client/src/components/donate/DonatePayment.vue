<template>
  <section class="pay" aria-labelledby="donate-pay-title">
    <h2 id="donate-pay-title" class="pay__title">{{ t("donate.pay.title") }}</h2>
    <AppTabs v-model="method" :tabs="tabs" :label="t('donate.pay.title')" :controls="false" />

    <p v-if="amount" class="pay__amount" role="status">
      {{ t("donate.pay.suggested", { amount: amount.toLocaleString(locale) }) }}
    </p>

    <AppCard v-if="method === 'promptpay'" class="pay__card">
      <div class="pay__qr">
        <img v-if="!qrError" src="/mePromptPay.png" :alt="t('donate.promptpay.alt')" class="pay__qr-img" @error="qrError = true" />
        <div v-else class="pay__fallback">
          <QrCode :size="48" aria-hidden="true" />
          <p>{{ t("donate.promptpay.fallback") }}</p>
        </div>
      </div>
      <div class="pay__meta">
        <p class="pay__name">{{ t("donate.promptpay.name") }}</p>
        <p class="pay__hint">{{ t("donate.promptpay.hint") }}</p>
      </div>
    </AppCard>

    <AppCard v-else class="pay__card">
      <div class="pay__bank">
        <span class="pay__logo">
          <img v-if="!logoError" src="/kbank-logo.png" :alt="t('donate.bank.name')" @error="logoError = true" />
          <Landmark v-else :size="32" aria-hidden="true" />
        </span>
        <div class="pay__meta">
          <p class="pay__name">{{ t("donate.bank.name") }}</p>
          <p class="pay__hint">{{ t("donate.bank.accountName") }}</p>
        </div>
      </div>
      <div class="pay__account">
        <span class="pay__number" :aria-label="ACCOUNT_NUMBER.split('').join(' ')">{{ formatted }}</span>
        <AppButton variant="outline" size="sm" @click="copy">
          <Check v-if="copied" :size="16" aria-hidden="true" />
          <Copy v-else :size="16" aria-hidden="true" />
          {{ copied ? t("donate.bank.copied") : t("donate.bank.copy") }}
        </AppButton>
        <span class="sr-only" role="status">{{ copied ? t("donate.bank.copied") : "" }}</span>
      </div>
    </AppCard>
  </section>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, ref } from "vue"
import { useI18n } from "vue-i18n"
import { Check, Copy, Landmark, QrCode } from "lucide-vue-next"
import AppButton from "@/components/ui/AppButton.vue"
import AppCard from "@/components/ui/AppCard.vue"
import AppTabs from "@/components/ui/AppTabs.vue"

defineProps<{ amount: number | null }>()

const ACCOUNT_NUMBER = "0453315436"
type Method = "promptpay" | "bank"

const { t, locale } = useI18n()
const method = ref<Method>("promptpay")
const qrError = ref(false)
const logoError = ref(false)
const copied = ref(false)
let timer: ReturnType<typeof setTimeout> | undefined

const tabs = computed(() => [
  { key: "promptpay" as const, label: t("donate.promptpay.name") },
  { key: "bank" as const, label: t("donate.bank.tab") },
])
const formatted = `${ACCOUNT_NUMBER.slice(0, 3)}-${ACCOUNT_NUMBER.slice(3, 4)}-${ACCOUNT_NUMBER.slice(4, 9)}-${ACCOUNT_NUMBER.slice(9)}`

async function copy() {
  try {
    await navigator.clipboard.writeText(ACCOUNT_NUMBER)
    copied.value = true
    clearTimeout(timer)
    timer = setTimeout(() => (copied.value = false), 2000)
  } catch {
    copied.value = false
  }
}
onBeforeUnmount(() => clearTimeout(timer))
</script>

<style scoped>
.pay {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}
.pay__title {
  margin: 0;
  font-size: var(--text-heading);
  font-weight: 700;
  color: var(--color-text-primary);
}
.pay__amount {
  margin: 0;
  font-size: var(--text-small);
  color: var(--color-accent);
}
.pay__card :deep(.app-card__body) {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 1.25rem;
  padding: 1.5rem;
}
.pay__qr {
  display: grid;
  place-items: center;
  width: min(100%, 16rem);
  aspect-ratio: 1;
  padding: 0.75rem;
  border-radius: var(--radius-card);
  background: var(--color-text-primary);
}
.pay__qr-img {
  width: 100%;
  height: 100%;
  object-fit: contain;
}
.pay__fallback {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0.5rem;
  color: var(--color-surface);
  font-size: var(--text-small);
  text-align: center;
}
.pay__meta {
  text-align: center;
}
.pay__name {
  margin: 0;
  font-weight: 700;
  color: var(--color-text-primary);
}
.pay__hint {
  margin: 0.25rem 0 0;
  font-size: var(--text-small);
  color: var(--color-text-secondary);
}
.pay__bank {
  display: flex;
  align-items: center;
  gap: 1rem;
}
.pay__bank .pay__meta {
  text-align: start;
}
.pay__logo {
  display: grid;
  place-items: center;
  width: 3.5rem;
  height: 3.5rem;
  overflow: hidden;
  border-radius: var(--radius-card);
  background: var(--color-text-primary);
  color: var(--color-surface);
}
.pay__logo img {
  width: 100%;
  height: 100%;
  object-fit: contain;
}
.pay__account {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: center;
  gap: 0.75rem;
  width: 100%;
  padding: 0.75rem 1rem;
  border: 1px solid var(--color-border);
  border-radius: var(--radius-card);
  background: var(--color-surface);
}
.pay__number {
  font-size: var(--text-title);
  font-weight: 700;
  letter-spacing: 0.06em;
  font-variant-numeric: tabular-nums;
  color: var(--color-text-primary);
}
.sr-only {
  position: absolute;
  width: 1px;
  height: 1px;
  overflow: hidden;
  clip-path: inset(50%);
  white-space: nowrap;
}
</style>
