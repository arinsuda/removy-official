<template>
  <BrowsePage :config="tvBrowse" />
</template>

<script setup lang="ts">
import BrowsePage from "@/components/browse/BrowsePage.vue"
import { tvBrowse } from "@/components/browse/browseConfig"
</script>
