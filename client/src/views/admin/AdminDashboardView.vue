<template>
  <div class="admin-dashboard">
    <AdminPageHeader :title="t('admin.dashboard.title')" :subtitle="t('admin.dashboard.subtitle')" />

    <AppErrorState
      v-if="loadError && !overview"
      :title="t('admin.dashboard.errorTitle')"
      :description="t('admin.table.errorDesc')"
      @retry="load"
    />

    <SkeletonGrid v-else-if="adminStore.isLoadingOverview && !overview" :count="4" min-width="220px" aspect="2 / 1" :caption="false" />

    <template v-else-if="overview">
      <section class="kpi-grid" :aria-label="t('admin.dashboard.title')">
        <AdminKpiCard :label="t('admin.dashboard.totalRegisteredUsers')" :icon="Users">
          <template #value>{{ overview.total_registered_users.toLocaleString() }}</template>
          <template #sub>
            {{ t("admin.dashboard.activeUsers") }} {{ overview.active_users_count.toLocaleString() }} ·
            {{ t("admin.dashboard.inactiveUsers") }} {{ overview.inactive_users_count.toLocaleString() }}
          </template>
        </AdminKpiCard>

        <AdminKpiCard :label="t('admin.dashboard.growth')" :icon="TrendingUp" tone="success">
          <template #value>
            <template v-if="overview.growth_percentage !== null">
              {{ overview.growth_percentage > 0 ? "+" : "" }}{{ overview.growth_percentage.toFixed(1) }}%
            </template>
            <span v-else class="kpi-note">{{ t("admin.dashboard.noPreviousBaseline") }}</span>
          </template>
          <template #sub>
            <span :class="overview.absolute_growth >= 0 ? 'is-up' : 'is-down'">
              {{ overview.absolute_growth >= 0 ? "+" : "" }}{{ overview.absolute_growth }}
            </span>
            {{
              t("admin.dashboard.vsLastMonth", {
                previous: overview.previous_month_registrations,
                current: overview.current_month_registrations,
              })
            }}
          </template>
        </AdminKpiCard>

        <AdminKpiCard :label="t('admin.dashboard.realtimeConnectedUsers')" :icon="Wifi" tone="info">
          <template #value>
            {{ overview.unique_online_users.toLocaleString() }}
            <span class="live-dot" aria-hidden="true" />
          </template>
          <template #sub>{{ t("admin.dashboard.socketsNow") }}</template>
        </AdminKpiCard>

        <AdminKpiCard :label="t('admin.dashboard.activeActors')" :icon="Activity" tone="warning">
          <template #value>{{ overview.dau_today.toLocaleString() }}</template>
          <template #sub>{{ t("admin.dashboard.dauToday") }}</template>
        </AdminKpiCard>
      </section>

      <div class="chart-row">
        <section class="panel panel--wide">
          <AdminBarChart
            v-if="growthPoints.length"
            :points="growthPoints"
            :caption="t('admin.dashboard.growthChartTitle')"
            :summary="chartSummary"
            :label-header="t('admin.dashboard.chartMonth')"
            :value-header="t('admin.dashboard.chartCount')"
          />
          <template v-else>
            <h2 class="panel__title">{{ t("admin.dashboard.growthChartTitle") }}</h2>
            <AppEmptyState :title="t('admin.dashboard.noGrowthData')" :icon="BarChart3" />
          </template>
        </section>

        <div class="side-stack">
          <section class="panel">
            <AdminStackBar :caption="t('admin.dashboard.accountStatus')" :summary="statusSummary" :segments="statusSegments" />
          </section>
          <section class="panel">
            <AdminMeterList :caption="t('admin.dashboard.activityTitle')" :items="activityItems" />
          </section>
        </div>
      </div>

      <section class="kpi-grid kpi-grid--compact">
        <AdminKpiCard :label="t('admin.dashboard.totalActivities')" :icon="Zap" tone="warning">
          <template #value>{{ overview.total_activity_events.toLocaleString() }}</template>
          <template #sub>{{ t("admin.dashboard.todayCount", { count: overview.activity_events_today.toLocaleString() }) }}</template>
        </AdminKpiCard>
        <AdminKpiCard :label="t('admin.dashboard.totalReviews')" :icon="MessageSquare">
          <template #value>{{ overview.total_reviews.toLocaleString() }}</template>
        </AdminKpiCard>
        <AdminKpiCard :label="t('admin.dashboard.totalMediaLikes')" :icon="Heart" tone="success">
          <template #value>{{ overview.total_media_likes.toLocaleString() }}</template>
        </AdminKpiCard>
      </section>
    </template>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from "vue"
import { useI18n } from "vue-i18n"
import { Activity, BarChart3, Heart, MessageSquare, TrendingUp, Users, Wifi, Zap } from "lucide-vue-next"
import AdminBarChart from "@/components/admin/AdminBarChart.vue"
import AdminKpiCard from "@/components/admin/AdminKpiCard.vue"
import AdminMeterList from "@/components/admin/AdminMeterList.vue"
import AdminPageHeader from "@/components/admin/AdminPageHeader.vue"
import AdminStackBar from "@/components/admin/AdminStackBar.vue"
import AppEmptyState from "@/components/ui/AppEmptyState.vue"
import AppErrorState from "@/components/ui/AppErrorState.vue"
import SkeletonGrid from "@/components/ui/skeleton/SkeletonGrid.vue"
import { useAdminStore } from "@/stores/admin"

const { t, locale } = useI18n()
const adminStore = useAdminStore()
const overview = computed(() => adminStore.overview)
const loadError = ref(false)

// The API sends months as "YYYY-MM"; show them in the active locale.
const monthFormat = computed(() => new Intl.DateTimeFormat(locale.value, { month: "short", year: "2-digit", timeZone: "UTC" }))
function monthLabel(month: string) {
  const d = new Date(`${month}-01T00:00:00Z`)
  return Number.isNaN(d.getTime()) ? month : monthFormat.value.format(d)
}

const growthPoints = computed(() => (adminStore.growth || []).map((p) => ({ label: monthLabel(p.month), value: p.user_count })))
const chartSummary = computed(() => {
  const pts = growthPoints.value
  const peak = pts.reduce((best, p) => (p.value > best.value ? p : best), pts[0] ?? { label: "", value: 0 })
  return t("admin.dashboard.chartSummary", { count: pts.length, peak: peak.value, month: peak.label })
})

const statusSegments = computed(() => [
  { label: t("admin.dashboard.activeUsers"), value: overview.value?.active_users_count ?? 0, tone: "success" as const },
  { label: t("admin.dashboard.inactiveUsers"), value: overview.value?.inactive_users_count ?? 0, tone: "danger" as const },
])
const statusSummary = computed(() => statusSegments.value.map((s) => `${s.label}: ${s.value}`).join(", "))

const activityItems = computed(() => [
  { label: t("admin.dashboard.windowDay"), value: overview.value?.dau_today ?? 0 },
  { label: t("admin.dashboard.windowWeek"), value: overview.value?.wau_7d ?? 0 },
  { label: t("admin.dashboard.windowMonth"), value: overview.value?.mau_30d ?? 0 },
])

async function load() {
  loadError.value = false
  const results = await Promise.allSettled([adminStore.fetchOverview(), adminStore.fetchGrowth()])
  loadError.value = results[0].status === "rejected"
}

onMounted(load)
</script>

<style scoped>
.admin-dashboard {
  display: flex;
  flex-direction: column;
  gap: 1.25rem;
}

.kpi-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(14rem, 1fr));
  gap: 1rem;
}

.kpi-grid--compact {
  grid-template-columns: repeat(auto-fit, minmax(12rem, 1fr));
}

.kpi-note {
  font-size: var(--text-small);
  font-weight: 600;
  color: var(--color-text-secondary);
}

.is-up {
  color: var(--color-success);
}

.is-down {
  color: var(--color-danger);
}

.live-dot {
  width: 0.625rem;
  height: 0.625rem;
  border-radius: 50%;
  background: var(--color-success);
  animation: live-pulse 2s infinite;
}

@keyframes live-pulse {
  0% {
    box-shadow: 0 0 0 0 color-mix(in srgb, var(--color-success) 60%, transparent);
  }
  70%,
  100% {
    box-shadow: 0 0 0 0.5rem transparent;
  }
}

.chart-row {
  display: grid;
  grid-template-columns: minmax(0, 1fr);
  gap: 1rem;
}

.panel {
  min-width: 0;
  padding: 1.25rem;
  background: var(--color-surface-2);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-card);
}

.panel__title {
  margin: 0 0 0.75rem;
  font-size: var(--text-small);
  font-weight: 700;
  color: var(--color-text-primary);
}

.side-stack {
  display: grid;
  grid-template-columns: minmax(0, 1fr);
  gap: 1rem;
  align-content: start;
}

@media (min-width: 1024px) {
  .chart-row {
    grid-template-columns: minmax(0, 2fr) minmax(0, 1fr);
  }
}

@media (prefers-reduced-motion: reduce) {
  .live-dot {
    animation: none;
  }
}
</style>
