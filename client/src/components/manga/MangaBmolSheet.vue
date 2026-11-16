<template>
  <BmolAddSheet
    v-model="open"
    :user-id="userId"
    media-type="manga"
    :items="bmol.items.value"
    :preselected="preselected"
    @add="onAdd"
  />
</template>

<script setup lang="ts">
import { computed } from "vue"
import BmolAddSheet from "@/components/bmol/BmolAddSheet.vue"
import { useBmol, type BmolCandidate } from "@/components/bmol/useBmol"
import type { MangaItem } from "@/types/manga"
import { mangaToBmolCandidate } from "./mangaBmol"

/** Owns the Best of Life query so it only loads once the sheet is opened. The viewed manga is preselected. */
const props = defineProps<{ modelValue: boolean; userId: number; manga: MangaItem }>()
const emit = defineEmits<{ "update:modelValue": [value: boolean] }>()

const open = computed({ get: () => props.modelValue, set: (v) => emit("update:modelValue", v) })
const bmol = useBmol(() => props.userId)
const preselected = computed<BmolCandidate[]>(() => [mangaToBmolCandidate(props.manga)])

// One at a time: they share a rank and each answer refreshes the list.
async function onAdd(candidates: BmolCandidate[], rank: number) {
  for (const candidate of candidates) await bmol.add(candidate, rank)
  open.value = false
}
</script>
