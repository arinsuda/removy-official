<template>
  <AppModal :model-value="open" size="sm" :title="title" @close="emit('close')">
    <form :id="formId" class="auth-form" @submit.prevent="submit">
      <AuthField
        v-model="email"
        type="email"
        :label="label"
        :placeholder="t('auth.login.emailPlaceholder')"
        :icon="IconMail"
        autocomplete="email"
      />
    </form>
    <template #footer>
      <AppButton variant="outline" @click="emit('close')">{{ t("common.cancel") }}</AppButton>
      <AppButton type="submit" :form="formId" :disabled="!email.trim()">{{ title }}</AppButton>
    </template>
  </AppModal>
</template>

<script setup lang="ts">
import { ref, useId } from "vue"
import { useI18n } from "vue-i18n"
import { Mail as IconMail } from "lucide-vue-next"
import AppButton from "@/components/ui/AppButton.vue"
import AppModal from "@/components/ui/AppModal.vue"
import AuthField from "./AuthField.vue"

defineProps<{ open: boolean; title: string; label: string }>()
const emit = defineEmits<{ close: []; submit: [email: string] }>()

const { t } = useI18n()
const formId = `resend-email-${useId()}`
const email = ref("")

function submit() {
  const trimmed = email.value.trim()
  if (trimmed) emit("submit", trimmed)
}
</script>
