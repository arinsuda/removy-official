<template>
  <AppEmptyState :icon="icon" :title="filtered ? t('profile.tabContent.common.noMatchTitle') : isOwner ? ownerTitle : visitorTitle" :description="filtered ? t('profile.tabContent.common.noMatchBody') : isOwner ? ownerBody : visitorBody">
    <template v-if="filtered || (isOwner && ctaLabel)" #action>
      <AppButton v-if="filtered" variant="outline" size="sm" @click="emit('clear')">{{ t("profile.tabContent.common.clearFilters") }}</AppButton>
      <AppButton v-else variant="primary" size="sm" :to="ctaTo">{{ ctaLabel }}</AppButton>
    </template>
  </AppEmptyState>
</template>

<script setup lang="ts">
import type { Component } from "vue"
import type { RouteLocationRaw } from "vue-router"
import { useI18n } from "vue-i18n"
import AppButton from "@/components/ui/AppButton.vue"
import AppEmptyState from "@/components/ui/AppEmptyState.vue"

withDefaults(
  defineProps<{
    icon: Component
    isOwner: boolean
    /** True when search or filters hid every item, as opposed to a truly empty list. */
    filtered?: boolean
    ownerTitle?: string
    ownerBody?: string
    visitorTitle?: string
    visitorBody?: string
    /** Owner call to action; hidden for visitors. */
    ctaLabel?: string
    ctaTo?: RouteLocationRaw
  }>(),
  { filtered: false, ownerTitle: "", ownerBody: "", visitorTitle: "", visitorBody: "", ctaTo: () => ({ name: "movies" }) },
)
const emit = defineEmits<{ clear: [] }>()
const { t } = useI18n()
</script>
