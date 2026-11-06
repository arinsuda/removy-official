<template>
  <div class="admin-dashboard">
    <h1 class="page-title">{{ $t("admin.dashboard.title") }}</h1>

    <div v-if="adminStore.isLoadingOverview" class="loading-state">
      <div class="spinner" />
    </div>

    <template v-else-if="overview">
      <!-- Top Overview Metrics Grid -->
      <div class="metrics-grid">
        <!-- Card 1: Registered Users -->
        <div class="stat-card">
          <div class="stat-header">
            <span class="stat-label">{{ $t("admin.dashboard.totalRegisteredUsers") }}</span>
            <div class="stat-icon-badge brand"><Users :size="20" /></div>
          </div>
          <div class="stat-value">{{ overview.total_registered_users.toLocaleString() }}</div>
          <div class="stat-sub text-muted">
            {{ $t("admin.dashboard.activeUsers") }}: {{ overview.active_users_count }} |
            {{ $t("admin.dashboard.inactiveUsers") }}: {{ overview.inactive_users_count }}
          </div>
        </div>

        <!-- Card 2: Growth Rate -->
        <div class="stat-card">
          <div class="stat-header">
            <span class="stat-label">{{ $t("admin.dashboard.growth") }}</span>
            <div class="stat-icon-badge success"><TrendingUp :size="20" /></div>
          </div>
          <div class="stat-value">
            <span v-if="overview.growth_percentage !== null">
              {{ overview.growth_percentage > 0 ? '+' : '' }}{{ overview.growth_percentage.toFixed(1) }}%
            </span>
            <span v-else class="text-secondary text-base">
              {{ $t("admin.dashboard.noPreviousBaseline") }}
            </span>
          </div>
          <div class="stat-sub">
            <span :class="overview.absolute_growth >= 0 ? 'text-success' : 'text-danger'">
              {{ overview.absolute_growth >= 0 ? '+' : '' }}{{ overview.absolute_growth }}
            </span>
            vs last month ({{ overview.previous_month_registrations }} → {{ overview.current_month_registrations }})
          </div>
        </div>

        <!-- Card 3: Realtime Connected Users -->
        <div class="stat-card">
          <div class="stat-header">
            <span class="stat-label">{{ $t("admin.dashboard.realtimeConnectedUsers") }}</span>
            <div class="stat-icon-badge info pulse"><Wifi :size="20" /></div>
          </div>
          <div class="stat-value flex-align">
            <span>{{ overview.unique_online_users }}</span>
            <span class="online-indicator" />
          </div>
          <div class="stat-sub text-muted">Sockets connected now</div>
        </div>

        <!-- Card 4: DAU / WAU / MAU -->
        <div class="stat-card">
          <div class="stat-header">
            <span class="stat-label">Active Actors (DAU / WAU / MAU)</span>
            <div class="stat-icon-badge warning"><Activity :size="20" /></div>
          </div>
          <div class="stat-value text-xl">
            DAU: {{ overview.dau_today }}
          </div>
          <div class="stat-sub text-muted">
            7d (WAU): {{ overview.wau_7d }} | 30d (MAU): {{ overview.mau_30d }}
          </div>
        </div>
      </div>

      <!-- Secondary Activity Metrics -->
      <div class="metrics-grid secondary">
        <div class="stat-card compact">
          <div class="stat-label">{{ $t("admin.dashboard.totalActivities") }}</div>
          <div class="stat-value-sm">{{ overview.total_activity_events.toLocaleString() }}</div>
          <div class="stat-sub text-muted">{{ overview.activity_events_today }} today</div>
        </div>
        <div class="stat-card compact">
          <div class="stat-label">{{ $t("admin.dashboard.totalReviews") }}</div>
          <div class="stat-value-sm">{{ overview.total_reviews.toLocaleString() }}</div>
        </div>
        <div class="stat-card compact">
          <div class="stat-label">{{ $t("admin.dashboard.totalMediaLikes") }}</div>
          <div class="stat-value-sm">{{ overview.total_media_likes.toLocaleString() }}</div>
        </div>
      </div>

      <!-- Growth Line Chart -->
      <div class="chart-card">
        <h2 class="chart-title">{{ $t("admin.dashboard.growthChartTitle") }}</h2>
        <div class="chart-container">
          <Line v-if="chartData.labels.length > 0" :data="chartData" :options="chartOptions" />
        </div>
      </div>
    </template>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted } from "vue"
import { useAdminStore } from "@/stores/admin"
import { Users, TrendingUp, Wifi, Activity } from "lucide-vue-next"
import {
  Chart as ChartJS,
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  Title,
  Tooltip,
  Legend,
  Filler,
} from "chart.js"
import { Line } from "vue-chartjs"

ChartJS.register(
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  Title,
  Tooltip,
  Legend,
  Filler
)

const adminStore = useAdminStore()
const overview = computed(() => adminStore.overview)

const chartData = computed(() => {
  const points = adminStore.growth || []
  return {
    labels: points.map((p) => p.month),
    datasets: [
      {
        label: "New Registrations",
        data: points.map((p) => p.user_count),
        borderColor: "#f5432c",
        backgroundColor: "rgba(245, 67, 44, 0.15)",
        fill: true,
        tension: 0.35,
        pointBackgroundColor: "#f5432c",
        pointBorderColor: "#ffffff",
        pointRadius: 4,
      },
    ],
  }
})

const chartOptions = {
  responsive: true,
  maintainAspectRatio: false,
  plugins: {
    legend: { display: false },
    tooltip: {
      backgroundColor: "#1f1f1f",
      titleColor: "#ffffff",
      bodyColor: "#f5432c",
      borderColor: "#2a2a2a",
      borderWidth: 1,
    },
  },
  scales: {
    x: {
      grid: { color: "rgba(255, 255, 255, 0.05)" },
      ticks: { color: "#a3a3a3" },
    },
    y: {
      beginAtZero: true,
      grid: { color: "rgba(255, 255, 255, 0.05)" },
      ticks: { color: "#a3a3a3", stepSize: 1 },
    },
  },
}

onMounted(() => {
  adminStore.fetchOverview()
  adminStore.fetchGrowth()
})
</script>

<style scoped>
.admin-dashboard {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}

.page-title {
  font-size: 1.6rem;
  font-weight: 800;
  margin: 0;
  color: var(--color-text-primary);
}

.metrics-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(260px, 1fr));
  gap: 1.25rem;
}

.metrics-grid.secondary {
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
}

.stat-card {
  background: var(--color-surface-2);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-card);
  padding: 1.25rem;
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.2);
}

.stat-card.compact {
  padding: 1rem;
}

.stat-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.stat-label {
  font-size: 0.825rem;
  font-weight: 600;
  color: var(--color-text-secondary);
}

.stat-icon-badge {
  width: 36px;
  height: 36px;
  border-radius: 0.5rem;
  display: flex;
  align-items: center;
  justify-content: center;
}
.stat-icon-badge.brand { background: rgba(245, 67, 44, 0.15); color: var(--color-brand); }
.stat-icon-badge.success { background: rgba(16, 185, 129, 0.15); color: #10b981; }
.stat-icon-badge.info { background: rgba(59, 130, 246, 0.15); color: #3b82f6; }
.stat-icon-badge.warning { background: rgba(245, 158, 11, 0.15); color: #f59e0b; }

.stat-value {
  font-size: 1.8rem;
  font-weight: 800;
  color: var(--color-text-primary);
  line-height: 1.2;
}

.stat-value.text-xl {
  font-size: 1.3rem;
}

.stat-value-sm {
  font-size: 1.4rem;
  font-weight: 800;
  color: var(--color-text-primary);
}

.flex-align {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.online-indicator {
  width: 10px;
  height: 10px;
  background-color: #10b981;
  border-radius: 50%;
  box-shadow: 0 0 8px #10b981;
  animation: pulse-online 2s infinite;
}

@keyframes pulse-online {
  0% { transform: scale(0.95); box-shadow: 0 0 0 0 rgba(16, 185, 129, 0.7); }
  70% { transform: scale(1); box-shadow: 0 0 0 8px rgba(16, 185, 129, 0); }
  100% { transform: scale(0.95); box-shadow: 0 0 0 0 rgba(16, 185, 129, 0); }
}

.stat-sub {
  font-size: 0.775rem;
}

.text-muted { color: var(--color-text-muted); }
.text-secondary { color: var(--color-text-secondary); }
.text-success { color: #10b981; }
.text-danger { color: #ef4444; }
.text-base { font-size: 0.9rem; font-weight: 600; }

.chart-card {
  background: var(--color-surface-2);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-card);
  padding: 1.5rem;
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.chart-title {
  font-size: 1.1rem;
  font-weight: 700;
  color: var(--color-text-primary);
  margin: 0;
}

.chart-container {
  height: 320px;
  position: relative;
}

.loading-state {
  display: flex;
  justify-content: center;
  padding: 4rem;
}

.spinner {
  width: 36px;
  height: 36px;
  border: 3px solid rgba(255, 255, 255, 0.1);
  border-top-color: var(--color-brand);
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}
</style>
