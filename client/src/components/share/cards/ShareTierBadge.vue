<template>
  <span v-if="tier !== 'none'" class="sc-tier" :class="`sc-tier--${tier}`">
    <span class="sc-tier__dot" />{{ t(`share.tier.${tier}`) }}
  </span>
</template>

<script setup lang="ts">
import { useI18n } from "vue-i18n"
import type { PopcornTier } from "@/components/movie/popcornTier"

defineProps<{ tier: PopcornTier }>()
const { t } = useI18n()
</script>
