<template>
  <template v-if="canLog">
    <TabIconButton data-action="history" :label="t('profile.tabContent.watched.history')" @click="emit('history')">
      <History :size="15" aria-hidden="true" />
    </TabIconButton>
    <TabIconButton data-action="log" :label="t('watchLog.recordAnother')" @click="emit('log')">
      <Plus :size="16" aria-hidden="true" />
    </TabIconButton>
  </template>
  <TabIconButton data-action="remove" danger :label="t('profile.tabContent.common.remove', { title })" @click="emit('remove')">
    <Trash2 :size="15" aria-hidden="true" />
  </TabIconButton>
</template>

<script setup lang="ts">
import { History, Plus, Trash2 } from "lucide-vue-next"
import { useI18n } from "vue-i18n"
import TabIconButton from "./TabIconButton.vue"

/** Watch logs exist for movies and TV only; the server refuses them for manga, so manga rows hide them. */
withDefaults(defineProps<{ title: string; canLog?: boolean }>(), { canLog: true })
const emit = defineEmits<{ history: []; log: []; remove: [] }>()
const { t } = useI18n()
</script>
