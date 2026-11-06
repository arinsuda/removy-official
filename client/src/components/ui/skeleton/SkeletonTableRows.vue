<template>
  <tr v-for="row in rows" :key="row" aria-hidden="true">
    <td v-for="col in columns" :key="col" class="skeleton-cell">
      <SkeletonBlock :width="col % 2 ? '75%' : '50%'" />
    </td>
  </tr>
</template>

<script setup lang="ts">
import SkeletonBlock from "./SkeletonBlock.vue"

withDefaults(defineProps<{ columns: number; rows?: number }>(), { rows: 6 })
</script>

<style scoped>
.skeleton-cell {
  padding: 16px 12px;
}
</style>
