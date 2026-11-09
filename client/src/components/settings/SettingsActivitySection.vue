<template>
  <SettingsCard id="activity" :title="t('settings.activity.title')" :description="t('settings.activity.description')">
    <ActivityToggleList />
  </SettingsCard>
</template>

<script setup lang="ts">
import { useI18n } from "vue-i18n"
import ActivityToggleList from "./ActivityToggleList.vue"
import SettingsCard from "./SettingsCard.vue"

const { t } = useI18n()
</script>
