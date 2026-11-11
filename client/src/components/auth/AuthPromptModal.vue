<template>
  <AppModal
    v-model="isOpen"
    size="sm"
    :title="t(`auth.prompt.reasons.${reason}.title`)"
  >
    <p class="auth-prompt__body">{{ t(`auth.prompt.reasons.${reason}.body`) }}</p>
    <template #footer>
      <AppButton variant="ghost" @click="close">
        {{ t("auth.prompt.actions.cancel") }}
      </AppButton>
      <AppButton
        :to="{ path: '/login', query: authQuery }"
        variant="outline"
        data-test="auth-prompt-login"
        @click="close"
      >
        {{ t("auth.prompt.actions.login") }}
      </AppButton>
      <AppButton
        :to="{ path: '/register', query: authQuery }"
        data-test="auth-prompt-signup"
        @click="close"
      >
        {{ t("auth.prompt.actions.signup") }}
      </AppButton>
    </template>
  </AppModal>
</template>

<script setup lang="ts">
import { computed } from "vue"
import { useI18n } from "vue-i18n"
import { useRoute } from "vue-router"
import AppButton from "@/components/ui/AppButton.vue"
import AppModal from "@/components/ui/AppModal.vue"
import { useAuthGate } from "@/composables/useAuthGate"
import { buildAuthQuery } from "@/utils/safeRedirect"

const { t } = useI18n()
const route = useRoute()
const { isOpen, reason, close } = useAuthGate()

const authQuery = computed(() =>
  buildAuthQuery(
    route?.fullPath ??
      (typeof window !== "undefined"
        ? `${window.location.pathname}${window.location.search}${window.location.hash}`
        : "/"),
  ),
)
</script>

<style scoped>
.auth-prompt__body {
  margin: 0;
  font-size: var(--text-small);
  line-height: 1.6;
  color: var(--color-text-secondary);
}
</style>
