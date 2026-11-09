<template>
  <BmolBoard :user-id="userId" />
</template>

<script setup lang="ts">
import BmolBoard from "@/components/bmol/BmolBoard.vue"

defineProps<{ userId: number }>()
</script>
