<template>
  <div class="handoff-page">
    <div class="handoff-card">
      <div class="brand"><img src="/popcorn-full-512.png" alt="POPCORN" class="brand-logo" /></div>

      <template v-if="state === 'working'">
        <p class="text">{{ t("navigation.link.handoff.working") }}</p>
      </template>

      <template v-else-if="state === 'failed'">
        <p class="text text--error">{{ failedMessage }}</p>
        <AppButton to="/login" block>
          {{ t("navigation.link.handoff.backToLogin") }}
        </AppButton>
      </template>

      <template v-else>
        <h1 class="title">{{ t("navigation.link.handoff.confirmTitle") }}</h1>
        <p class="text">
          {{ t("navigation.link.handoff.confirmDesc", { email: pending?.email }) }}
        </p>

        <form
          v-if="pending?.methods.includes('local_password')"
          class="block"
          @submit.prevent="submitPassword"
        >
          <input
            v-model="password"
            type="password"
            class="input"
            autocomplete="current-password"
            :placeholder="t('navigation.link.handoff.password')"
          />
          <AppButton type="submit" block :disabled="busy || !password">
            {{ t("navigation.link.handoff.confirmPassword") }}
          </AppButton>
        </form>

        <div v-if="pending?.methods.includes('email_code')" class="block">
          <AppButton variant="outline" block :disabled="busy" @click="sendCode">
            {{ t("navigation.link.handoff.sendCode") }}
          </AppButton>
          <template v-if="codeSent">
            <p class="text text--ok">{{ t("navigation.link.handoff.codeSent") }}</p>
            <form class="block" @submit.prevent="submitCode">
              <input
                v-model="code"
                inputmode="numeric"
                maxlength="6"
                class="input"
                autocomplete="one-time-code"
                :placeholder="t('navigation.link.handoff.code')"
              />
              <AppButton type="submit" block :disabled="busy || code.length !== 6">
                {{ t("navigation.link.handoff.confirmCode") }}
              </AppButton>
            </form>
          </template>
        </div>

        <p v-if="errorMsg" class="text text--error">{{ errorMsg }}</p>
      </template>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from "vue"
import { useRouter } from "vue-router"
import { useI18n } from "vue-i18n"
import AppButton from "@/components/ui/AppButton.vue"
import { linkApi } from "@/api/endpoints/link"
import type { LinkConfirmRequired, LinkSessionResponse } from "@/api/endpoints/link"
import { useAuthStore } from "@/stores/auth"
import { playPlatformEnter } from "@/utils/platformTransition"

const { t } = useI18n()
const router = useRouter()
const authStore = useAuthStore()

const state = ref<"working" | "failed" | "confirm">("working")
const failedMessage = ref("")
const pending = ref<LinkConfirmRequired | null>(null)
const password = ref("")
const code = ref("")
const codeSent = ref(false)
const busy = ref(false)
const errorMsg = ref("")

function done(data: LinkSessionResponse) {
  authStore.applySession(data)
  void router.replace({ name: "home" }).finally(() => void playPlatformEnter())
}

function readToken(): string {
  const m = /(?:^#|&)t=([^&]+)/.exec(window.location.hash)
  return m && m[1] ? decodeURIComponent(m[1]) : ""
}

function errorFor(err: unknown): string {
  const status = (err as { response?: { status?: number } })?.response?.status
  return status === 429
    ? t("navigation.link.handoff.tooMany")
    : t("navigation.link.handoff.wrong")
}

onMounted(async () => {
  const token = readToken()
  // never leave the one-time token in the address bar / history
  history.replaceState(history.state, "", window.location.pathname)
  if (!token) {
    failedMessage.value = t("navigation.link.handoff.failed")
    state.value = "failed"
    void playPlatformEnter()
    return
  }
  // Reveal this page if redeeming is slow; otherwise the destination slides in.
  const slowReveal = window.setTimeout(() => void playPlatformEnter(), 800)
  try {
    const res = await linkApi.redeem(token)
    window.clearTimeout(slowReveal)
    done(res.data)
  } catch (err: unknown) {
    window.clearTimeout(slowReveal)
    const resp = (err as { response?: { status?: number; data?: LinkConfirmRequired } }).response
    if (resp?.status === 409 && resp.data?.code === "LINK_CONFIRM_REQUIRED") {
      pending.value = resp.data
      state.value = "confirm"
    } else {
      const code = (resp?.data as { code?: string } | undefined)?.code
      failedMessage.value =
        code === "REMOTE_EMAIL_UNVERIFIED"
          ? t("navigation.link.handoff.remoteUnverified")
          : t("navigation.link.handoff.failed")
      state.value = "failed"
    }
    void playPlatformEnter()
  }
})

async function submitPassword() {
  if (!pending.value) return
  busy.value = true
  errorMsg.value = ""
  try {
    const res = await linkApi.confirm({
      pending_id: pending.value.pending_id,
      method: "local_password",
      password: password.value,
    })
    done(res.data)
  } catch (err) {
    errorMsg.value = errorFor(err)
  } finally {
    busy.value = false
  }
}

async function sendCode() {
  if (!pending.value) return
  busy.value = true
  errorMsg.value = ""
  try {
    await linkApi.sendCode(pending.value.pending_id)
    codeSent.value = true
  } catch (err) {
    errorMsg.value = errorFor(err)
  } finally {
    busy.value = false
  }
}

async function submitCode() {
  if (!pending.value) return
  busy.value = true
  errorMsg.value = ""
  try {
    const res = await linkApi.confirm({
      pending_id: pending.value.pending_id,
      method: "email_code",
      code: code.value,
    })
    done(res.data)
  } catch (err) {
    errorMsg.value = errorFor(err)
  } finally {
    busy.value = false
  }
}
</script>

<style scoped>
.handoff-page {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 24px 16px;
  background: #050505;
}
.handoff-card {
  width: 100%;
  max-width: 420px;
  background: #0f0f0f;
  border: 1px solid #1c1c1c;
  border-radius: 6px;
  padding: 36px 28px;
  text-align: center;
  color: #fff;
}
.brand {
  font-size: 18px;
  font-weight: 700;
  letter-spacing: 4px;
  margin-bottom: 24px;
}
.brand-logo {
  display: block;
  width: 140px;
  margin: 0 auto;
}
.brand-re {
  color: var(--color-brand);
}
.title {
  font-size: 18px;
  font-weight: 600;
  margin: 0 0 12px;
}
.text {
  color: #8f8f8f;
  font-size: 14px;
  line-height: 1.7;
  margin: 0 0 16px;
}
.text--error {
  color: #ff6b6b;
}
.text--ok {
  color: #4cd98a;
}
.block {
  display: flex;
  flex-direction: column;
  gap: 10px;
  margin-bottom: 16px;
}
.input {
  width: 100%;
  box-sizing: border-box;
  padding: 12px 14px;
  background: #151515;
  border: 1px solid #2a2a2a;
  border-radius: 6px;
  color: #fff;
  font-size: 14px;
}
</style>
