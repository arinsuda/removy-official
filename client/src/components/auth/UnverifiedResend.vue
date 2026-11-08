<template>
  <div class="unverified" role="status">
    <AuthField
      v-if="!identifierIsEmail"
      v-model="typedEmail"
      :label="t('auth.login.resendEmailLabel')"
      :icon="IconMail"
      type="email"
      autocomplete="email"
      :disabled="sending"
    />
    <AppButton
      type="button"
      variant="outline"
      size="sm"
      block
      :loading="sending"
      :disabled="!targetEmail"
      @click="resend"
    >
      {{ t("auth.login.resendVerification") }}
    </AppButton>
    <p v-if="note" class="unverified__note">{{ note }}</p>
  </div>
</template>

<script setup lang="ts">
  import { computed, ref } from "vue"
  import { useI18n } from "vue-i18n"
  import { Mail as IconMail } from "lucide-vue-next"
  import AuthField from "@/components/auth/AuthField.vue"
  import AppButton from "@/components/ui/AppButton.vue"
  import { authApi } from "@/api/api"
  import { useUIStore } from "@/stores/ui"
  import { isRateLimited } from "@/utils/rateLimit"

  const props = defineProps<{ identifier: string }>()
  const { t } = useI18n()
  const ui = useUIStore()

  const typedEmail = ref("")
  const sending = ref(false)
  const note = ref("")

  const identifierIsEmail = computed(() => props.identifier.includes("@"))
  const targetEmail = computed(() =>
    (identifierIsEmail.value ? props.identifier : typedEmail.value).trim(),
  )

  async function resend() {
    if (!targetEmail.value || sending.value) return
    sending.value = true
    note.value = ""
    try {
      await authApi.resendVerification(targetEmail.value)
      ui.toastSuccess(t("auth.login.resendSent"))
    } catch (e) {
      const status = (e as { response?: { status?: number } })?.response?.status
      if (isRateLimited(e)) note.value = t("auth.login.resendRateLimited")
      else if (status === 409) note.value = t("auth.login.resendAlreadyVerified")
      else note.value = t("auth.login.resendFailed")
    } finally {
      sending.value = false
    }
  }
</script>

<style scoped>
  .unverified {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
  }
  .unverified__note {
    margin: 0;
    font-size: 0.8125rem;
    color: var(--color-text-secondary);
  }
</style>
