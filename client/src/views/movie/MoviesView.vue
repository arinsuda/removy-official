<template>
  <BrowsePage :config="movieBrowse" />
</template>

<script setup lang="ts">
import BrowsePage from "@/components/browse/BrowsePage.vue"
import { movieBrowse } from "@/components/browse/browseConfig"
</script>
