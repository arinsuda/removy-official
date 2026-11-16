<template>
  <AppEmptyState :icon="ShieldAlert" :title="t('manga.ageRestricted.title')" :description="t('manga.ageRestricted.description')">
    <template #action>
      <AppButton :to="{ path: '/settings', hash: '#profile' }" variant="outline" size="sm">{{ t("manga.ageRestricted.settings") }}</AppButton>
    </template>
  </AppEmptyState>
</template>

<script setup lang="ts">
import { useI18n } from "vue-i18n"
import { ShieldAlert } from "lucide-vue-next"
import AppButton from "@/components/ui/AppButton.vue"
import AppEmptyState from "@/components/ui/AppEmptyState.vue"

const { t } = useI18n()
</script>
