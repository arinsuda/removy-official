<template>
  <div class="linked">
    <div class="linked__head">{{ t("navigation.link.title") }}</div>
    <div class="linked__row">
      <span :class="planzi ? 'linked__ok' : 'linked__muted'">
        {{
          planzi
            ? t("navigation.link.planziLinked")
            : t("navigation.link.planziNotLinked")
        }}
      </span>
      <button
        v-if="planzi"
        type="button"
        class="linked__btn"
        :disabled="busy"
        @click="unlink"
      >
        {{ t("navigation.link.unlink") }}
      </button>
    </div>
    <p v-if="planzi" class="linked__muted linked__small">
      {{ t("navigation.link.linkedAt", { date: linkedDate }) }}
    </p>
    <p v-else class="linked__muted linked__small">{{ t("navigation.link.linkHint") }}</p>
    <p v-if="error" class="linked__err">{{ error }}</p>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from "vue"
import { useI18n } from "vue-i18n"
import { linkApi } from "@/api/endpoints/link"
import type { LinkStatusResponse } from "@/api/endpoints/link"

const { t } = useI18n()
const links = ref<LinkStatusResponse["links"]>([])
const busy = ref(false)
const error = ref("")

const planzi = computed(() => links.value.find(l => l.platform === "planzi"))
const linkedDate = computed(() =>
  planzi.value ? new Date(planzi.value.linked_at).toLocaleDateString() : "",
)

async function load() {
  try {
    const { data } = await linkApi.status()
    links.value = data.links ?? []
  } catch {
    links.value = []
  }
}

async function unlink() {
  if (!window.confirm(t("navigation.link.unlinkConfirm"))) return
  busy.value = true
  error.value = ""
  try {
    await linkApi.unlink("planzi")
    await load()
  } catch {
    error.value = t("navigation.link.unlinkFailed")
  } finally {
    busy.value = false
  }
}

onMounted(load)
</script>

<style scoped>
.linked {
  padding: 14px 0;
  border-top: 1px solid #1c1c1c;
  margin-top: 16px;
}
.linked__head {
  font-size: 13px;
  font-weight: 600;
  margin-bottom: 8px;
}
.linked__row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}
.linked__ok {
  color: #4cd98a;
  font-size: 13px;
}
.linked__muted {
  color: #8f8f8f;
  font-size: 13px;
}
.linked__small {
  font-size: 12px;
  margin: 6px 0 0;
}
.linked__err {
  color: #ff6b6b;
  font-size: 12px;
  margin: 6px 0 0;
}
.linked__btn {
  background: transparent;
  border: 1px solid #2a2a2a;
  border-radius: 6px;
  color: #fff;
  padding: 6px 12px;
  font-size: 12px;
  cursor: pointer;
}
.linked__btn:disabled {
  opacity: 0.5;
}
</style>
