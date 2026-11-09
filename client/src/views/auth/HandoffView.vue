<template>
  <AuthShell :title="title">
    <AuthStatus v-if="state === 'working'">
      <AppSpinner :size="32" class="handoff-spinner" />
      <p role="status">{{ t("navigation.link.handoff.working") }}</p>
    </AuthStatus>

    <AuthStatus v-else-if="state === 'failed'" :icon="IconAlert" tone="danger">
      <p role="alert">{{ failedMessage }}</p>
      <div class="auth-status__actions">
        <AppButton to="/login" block>
          {{ t("navigation.link.handoff.backToLogin") }}
        </AppButton>
        <HandoffBackLink :url="planziUrl" />
      </div>
    </AuthStatus>

    <div v-else-if="step === 'code'" class="auth-form">
      <p class="handoff-text handoff-text--ok" role="status">
        {{ t("navigation.link.handoff.codeSent", { email: pending?.email }) }}
      </p>
      <form ref="codeForm" class="auth-form" @submit.prevent="submitCode">
        <AuthField
          v-model="code"
          :label="t('navigation.link.handoff.code')"
          :icon="IconKey"
          inputmode="numeric"
          maxlength="6"
          autocomplete="one-time-code"
        />
        <AppButton type="submit" block :loading="busy" :disabled="busy || code.length !== CODE_LENGTH">
          {{ t("navigation.link.handoff.confirmCode") }}
        </AppButton>
      </form>
      <div class="handoff-actions">
        <button type="button" class="handoff-link" :disabled="busy || cooldown > 0" @click="sendCode">
          {{ cooldown > 0 ? t("navigation.link.handoff.resendIn", { seconds: cooldown }) : t("navigation.link.handoff.resendCode") }}
        </button>
        <button v-if="hasPassword" type="button" class="handoff-link" :disabled="busy" @click="usePassword">
          {{ t("navigation.link.handoff.usePassword") }}
        </button>
      </div>
      <AuthAlert :message="errorMsg" @dismiss="errorMsg = ''" />
      <HandoffBackLink :url="planziUrl" />
    </div>

    <div v-else class="auth-form">
      <p class="handoff-text form-item">
        {{ t("navigation.link.handoff.confirmDesc", { email: pending?.email }) }}
      </p>

      <form v-if="hasPassword" class="auth-form" @submit.prevent="submitPassword">
        <AuthField
          v-model="password"
          type="password"
          :label="t('navigation.link.handoff.password')"
          :icon="IconLock"
          autocomplete="current-password"
        />
        <AppButton type="submit" block :loading="busy" :disabled="busy || !password">
          {{ t("navigation.link.handoff.confirmPassword") }}
        </AppButton>
      </form>

      <p v-if="hasPassword && hasEmailCode" class="handoff-or" aria-hidden="true">
        <span>{{ t("navigation.link.handoff.or") }}</span>
      </p>

      <AppButton v-if="hasEmailCode" :variant="hasPassword ? 'outline' : 'primary'" block :loading="busy && !hasPassword" :disabled="busy" @click="sendCode">
        {{ t("navigation.link.handoff.sendCode") }}
      </AppButton>

      <p v-if="pending && !methods.length" class="handoff-text handoff-text--error" role="alert">
        {{ t("navigation.link.handoff.noMethods") }}
      </p>
      <AuthAlert :message="errorMsg" @dismiss="errorMsg = ''" />
      <HandoffBackLink :url="planziUrl" />
    </div>
  </AuthShell>
</template>

<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, ref } from "vue"
import { useRouter } from "vue-router"
import { useI18n } from "vue-i18n"
import { AlertCircle as IconAlert, KeyRound as IconKey, Lock as IconLock } from "lucide-vue-next"
import AuthAlert from "@/components/auth/AuthAlert.vue"
import AuthField from "@/components/auth/AuthField.vue"
import AuthShell from "@/components/auth/AuthShell.vue"
import AuthStatus from "@/components/auth/AuthStatus.vue"
import AppButton from "@/components/ui/AppButton.vue"
import AppSpinner from "@/components/ui/AppSpinner.vue"
import HandoffBackLink from "@/components/auth/HandoffBackLink.vue"
import { linkApi } from "@/api/endpoints/link"
import type { LinkConfirmRequired, LinkSessionResponse } from "@/api/endpoints/link"
import { useAuthStore } from "@/stores/auth"
import { safeWebUrl } from "@/utils/safeWebUrl"
import { handoffErrorKey } from "@/utils/handoffError"
import { playPlatformEnter } from "@/utils/platformTransition"

/** Wait before the email code can be requested again. */
const RESEND_COOLDOWN_SECONDS = 30
const SECOND_MS = 1000
const CODE_LENGTH = 6

const { t } = useI18n()
const router = useRouter()
const authStore = useAuthStore()

const state = ref<"working" | "failed" | "confirm">("working")
const failedMessage = ref("")
const pending = ref<LinkConfirmRequired | null>(null)
const password = ref("")
const code = ref("")
const step = ref<"methods" | "code">("methods")
const cooldown = ref(0)
const codeForm = ref<HTMLFormElement | null>(null)
let cooldownTimer: number | undefined

const methods = computed(() => pending.value?.methods ?? [])
const hasPassword = computed(() => methods.value.includes("local_password"))
const hasEmailCode = computed(() => methods.value.includes("email_code"))
const busy = ref(false)
const errorMsg = ref("")
const planziUrl = ref("")

const title = computed(() =>
  state.value === "confirm"
    ? t(step.value === "code" ? "navigation.link.handoff.codeTitle" : "navigation.link.handoff.confirmTitle")
    : state.value === "failed"
      ? t("navigation.link.handoff.failedTitle")
      : t("navigation.link.handoff.title"),
)

async function loadPlanziUrl() {
  try {
    planziUrl.value = safeWebUrl((await linkApi.remote()).data.web_url)
  } catch {
    planziUrl.value = "" // linking disabled or unreachable: hide the button
  }
}

function done(data: LinkSessionResponse) {
  authStore.applySession(data)
  void router.replace({ name: "home" }).finally(() => void playPlatformEnter())
}

function readToken(): string {
  const m = /(?:^#|&)t=([^&]+)/.exec(window.location.hash)
  return m && m[1] ? decodeURIComponent(m[1]) : ""
}

function errorFor(err: unknown): string {
  return t(handoffErrorKey(err))
}

onMounted(async () => {
  void loadPlanziUrl()
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
    code.value = ""
    step.value = "code"
    startCooldown()
    await nextTick()
    codeForm.value?.querySelector("input")?.focus()
  } catch (err) {
    errorMsg.value = errorFor(err)
  } finally {
    busy.value = false
  }
}

function startCooldown() {
  window.clearInterval(cooldownTimer)
  cooldown.value = RESEND_COOLDOWN_SECONDS
  cooldownTimer = window.setInterval(() => {
    cooldown.value -= 1
    if (cooldown.value <= 0) window.clearInterval(cooldownTimer)
  }, SECOND_MS)
}

function usePassword() {
  errorMsg.value = ""
  step.value = "methods"
}

onBeforeUnmount(() => window.clearInterval(cooldownTimer))

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
.handoff-text {
  margin: 0;
  font-size: var(--text-small);
  line-height: 1.7;
  color: var(--color-text-secondary);
}

.handoff-text--error {
  color: var(--color-danger);
}

.handoff-text--ok {
  color: var(--color-success);
}

.handoff-actions {
  display: flex;
  flex-wrap: wrap;
  justify-content: center;
  gap: 0.5rem 1.25rem;
}

.handoff-link {
  padding: 0.25rem 0;
  font: inherit;
  font-size: var(--text-small);
  color: var(--color-text-secondary);
  background: none;
  border: 0;
  cursor: pointer;
  text-decoration: underline;
  text-underline-offset: 3px;
}

.handoff-link:hover:not(:disabled) {
  color: var(--color-text-primary);
}

.handoff-link:disabled {
  color: var(--color-text-muted);
  text-decoration: none;
  cursor: default;
}

.handoff-or {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  margin: 0;
  font-size: var(--text-caption);
  color: var(--color-text-muted);
}

.handoff-or::before,
.handoff-or::after {
  content: "";
  flex: 1;
  height: 1px;
  background: var(--color-border);
}

.handoff-spinner {
  color: var(--color-brand);
}
</style>
