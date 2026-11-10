<template>
  <AuthShell :title="title">
    <div v-if="status === 'form'" class="verify-form">
      <form v-if="!confirmingSkip" class="auth-form" @submit.prevent="submit({ password })">
        <p class="verify-text form-item">{{ t("auth.verifyEmail.passwordPrompt") }}</p>
        <AuthAlert :message="wrongPassword ? t('auth.verifyEmail.wrongPassword') : ''" @dismiss="wrongPassword = false" />
        <AuthField
          v-model="password"
          :label="t('auth.verifyEmail.passwordLabel')"
          type="password"
          :icon="IconLock"
          autocomplete="current-password"
        />
        <AppButton type="submit" size="lg" block class="form-item" :disabled="!password">
          {{ t("auth.verifyEmail.confirm") }}
        </AppButton>
        <button type="button" class="auth-link form-item" @click="confirmingSkip = true">
          {{ t("auth.verifyEmail.skipPassword") }}
        </button>
      </form>
      <div v-else class="auth-form" role="alertdialog" :aria-label="t('auth.verifyEmail.skipPassword')">
        <AuthAlert :message="t('auth.verifyEmail.skipWarning')" @dismiss="confirmingSkip = false" />
        <AppButton variant="danger" size="lg" block @click="submit({ skipPassword: true })">
          {{ t("auth.verifyEmail.skipConfirm") }}
        </AppButton>
        <button type="button" class="auth-link" @click="confirmingSkip = false">
          {{ t("auth.verifyEmail.skipCancel") }}
        </button>
      </div>
    </div>

    <AuthStatus v-else-if="status === 'loading'">
      <AppSpinner :size="32" class="verify-spinner" />
      <p>{{ t("auth.verifyEmail.verifying") }}</p>
    </AuthStatus>

    <AuthStatus v-else-if="status === 'success'" :icon="IconCircleCheck" tone="success">
      <p v-if="passwordResetRequired">{{ t("auth.verifyEmail.resetRequired") }}</p>
      <p v-else>{{ t("auth.verifyEmail.successDesc") }}</p>
      <div class="auth-status__actions">
        <AppButton v-if="passwordResetRequired" to="/forgot-password" block>
          {{ t("auth.verifyEmail.resetCta") }}
        </AppButton>
        <AppButton v-else to="/login" block>{{ t("auth.verifyEmail.loginNow") }}</AppButton>
      </div>
    </AuthStatus>

    <AuthStatus v-else :icon="IconAlertCircle" tone="danger">
      <p v-if="rateLimited" role="alert">{{ t("errors.rateLimited") }}</p>
      <p v-else-if="attemptsExceeded">{{ t("auth.verifyEmail.attemptsExceeded") }}</p>
      <p v-else>{{ t("auth.verifyEmail.expiredDesc") }} {{ t("auth.verifyEmail.expiredHint") }}</p>
      <div class="auth-status__actions">
        <AppButton
          block
          :loading="resendStatus === 'sending'"
          :disabled="resendStatus === 'sending' || resendStatus === 'sent'"
          @click="resend"
        >
          {{ resendLabel }}
        </AppButton>
        <AppButton to="/login" variant="ghost" block>{{ t("auth.verifyEmail.backToLogin") }}</AppButton>
      </div>
    </AuthStatus>

    <ResendEmailModal
      :open="promptOpen"
      :title="t('auth.verifyEmail.resendLink')"
      :label="t('auth.verifyEmail.promptEmail')"
      @close="promptOpen = false"
      @submit="confirmEmail"
    />
  </AuthShell>
</template>

<script setup lang="ts">
import AppSpinner from "@/components/ui/AppSpinner.vue"
  import AppButton from "@/components/ui/AppButton.vue"
  import ResendEmailModal from "@/components/auth/ResendEmailModal.vue"
  import AuthAlert from "@/components/auth/AuthAlert.vue"
  import AuthField from "@/components/auth/AuthField.vue"
  import AuthShell from "@/components/auth/AuthShell.vue"
  import AuthStatus from "@/components/auth/AuthStatus.vue"
  import { computed, ref, onMounted } from "vue"
  import { useRouter } from "vue-router"
  import { useI18n } from "vue-i18n"
  import { isRateLimited } from "@/utils/rateLimit"
  import { consumeUrlSecrets, URL_SECRET_KEYS } from "@/utils/urlSecrets"
  import { authApi } from "@/api/api"
  import {
    CircleCheck as IconCircleCheck,
    AlertCircle as IconAlertCircle,
    Lock as IconLock,
  } from "lucide-vue-next"

  const router = useRouter()
  const { t } = useI18n()

  type Status = "form" | "loading" | "success" | "error"
  type ResendStatus = "idle" | "sending" | "sent" | "failed"

  const status = ref<Status>("form")
  const password = ref("")
  const token = ref("")
  const passwordResetRequired = ref(false)
  const wrongPassword = ref(false)
  const confirmingSkip = ref(false)
  const attemptsExceeded = ref(false)
  const rateLimited = ref(false)
  const resendStatus = ref<ResendStatus>("idle")
  const email = ref("")
  const promptOpen = ref(false)

  const title = computed(() => {
    if (status.value === "loading") return t("auth.verifyEmail.verifyingTitle")
    if (status.value === "success") return t("auth.verifyEmail.successTitle")
    if (status.value === "error") return t("auth.verifyEmail.expiredTitle")
    return t("auth.verifyEmail.title")
  })
  const resendLabel = computed(() =>
    resendStatus.value === "sent"
      ? t("auth.verifyEmail.resendSent")
      : resendStatus.value === "failed"
        ? t("auth.verifyEmail.resendError")
        : t("auth.verifyEmail.resendLink"),
  )

  // Read in setup, before any navigation can copy the token into route.fullPath / history.state.
  const secrets = consumeUrlSecrets([URL_SECRET_KEYS.token], { legacyQuery: true })
  token.value = secrets[URL_SECRET_KEYS.token] ?? ""
  void router.replace({ query: {}, hash: "" })

  onMounted(() => {
    // Opening the link must not verify anything by itself; the person confirms below.
    if (!token.value) status.value = "error"
  })

  // A wrong password keeps the form (and the token); only an explicit skip wipes credentials.
  async function submit(choice: { password: string } | { skipPassword: true }) {
    status.value = "loading"
    wrongPassword.value = false
    try {
      const { data } = await authApi.verifyEmail(token.value, choice)
      passwordResetRequired.value = data.password_reset_required
      status.value = "success"
    } catch (err) {
      const code = (err as { response?: { data?: { code?: string } } }).response?.data?.code
      if (code === "WRONG_PASSWORD") {
        wrongPassword.value = true
        confirmingSkip.value = false
        status.value = "form"
        return
      }
      attemptsExceeded.value = code === "VERIFY_ATTEMPTS_EXCEEDED"
      rateLimited.value = isRateLimited(err)
      status.value = "error"
    }
  }

  function confirmEmail(address: string) {
    email.value = address
    promptOpen.value = false
    void resend()
  }

  async function resend() {
    if (!email.value) {
      promptOpen.value = true
      return
    }

    rateLimited.value = false
    resendStatus.value = "sending"
    try {
      await authApi.resendVerification(email.value)
      resendStatus.value = "sent"
    } catch (e) {
      rateLimited.value = isRateLimited(e)
      resendStatus.value = "failed"
    }
  }
</script>

<style scoped>
  .verify-form {
    display: contents;
  }

  .verify-spinner {
    color: var(--color-brand);
  }
  .verify-text {
    margin: 0;
    font-size: var(--text-small);
    line-height: 1.6;
    color: var(--color-text-secondary);
  }
</style>
