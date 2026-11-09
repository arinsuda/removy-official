<template>
  <div class="credit-grid">
    <HomePosterCard
      v-for="credit in credits"
      :key="credit.key"
      :item="credit.media"
      rail="actor-credits"
      :sizes="SIZES"
      :caption="creditCaption(credit, t)"
    />
  </div>
</template>

<script setup lang="ts">
import { useI18n } from "vue-i18n"
import HomePosterCard from "@/components/home/HomePosterCard.vue"
import type { ActorCredit } from "./actorCredits"
import { creditCaption } from "./creditCaption"

defineProps<{ credits: ActorCredit[] }>()

const SIZES = "(min-width: 1280px) 190px, (min-width: 1024px) 18vw, (min-width: 768px) 23vw, (min-width: 480px) 31vw, 47vw"
const { t } = useI18n()
</script>

<style scoped>
.credit-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 1.25rem 0.75rem;
}

@media (min-width: 480px) {
  .credit-grid {
    grid-template-columns: repeat(3, minmax(0, 1fr));
  }
}
@media (min-width: 768px) {
  .credit-grid {
    grid-template-columns: repeat(4, minmax(0, 1fr));
    gap: 1.5rem 1rem;
  }
}
@media (min-width: 1024px) {
  .credit-grid {
    grid-template-columns: repeat(5, minmax(0, 1fr));
  }
}
@media (min-width: 1280px) {
  .credit-grid {
    grid-template-columns: repeat(6, minmax(0, 1fr));
  }
}
</style>
