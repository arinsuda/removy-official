<template>
  <div v-if="applicable" class="auto-publish">
    <button
      type="button"
      role="switch"
      class="switch"
      :class="{ 'is-on': event.auto_publish }"
      :aria-checked="event.auto_publish"
      :aria-labelledby="labelId"
      :title="t('admin.events.autoPublish.hint')"
      :disabled="saving"
      @click="toggle"
    >
      <span class="switch__knob" />
    </button>
    <span :id="labelId" class="auto-publish__label">{{ t("admin.events.autoPublish.label") }}</span>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, useId } from "vue"
import { useI18n } from "vue-i18n"
import { adminEventApi } from "@/api/endpoints/event"
import { useToast } from "@/composables/useToast"
import type { AdminEvent } from "@/types/event"
import { autoPublishErrorKey, isGeneratedRoundType } from "@/utils/events"

const props = defineProps<{ event: AdminEvent }>()
const emit = defineEmits<{ changed: [] }>()

const { t } = useI18n()
const toast = useToast()
const labelId = `auto-publish-${useId()}`
const saving = ref(false)

/** Only voting drafts can publish on their own; a published or cancelled round has nothing to switch. */
const applicable = computed(() => props.event.status === "draft" && isGeneratedRoundType(props.event.type))

async function toggle() {
  saving.value = true
  try {
    await adminEventApi.setAutoPublish(props.event.id, !props.event.auto_publish)
    toast.success(t("admin.events.autoPublish.saved"))
    emit("changed")
  } catch (err) {
    toast.error(t(autoPublishErrorKey(err)))
  } finally {
    saving.value = false
  }
}
</script>

<style scoped>
.auto-publish {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  margin-top: 0.375rem;
}

.auto-publish__label {
  font-size: var(--text-caption);
  color: var(--color-text-muted);
}

.switch {
  position: relative;
  flex: none;
  width: 2.25rem;
  height: 1.25rem;
  cursor: pointer;
  background: var(--color-surface-3);
  border: 1px solid var(--color-border-strong);
  border-radius: var(--radius-full);
  transition: background-color var(--motion-base);
}

.switch:disabled {
  cursor: progress;
  opacity: 0.6;
}

.switch.is-on {
  background: var(--color-brand);
  border-color: var(--color-brand);
}

.switch__knob {
  position: absolute;
  top: 0.125rem;
  left: 0.125rem;
  width: 0.875rem;
  height: 0.875rem;
  background: var(--color-text-primary);
  border-radius: var(--radius-full);
  transition: transform var(--motion-base);
}

.switch.is-on .switch__knob {
  transform: translateX(1rem);
}

@media (prefers-reduced-motion: reduce) {
  .switch,
  .switch__knob {
    transition: none;
  }
}
</style>
