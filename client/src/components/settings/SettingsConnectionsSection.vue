<template>
  <SettingsCard id="connections" :title="t('settings.connections.title')">
    <div class="google">
      <svg class="google__logo" viewBox="0 0 24 24" width="20" height="20" aria-hidden="true">
        <path fill="#4285F4" d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92c-.26 1.37-1.04 2.53-2.21 3.31v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.09z" />
        <path fill="#34A853" d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23z" />
        <path fill="#FBBC05" d="M5.84 14.1c-.22-.66-.35-1.36-.35-2.1s.13-1.44.35-2.1V7.06H2.18C1.43 8.55 1 10.22 1 12s.43 3.45 1.18 4.94l2.85-2.22.81-.62z" />
        <path fill="#EA4335" d="M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.06l3.66 2.84c.87-2.6 3.3-4.52 6.16-4.52z" />
      </svg>
      <div class="google__info">
        <p class="google__title">{{ t("settings.connections.google") }}</p>
        <p v-if="status.google_connected" class="google__ok">{{ t("settings.connections.connectedAs", { email: status.google_email }) }}</p>
        <p v-else class="google__muted">{{ t("settings.connections.notConnected") }}</p>
      </div>
      <AppButton
        v-if="status.google_connected"
        variant="outline"
        size="sm"
        :loading="loading"
        :disabled="!status.can_disconnect"
        :title="!status.can_disconnect ? t('settings.connections.onlyMethod') : undefined"
        @click="disconnect"
      >
        {{ t("settings.connections.disconnect") }}
      </AppButton>
      <AppButton v-else variant="primary" size="sm" :loading="loading" @click="connect">{{ t("settings.connections.connect") }}</AppButton>
    </div>
    <p v-if="error" class="google__err" role="alert">{{ error }}</p>
    <LinkedAccounts />
  </SettingsCard>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from "vue"
import { useI18n } from "vue-i18n"
import { authApi } from "@/api/api"
import AppButton from "@/components/ui/AppButton.vue"
import LinkedAccounts from "@/components/profile/LinkedAccounts.vue"
import SettingsCard from "./SettingsCard.vue"
import { RETURN_FLAG } from "./connectionsReturn"

const { t } = useI18n()
const loading = ref(false)
const error = ref("")
const status = reactive({ enabled: false, google_connected: false, can_disconnect: false, google_email: "" })

async function fetchStatus() {
  try {
    const { data } = await authApi.getGoogleStatus()
    status.enabled = data.enabled
    status.google_connected = data.google_connected
    status.can_disconnect = data.can_disconnect
    status.google_email = data.google_email ?? ""
  } catch {
    /* ignore */
  }
}

async function connect() {
  loading.value = true
  error.value = ""
  try {
    // The server sets return_url as a URL path, so a fragment cannot be part of it.
    // Remember to scroll to this section when Google sends the user back.
    sessionStorage.setItem(RETURN_FLAG, "1")
    const { data } = await authApi.startGoogleLink("/settings")
    if (data.authorization_url) window.location.assign(data.authorization_url)
  } catch (e: unknown) {
    error.value = (e as { response?: { data?: { error?: string } } })?.response?.data?.error ?? t("settings.connections.connectFailed")
  } finally {
    loading.value = false
  }
}

async function disconnect() {
  loading.value = true
  error.value = ""
  try {
    await authApi.disconnectGoogle()
    await fetchStatus()
  } catch (e: unknown) {
    error.value = (e as { response?: { data?: { error?: string } } })?.response?.data?.error ?? t("settings.connections.disconnectFailed")
  } finally {
    loading.value = false
  }
}

onMounted(fetchStatus)
</script>

<style scoped>
.google { display: flex; align-items: center; gap: 0.75rem; }
.google__logo { flex: none; }
.google__info { flex: 1; min-width: 0; }
.google__title { font-size: 0.9375rem; font-weight: 600; color: var(--color-text-primary); }
.google__ok { font-size: 0.8125rem; color: var(--color-success); overflow-wrap: anywhere; }
.google__muted { font-size: 0.8125rem; color: var(--color-text-muted); }
.google__err { font-size: 0.75rem; color: var(--color-danger); }
</style>
