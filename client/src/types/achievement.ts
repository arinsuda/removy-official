export interface PaginationMeta {
  page: number
  limit: number
  total: number
  total_pages: number
}

export interface PaginationParams {
  page?: number
  limit?: number
}

export interface Achievement {
  id: number
  name: string
  description: string
  exp_reward: number
  action_type: string
  target_count: number
}

export interface UserAchievement {
  achievement_id: number
  achievement: Achievement
  current_count: number
  is_unlocked: boolean
  unlocked_at: string | null
  progress_pct: number
}

export interface ListAchievementsResponse {
  data: Achievement[]
  pagination: PaginationMeta
}

export interface ListUserAchievementsResponse {
  data: UserAchievement[]
  pagination: PaginationMeta
}

export interface NewlyUnlocked {
  achievement: Achievement
  exp_gained: number
}

export type AchievementQueryParams = PaginationParams

export interface UserAchievementQueryParams extends PaginationParams {
  unlocked?: boolean
  action_type?: string
}
