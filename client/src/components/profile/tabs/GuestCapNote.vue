<template>
  <p class="mt-4 text-center text-sm text-text-muted" role="note">{{ t("profile.tabContent.common.guestCap") }}</p>
</template>

<script setup lang="ts">
// Shown under a profile list when the server cut a guest's result at its cap (`has_more`).
import { useI18n } from "vue-i18n"

const { t } = useI18n()
</script>
