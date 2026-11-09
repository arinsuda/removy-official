<template>
  <AuthShell
    ref="shellRef"
    :title="sent ? t('auth.forgotPassword.sentTitle') : t('auth.forgotPassword.title')"
    :subtitle="sent ? undefined : t('auth.forgotPassword.subtitle')"
  >
    <AuthStatus v-if="sent" :icon="IconMailCheck">
      <p>{{ t("auth.forgotPassword.sentDesc", { email: submittedEmail }) }}</p>
      <AuthAlert :message="resendError" @dismiss="resendError = ''" />
      <div class="auth-status__actions">
        <AppButton variant="outline" block :disabled="resendCooldown > 0" @click="handleResend">
          {{
            resendCooldown > 0
              ? t("auth.forgotPassword.resendIn", { seconds: resendCooldown })
              : t("auth.forgotPassword.resend")
          }}
        </AppButton>
        <AppButton to="/login" block>{{ t("auth.forgotPassword.backToLogin") }}</AppButton>
      </div>
    </AuthStatus>

    <form v-else class="auth-form" novalidate @submit.prevent="handleSubmit">
      <AuthField
        v-model="email"
        type="email"
        :label="t('auth.forgotPassword.emailLabel')"
        :icon="IconMail"
        :placeholder="t('auth.forgotPassword.emailPlaceholder')"
        autocomplete="email"
        :error="touched ? error : ''"
        :disabled="isLoading"
        @input="validate"
        @blur="handleBlur"
      />
      <AuthAlert :message="serverError" @dismiss="serverError = ''" />
      <AppButton type="submit" size="lg" block class="form-item" :loading="isLoading" :disabled="!!error || !email">
        {{ t("auth.forgotPassword.submit") }}
      </AppButton>
      <p class="auth-footer form-item">
        <RouterLink to="/login" class="auth-link">{{ t("auth.forgotPassword.backToLogin") }}</RouterLink>
      </p>
    </form>
  </AuthShell>
</template>

<script setup lang="ts">
  import { ref, onUnmounted } from "vue"
  import { useI18n } from "vue-i18n"
  import { Mail as IconMail, MailCheck as IconMailCheck } from "lucide-vue-next"
  import AuthAlert from "@/components/auth/AuthAlert.vue"
  import AuthField from "@/components/auth/AuthField.vue"
  import AuthShell from "@/components/auth/AuthShell.vue"
  import AuthStatus from "@/components/auth/AuthStatus.vue"
  import AppButton from "@/components/ui/AppButton.vue"
  import { authApi } from "@/api/api"
  import { isRateLimited } from "@/utils/rateLimit"

  const { t } = useI18n()
  const shellRef = ref<InstanceType<typeof AuthShell> | null>(null)

  const email = ref("")
  const touched = ref(false)
  const error = ref("")
  const isLoading = ref(false)
  const serverError = ref("")
  const resendError = ref("")
  const sent = ref(false)
  const submittedEmail = ref("")
  const resendCooldown = ref(0)
  let cooldownTimer: ReturnType<typeof setInterval> | null = null

  const EMAIL_RE = /^[^\s@]+@[^\s@]+\.[^\s@]+$/

  function validate() {
    const v = email.value.trim()
    if (!v) {
      error.value = t("validation.email.required")
      return
    }
    if (!EMAIL_RE.test(v)) {
      error.value = t("validation.email.invalid")
      return
    }
    error.value = ""
  }

  function startCooldown() {
    resendCooldown.value = 60
    cooldownTimer = setInterval(() => {
      resendCooldown.value--
      if (resendCooldown.value <= 0 && cooldownTimer) {
        clearInterval(cooldownTimer)
        cooldownTimer = null
      }
    }, 1000)
  }

  async function handleSubmit() {
    touched.value = true
    validate()
    if (error.value || !email.value) {
      shellRef.value?.shake()
      return
    }
    serverError.value = ""
    isLoading.value = true
    try {
      await authApi.forgotPassword({ email: email.value.trim() })
      submittedEmail.value = email.value.trim()
      sent.value = true
      startCooldown()
    } catch (e) {
      serverError.value = isRateLimited(e) ? t("errors.rateLimited") : t("auth.forgotPassword.failed")
      shellRef.value?.shake()
    } finally {
      isLoading.value = false
    }
  }

  async function handleResend() {
    if (resendCooldown.value > 0) return
    resendError.value = ""
    try {
      await authApi.forgotPassword({ email: submittedEmail.value })
      startCooldown()
    } catch (e) {
      if (isRateLimited(e)) resendError.value = t("errors.rateLimited")
    }
  }

  function handleBlur() {
    touched.value = true
    validate()
  }

  onUnmounted(() => {
    if (cooldownTimer) clearInterval(cooldownTimer)
  })
</script>
