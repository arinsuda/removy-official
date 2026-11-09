<template>
  <SettingsCard id="password" :title="t('settings.password.title')">
    <form class="form" @submit.prevent="submit">
      <AppInput v-model="pw.old_password" :type="show ? 'text' : 'password'" autocomplete="current-password" :label="t('settings.password.current')" :disabled="loading" />
      <div>
        <AppInput v-model="pw.new_password" :type="show ? 'text' : 'password'" autocomplete="new-password" :label="t('settings.password.new')" :hint="t('settings.password.newHint')" :disabled="loading" />
        <PasswordStrength :password="pw.new_password" />
      </div>
      <AppInput v-model="pw.confirm_password" :type="show ? 'text' : 'password'" autocomplete="new-password" :label="t('settings.password.confirm')" :error="mismatch ? t('settings.password.mismatch') : ''" :disabled="loading" />
      <label class="show"><input v-model="show" type="checkbox" />{{ t("settings.password.show") }}</label>
      <p v-if="error" class="msg msg--error" role="alert">{{ error }}</p>
      <p v-if="success" class="msg msg--ok" role="status">{{ t("settings.password.success") }}</p>
      <AppButton type="submit" variant="primary" :loading="loading" :disabled="!canSubmit">{{ t("settings.password.submit") }}</AppButton>
    </form>
  </SettingsCard>
</template>

<script setup lang="ts">
import { computed, reactive, ref } from "vue"
import { useI18n } from "vue-i18n"
import { userApi } from "@/api/api"
import { resume as resumeSocket } from "@/composables/useSocket"
import PasswordStrength from "@/components/auth/PasswordStrength.vue"
import AppButton from "@/components/ui/AppButton.vue"
import AppInput from "@/components/ui/AppInput.vue"
import type { UserProfile } from "@/types/user"
import SettingsCard from "./SettingsCard.vue"

const props = defineProps<{ user: UserProfile }>()
const { t } = useI18n()

const loading = ref(false)
const error = ref("")
const success = ref(false)
const show = ref(false)
const pw = reactive({ old_password: "", new_password: "", confirm_password: "" })

const mismatch = computed(() => !!pw.confirm_password && pw.new_password !== pw.confirm_password)
const canSubmit = computed(() => pw.old_password.length > 0 && pw.new_password.length >= 8 && pw.new_password === pw.confirm_password)

async function submit() {
  if (!canSubmit.value) return
  error.value = ""
  success.value = false
  loading.value = true
  try {
    const { data } = await userApi.changePassword(props.user.id, {
      old_password: pw.old_password,
      new_password: pw.new_password,
      confirm_password: pw.confirm_password,
    })
    // The change ended every session, this one included; the server sends a fresh one.
    // Keep it, or the next request fails to refresh and signs the user out.
    if (data.access_token) localStorage.setItem("access_token", data.access_token)
    if (data.refresh_token) localStorage.setItem("refresh_token", data.refresh_token)
    resumeSocket()
    success.value = true
    pw.old_password = ""
    pw.new_password = ""
    pw.confirm_password = ""
    setTimeout(() => {
      success.value = false
    }, 4000)
  } catch (e: unknown) {
    const res = (e as { response?: { status?: number; data?: { code?: string; error?: string } } })?.response
    if (res?.data?.code === "PASSWORD_NOT_SET") error.value = t("settings.password.notSet")
    else if (res?.status === 401) error.value = t("settings.password.wrongCurrent")
    else if (res?.status === 400) error.value = res.data?.error ?? t("settings.password.invalid")
    else error.value = t("settings.password.failed")
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.form { display: flex; flex-direction: column; gap: 0.875rem; max-width: 24rem; }
.show { display: flex; align-items: center; gap: 0.5rem; font-size: 0.8125rem; color: var(--color-text-secondary); }
.msg { font-size: 0.8125rem; }
.msg--error { color: var(--color-danger); }
.msg--ok { color: var(--color-success); }
</style>
