<template>
  <div class="home">
    <HomeHero :items="data.hero.value.items" :source="data.hero.value.source" :loading="data.poolLoading.value" />

    <div class="home__content">
      <HomeTierRail
        :items="data.pool.value"
        :loading="data.poolLoading.value"
        :error="data.poolError.value"
        @retry="data.refetchPool()"
      />

      <HomeFriendsStrip
        :items="data.friendItems.value"
        :loading="data.friends.isLoading.value"
        :error="data.friends.isError.value"
        @retry="data.friends.refetch()"
        @visible="data.reveal.friends.value = true"
      />

      <HomeRail
        v-if="data.hasGenres.value"
        :title="t('home.recommended.title')"
        :subtitle="t('home.recommended.subtitle')"
        :loading="data.recommended.isLoading.value || !data.reveal.recommended.value"
        :error="data.recommended.isError.value"
        :empty="!data.recommendedItems.value.length"
        @retry="data.recommended.refetch()"
        @visible="data.reveal.recommended.value = true"
      >
        <HomePosterCard v-for="item in data.recommendedItems.value" :key="item.key" :item="item" rail="recommended" />
      </HomeRail>
      <HomeGenreCta v-else />

      <HomeRail
        v-if="data.unrated.value.length || data.nowPlaying.isLoading.value"
        :title="t('home.unrated.title')"
        :subtitle="t('home.unrated.subtitle')"
        :loading="data.nowPlaying.isLoading.value"
        :empty="false"
      >
        <HomePosterCard v-for="item in data.unrated.value" :key="item.key" :item="item" rail="unrated" />
      </HomeRail>

      <HomeRail
        :title="t('home.nowPlaying.title')"
        :subtitle="t('home.nowPlaying.subtitle')"
        :see-all-to="{ name: 'movies', query: { tab: 'now' } }"
        :loading="data.nowPlaying.isLoading.value"
        :error="data.nowPlaying.isError.value"
        :empty="!data.nowPlayingItems.value.length"
        @retry="data.nowPlaying.refetch()"
      >
        <HomePosterCard v-for="item in data.nowPlayingItems.value" :key="item.key" :item="item" rail="nowPlaying" />
      </HomeRail>

      <HomeRail
        :title="t('home.topRated.title')"
        :subtitle="t('home.topRated.subtitle')"
        :see-all-to="{ name: 'movies', query: { tab: 'top' } }"
        :loading="data.topRated.isLoading.value || !data.reveal.topRated.value"
        :error="data.topRated.isError.value"
        :empty="!data.topRatedItems.value.length"
        @retry="data.topRated.refetch()"
        @visible="data.reveal.topRated.value = true"
      >
        <HomePosterCard v-for="item in data.topRatedItems.value" :key="item.key" :item="item" rail="topRated" />
      </HomeRail>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from "vue-i18n"
import HomeFriendsStrip from "@/components/home/HomeFriendsStrip.vue"
import HomeGenreCta from "@/components/home/HomeGenreCta.vue"
import HomeHero from "@/components/home/HomeHero.vue"
import HomePosterCard from "@/components/home/HomePosterCard.vue"
import HomeRail from "@/components/home/HomeRail.vue"
import HomeTierRail from "@/components/home/HomeTierRail.vue"
import { useHomeData } from "@/composables/useHomeData"
import { provideHomePopup } from "@/composables/useHomePopup"

const { t } = useI18n()
provideHomePopup()
const data = useHomeData()
</script>

<style scoped>
.home {
  min-height: 100vh;
  background: var(--color-surface);
  color: var(--color-text-primary);
}
.home__content {
  display: flex;
  flex-direction: column;
  gap: 2rem;
  max-width: 1240px;
  margin: 0 auto;
  padding: 1.25rem 1rem 3rem;
}
@media (min-width: 768px) {
  .home__content {
    gap: 3rem;
    padding: 1.75rem 1.5rem 4rem;
  }
}
</style>
