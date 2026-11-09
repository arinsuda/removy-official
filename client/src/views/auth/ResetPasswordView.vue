<template>
  <AuthShell
    ref="shellRef"
    :title="title"
    :subtitle="!tokenError && !success ? t('auth.resetPassword.subtitle') : undefined"
  >
    <AuthStatus v-if="tokenError" :icon="IconShieldOff" tone="danger">
      <p>{{ t("auth.resetPassword.invalidDesc") }}</p>
      <div class="auth-status__actions">
        <AppButton to="/forgot-password" block>{{ t("auth.resetPassword.requestNew") }}</AppButton>
        <AppButton to="/login" variant="ghost" block>{{ t("auth.forgotPassword.backToLogin") }}</AppButton>
      </div>
    </AuthStatus>

    <AuthStatus v-else-if="success" :icon="IconShieldCheck" tone="success">
      <p>{{ t("auth.resetPassword.doneDesc") }}</p>
      <div class="auth-status__actions">
        <AppButton to="/login" block>{{ t("auth.resetPassword.goToLogin") }}</AppButton>
      </div>
    </AuthStatus>

    <form v-else class="auth-form" novalidate @submit.prevent="handleSubmit">
      <AuthField
        v-model="form.new_password"
        type="password"
        :label="t('auth.resetPassword.passwordLabel')"
        :icon="IconLock"
        :placeholder="t('auth.resetPassword.passwordPlaceholder')"
        :hint="t('auth.resetPassword.hint')"
        autocomplete="new-password"
        :disabled="isLoading"
      >
        <PasswordStrength :password="form.new_password" />
      </AuthField>
      <AuthField
        v-model="form.confirm_password"
        type="password"
        :label="t('auth.resetPassword.confirmPasswordLabel')"
        :icon="IconLock"
        :placeholder="t('auth.resetPassword.confirmPasswordPlaceholder')"
        :error="mismatch ? t('auth.resetPassword.mismatch') : ''"
        autocomplete="new-password"
        :disabled="isLoading"
      />
      <AuthAlert :message="errorMsg" @dismiss="errorMsg = ''" />
      <AppButton type="submit" size="lg" block class="form-item" :loading="isLoading" :disabled="!canSubmit">
        {{ t("auth.resetPassword.submit") }}
      </AppButton>
      <p class="auth-footer form-item">
        <RouterLink to="/login" class="auth-link">{{ t("auth.forgotPassword.backToLogin") }}</RouterLink>
      </p>
    </form>
  </AuthShell>
</template>

<script setup lang="ts">
  import { ref, computed } from "vue"
  import { useRoute } from "vue-router"
  import { useI18n } from "vue-i18n"
  import { Lock as IconLock, ShieldOff as IconShieldOff, ShieldCheck as IconShieldCheck } from "lucide-vue-next"
  import AuthAlert from "@/components/auth/AuthAlert.vue"
  import AuthField from "@/components/auth/AuthField.vue"
  import AuthShell from "@/components/auth/AuthShell.vue"
  import AuthStatus from "@/components/auth/AuthStatus.vue"
  import PasswordStrength from "@/components/auth/PasswordStrength.vue"
  import AppButton from "@/components/ui/AppButton.vue"
  import { authApi } from "@/api/api"

  const route = useRoute()
  const { t } = useI18n()
  const shellRef = ref<InstanceType<typeof AuthShell> | null>(null)

  const token = route.query.token as string | undefined
  const userId = route.query.uid ? Number(route.query.uid) : undefined

  const tokenError = ref(!token || !userId || isNaN(userId as number))
  const success = ref(false)
  const isLoading = ref(false)
  const errorMsg = ref("")

  const form = ref({ new_password: "", confirm_password: "" })

  const title = computed(() =>
    tokenError.value
      ? t("auth.resetPassword.invalidTitle")
      : success.value
        ? t("auth.resetPassword.doneTitle")
        : t("auth.resetPassword.title"),
  )
  const mismatch = computed(
    () => !!form.value.confirm_password && form.value.new_password !== form.value.confirm_password,
  )
  const canSubmit = computed(
    () => form.value.new_password.length >= 8 && form.value.new_password === form.value.confirm_password,
  )

  async function handleSubmit() {
    if (!canSubmit.value) return
    errorMsg.value = ""
    isLoading.value = true
    try {
      await authApi.resetPassword({
        token: token!,
        user_id: userId!,
        new_password: form.value.new_password,
        confirm_password: form.value.confirm_password,
      })
      success.value = true
    } catch (e: any) {
      const status = e?.response?.status
      if (status === 404 || status === 410 || status === 422) {
        tokenError.value = true
      } else if (status === 400) {
        errorMsg.value = e?.response?.data?.error ?? t("auth.resetPassword.invalidRequest")
      } else {
        errorMsg.value = t("auth.resetPassword.failed")
      }
      shellRef.value?.shake()
    } finally {
      isLoading.value = false
    }
  }
</script>
