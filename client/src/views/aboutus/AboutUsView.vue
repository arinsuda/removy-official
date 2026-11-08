<template>
  <main class="about-page">
    <AboutHero />
    <AboutMission />
    <AboutFeatures />
    <AboutCreator />
    <AboutCta />
  </main>
</template>

<script setup lang="ts">
import AboutCreator from "@/components/about/AboutCreator.vue"
import AboutCta from "@/components/about/AboutCta.vue"
import AboutFeatures from "@/components/about/AboutFeatures.vue"
import AboutHero from "@/components/about/AboutHero.vue"
import AboutMission from "@/components/about/AboutMission.vue"
</script>

<style scoped>
.about-page {
  display: flex;
  flex-direction: column;
  gap: 2.5rem;
  width: 100%;
  max-width: 960px;
  margin: 0 auto;
  padding: 1rem 1rem 4rem;
}
@media (min-width: 768px) {
  .about-page {
    gap: 3.5rem;
    padding: 2rem 1.5rem 5rem;
  }
}
</style>
