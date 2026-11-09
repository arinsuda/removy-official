<template>
  <AuthShell :title="t('auth.checkEmail.title')">
    <AuthStatus :icon="IconMailCheck">
      <p>
        {{ t("auth.checkEmail.sentText") }}<br />
        <strong>{{ email || t("auth.checkEmail.emailFallback") }}</strong>
      </p>
      <p class="auth-status__note">{{ t("auth.checkEmail.spamHint") }}</p>
      <div class="auth-status__actions">
        <AppButton
          block
          :loading="resendStatus === 'sending'"
          :disabled="resendStatus === 'sending' || resendStatus === 'sent'"
          @click="resend"
        >
          {{ resendLabel }}
        </AppButton>
      </div>
      <AuthAlert :message="rateLimited ? t('errors.rateLimited') : ''" @dismiss="rateLimited = false" />
      <RouterLink to="/login" class="auth-link">{{ t("auth.checkEmail.backToLogin") }}</RouterLink>
    </AuthStatus>

    <ResendEmailModal
      :open="promptOpen"
      :title="t('auth.checkEmail.resendBtn')"
      :label="t('auth.checkEmail.promptEmail')"
      @close="promptOpen = false"
      @submit="confirmEmail"
    />
  </AuthShell>
</template>

<script setup lang="ts">
  import { computed, ref } from "vue"
  import { RouterLink } from "vue-router"
  import { useI18n } from "vue-i18n"
  import { MailCheck as IconMailCheck } from "lucide-vue-next"
  import AuthAlert from "@/components/auth/AuthAlert.vue"
  import AuthShell from "@/components/auth/AuthShell.vue"
  import AuthStatus from "@/components/auth/AuthStatus.vue"
  import AppButton from "@/components/ui/AppButton.vue"
  import ResendEmailModal from "@/components/auth/ResendEmailModal.vue"
  import { useAuthStore } from "@/stores/auth"
  import { authApi } from "@/api/api"
  import { isRateLimited } from "@/utils/rateLimit"

  const { t } = useI18n()
  const authStore = useAuthStore()
  const email = ref(authStore.user?.email ?? "")
  const promptOpen = ref(false)

  type ResendStatus = "idle" | "sending" | "sent" | "failed"
  const resendStatus = ref<ResendStatus>("idle")
  const rateLimited = ref(false)

  const resendLabel = computed(() =>
    resendStatus.value === "sent"
      ? t("auth.checkEmail.resendSuccess")
      : resendStatus.value === "failed"
        ? t("auth.checkEmail.resendError")
        : t("auth.checkEmail.resendBtn"),
  )

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
