<template>
  <DetailPage :config="movieDetail" :id="movieId" />
</template>

<script setup lang="ts">
import { computed } from "vue"
import { useRoute } from "vue-router"
import DetailPage from "@/components/detail/DetailPage.vue"
import { movieDetail } from "@/components/detail/detailConfig"

const route = useRoute()
const movieId = computed(() => Number(route.params.id))
</script>
