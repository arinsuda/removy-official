import { defineStore } from "pinia"
import { ref } from "vue"
import { adminApi } from "@/api/api"
import type {
  AdminOverviewStats,
  GrowthPoint,
  AdminUserRow,
  AdminReviewRow,
  AdminAuditLogRow,
  UserFilterQuery,
  EraseUserBody,
  ReviewFilterQuery,
  AuditLogFilterQuery,
} from "@/types"

export const useAdminStore = defineStore("admin", () => {
  const isAdminModeActive = ref(localStorage.getItem("admin_mode_active") === "true")

  const overview = ref<AdminOverviewStats | null>(null)
  const growth = ref<GrowthPoint[]>([])

  const users = ref<AdminUserRow[]>([])
  const usersTotal = ref(0)
  const usersPage = ref(1)
  const usersLimit = ref(20)
  const usersTotalPages = ref(1)

  const reviews = ref<AdminReviewRow[]>([])
  const reviewsTotal = ref(0)
  const reviewsPage = ref(1)
  const reviewsLimit = ref(20)
  const reviewsTotalPages = ref(1)

  const auditLogs = ref<AdminAuditLogRow[]>([])
  const auditLogsTotal = ref(0)
  const auditLogsPage = ref(1)
  const auditLogsLimit = ref(20)
  const auditLogsTotalPages = ref(1)

  const isLoadingOverview = ref(false)
  const isLoadingGrowth = ref(false)
  const isLoadingUsers = ref(false)
  const isLoadingReviews = ref(false)
  const isLoadingAuditLogs = ref(false)

  function enterAdminMode() {
    isAdminModeActive.value = true
    localStorage.setItem("admin_mode_active", "true")
  }

  function exitAdminMode() {
    isAdminModeActive.value = false
    localStorage.removeItem("admin_mode_active")
  }

  function reset() {
    exitAdminMode()
    overview.value = null
    growth.value = []
    users.value = []
    reviews.value = []
    auditLogs.value = []
  }

  async function fetchOverview() {
    isLoadingOverview.value = true
    try {
      const res = await adminApi.getOverview()
      overview.value = res.data.overview
    } finally {
      isLoadingOverview.value = false
    }
  }

  async function fetchGrowth() {
    isLoadingGrowth.value = true
    try {
      const res = await adminApi.getGrowth()
      growth.value = res.data.growth
    } finally {
      isLoadingGrowth.value = false
    }
  }

  async function fetchUsers(params: UserFilterQuery = {}) {
    isLoadingUsers.value = true
    try {
      const res = await adminApi.listUsers(params)
      users.value = res.data.items
      usersTotal.value = res.data.total
      usersPage.value = res.data.page
      usersLimit.value = res.data.limit
      usersTotalPages.value = res.data.total_pages
    } finally {
      isLoadingUsers.value = false
    }
  }

  async function updateUserRole(userId: number, role: string, reason?: string) {
    await adminApi.updateUserRole(userId, role, reason)
  }

  async function updateUserStatus(userId: number, isActive: boolean, reason?: string) {
    await adminApi.updateUserStatus(userId, isActive, reason)
  }

  async function deleteUser(userId: number, reason?: string) {
    await adminApi.deleteUser(userId, reason)
  }

  async function eraseUser(userId: number, body: EraseUserBody) {
    await adminApi.eraseUser(userId, body)
  }

  async function fetchReviews(params: ReviewFilterQuery = {}) {
    isLoadingReviews.value = true
    try {
      const res = await adminApi.listReviews(params)
      reviews.value = res.data.items
      reviewsTotal.value = res.data.total
      reviewsPage.value = res.data.page
      reviewsLimit.value = res.data.limit
      reviewsTotalPages.value = res.data.total_pages
    } finally {
      isLoadingReviews.value = false
    }
  }

  async function deleteReview(reviewId: number, reason?: string) {
    await adminApi.deleteReview(reviewId, reason)
  }

  async function fetchAuditLogs(params: AuditLogFilterQuery = {}) {
    isLoadingAuditLogs.value = true
    try {
      const res = await adminApi.listAuditLogs(params)
      auditLogs.value = res.data.items
      auditLogsTotal.value = res.data.total
      auditLogsPage.value = res.data.page
      auditLogsLimit.value = res.data.limit
      auditLogsTotalPages.value = res.data.total_pages
    } finally {
      isLoadingAuditLogs.value = false
    }
  }

  return {
    isAdminModeActive,
    overview,
    growth,
    users,
    usersTotal,
    usersPage,
    usersLimit,
    usersTotalPages,
    reviews,
    reviewsTotal,
    reviewsPage,
    reviewsLimit,
    reviewsTotalPages,
    auditLogs,
    auditLogsTotal,
    auditLogsPage,
    auditLogsLimit,
    auditLogsTotalPages,
    isLoadingOverview,
    isLoadingGrowth,
    isLoadingUsers,
    isLoadingReviews,
    isLoadingAuditLogs,
    enterAdminMode,
    exitAdminMode,
    reset,
    fetchOverview,
    fetchGrowth,
    fetchUsers,
    updateUserRole,
    updateUserStatus,
    deleteUser,
    eraseUser,
    fetchReviews,
    deleteReview,
    fetchAuditLogs,
  }
})
