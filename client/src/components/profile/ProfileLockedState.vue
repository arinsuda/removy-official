<template>
  <section class="locked">
    <AppEmptyState :icon="Lock" :title="t('profile.states.lockedTitle')" :description="t('profile.states.lockedBody')">
      <template v-if="canFollow" #action>
        <AppButton v-if="followState === 'pending'" variant="outline" :loading="busy" @click="emit('cancel')">
          {{ t("profile.follow.requested") }}
        </AppButton>
        <AppButton v-else variant="primary" :loading="busy" @click="emit('follow')">
          {{ isGuest ? t("profile.follow.logInToFollow") : t("profile.follow.follow") }}
        </AppButton>
      </template>
    </AppEmptyState>
  </section>
</template>

<script setup lang="ts">
import { useI18n } from "vue-i18n"
import { Lock } from "lucide-vue-next"
import AppButton from "@/components/ui/AppButton.vue"
import AppEmptyState from "@/components/ui/AppEmptyState.vue"

defineProps<{ canFollow: boolean; isGuest?: boolean; followState?: "pending" | "accepted"; busy?: boolean }>()
const emit = defineEmits<{ follow: []; cancel: [] }>()
const { t } = useI18n()
</script>

<style scoped>
.locked { background: var(--color-surface-2); border: 1px solid var(--color-border); border-radius: var(--radius-card); }
</style>
