<template>
  <HomeRail v-if="credits.length" :title="t('actor.knownFor')" :subtitle="t('actor.knownForSubtitle')">
    <HomePosterCard v-for="credit in credits" :key="credit.key" :item="credit.media" rail="actor-known-for" :caption="caption(credit)" />
  </HomeRail>
</template>

<script setup lang="ts">
import { useI18n } from "vue-i18n"
import HomePosterCard from "@/components/home/HomePosterCard.vue"
import HomeRail from "@/components/home/HomeRail.vue"
import type { ActorCredit } from "./actorCredits"
import { creditCaption } from "./creditCaption"

defineProps<{ credits: ActorCredit[] }>()
const { t } = useI18n()
const caption = (credit: ActorCredit) => creditCaption(credit, t)
</script>
