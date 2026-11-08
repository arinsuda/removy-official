<template>
  <DetailPage :config="tvDetail" :id="tvId" />
</template>

<script setup lang="ts">
import { computed } from "vue"
import { useRoute } from "vue-router"
import DetailPage from "@/components/detail/DetailPage.vue"
import { tvDetail } from "@/components/detail/detailConfig"

const route = useRoute()
const tvId = computed(() => Number(route.params.id))
</script>
