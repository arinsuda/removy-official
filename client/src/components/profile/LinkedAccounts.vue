<template>
  <div class="linked">
    <div class="linked__row">
      <div>
        <p class="linked__title">{{ t("navigation.link.title") }}</p>
        <p :class="planzi ? 'linked__ok' : 'linked__muted'">
          {{ planzi ? t("navigation.link.planziLinked") : t("navigation.link.planziNotLinked") }}
        </p>
      </div>
      <AppButton v-if="planzi" variant="outline" size="sm" :loading="busy" @click="confirmOpen = true">
        {{ t("navigation.link.unlink") }}
      </AppButton>
    </div>
    <p v-if="planzi" class="linked__muted linked__small">{{ t("navigation.link.linkedAt", { date: linkedDate }) }}</p>
    <p v-else class="linked__muted linked__small">{{ t("navigation.link.linkHint") }}</p>
    <p v-if="error" class="linked__err" role="alert">{{ error }}</p>

    <ConfirmModal
      v-model="confirmOpen"
      list-type="unlink_account"
      item-name="PLANZI"
      :confirm-disabled="busy"
      @confirm="unlink"
      @cancel="confirmOpen = false"
    />
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from "vue"
import { useI18n } from "vue-i18n"
import { linkApi } from "@/api/endpoints/link"
import type { LinkStatusResponse } from "@/api/endpoints/link"
import AppButton from "@/components/ui/AppButton.vue"
import ConfirmModal from "@/components/profile/components/ConfirmModal.vue"

const { t } = useI18n()
const links = ref<LinkStatusResponse["links"]>([])
const busy = ref(false)
const error = ref("")
const confirmOpen = ref(false)

const planzi = computed(() => links.value.find(l => l.platform === "planzi"))
const linkedDate = computed(() => (planzi.value ? new Date(planzi.value.linked_at).toLocaleDateString() : ""))

async function load() {
  try {
    const { data } = await linkApi.status()
    links.value = data.links ?? []
  } catch {
    links.value = []
  }
}

async function unlink() {
  busy.value = true
  error.value = ""
  try {
    await linkApi.unlink("planzi")
    await load()
    confirmOpen.value = false
  } catch {
    error.value = t("navigation.link.unlinkFailed")
  } finally {
    busy.value = false
  }
}

onMounted(load)
</script>

<style scoped>
.linked { padding-top: 1rem; border-top: 1px solid var(--color-border); }
.linked__row { display: flex; align-items: center; justify-content: space-between; gap: 0.75rem; }
.linked__title { font-size: 0.9375rem; font-weight: 600; color: var(--color-text-primary); }
.linked__ok { font-size: 0.8125rem; color: var(--color-success); }
.linked__muted { font-size: 0.8125rem; color: var(--color-text-muted); }
.linked__small { margin-top: 0.375rem; font-size: 0.75rem; }
.linked__err { margin-top: 0.375rem; font-size: 0.75rem; color: var(--color-danger); }
</style>
