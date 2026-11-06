<template>
  <div class="mb-5 flex gap-4 rounded-2xl border border-white/5 bg-neutral-900/20 p-5">
    <SkeletonBlock circle width="40px" height="40px" />
    <div class="flex-1 space-y-3.5">
      <div class="flex items-center justify-between">
        <SkeletonBlock width="33%" height="16px" />
        <SkeletonBlock width="80px" height="20px" radius="999px" />
      </div>
      <SkeletonBlock width="25%" height="12px" />
      <SkeletonBlock height="80px" radius="12px" />
    </div>
  </div>
</template>

<script setup lang="ts">
import SkeletonBlock from "@/components/ui/skeleton/SkeletonBlock.vue"
</script>
