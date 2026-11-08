<template>
  <AppButton v-if="url" :href="url" rel="noopener" variant="ghost" block class="handoff-back">
    <ArrowLeft :size="16" aria-hidden="true" />
    {{ t("navigation.link.handoff.backToPlanzi") }}
  </AppButton>
</template>

<script setup lang="ts">
import { useI18n } from "vue-i18n"
import { ArrowLeft } from "lucide-vue-next"
import AppButton from "@/components/ui/AppButton.vue"

defineProps<{ url: string }>()
const { t } = useI18n()
</script>

<style scoped>
.handoff-back {
  margin-top: 8px;
}
</style>
