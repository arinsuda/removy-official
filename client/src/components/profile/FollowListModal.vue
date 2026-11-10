<template>
  <AppModal :model-value="modelValue" :title="t('profile.follow.listTitle')" size="sm" @update:model-value="emit('update:modelValue', $event)">
    <AppTabs v-model="tab" :tabs="tabs" :label="t('profile.follow.listTitle')" :controls="false" />
    <div class="fl" aria-live="polite">
      <AppSpinner v-if="query.isPending.value" />
      <p v-else-if="forbidden" class="fl__msg">{{ t("profile.follow.privateList") }}</p>
      <AppErrorState v-else-if="query.isError.value" :title="t('profile.states.errorTitle')" @retry="query.refetch()" />
      <p v-else-if="!rows.length" class="fl__msg">{{ t(`profile.follow.empty.${tab}`) }}</p>
      <ul v-else class="fl__list">
        <li v-for="u in rows" :key="u.id" class="fl__row">
          <RouterLink :to="`/users/${u.id}`" class="fl__user" @click="emit('update:modelValue', false)">
            <AppAvatar :src="u.avatar_url" :name="u.display_name || u.username" size="sm" decorative />
            <span class="fl__names">
              <span class="fl__name">{{ u.display_name || u.username }}</span>
              <span class="fl__handle">@{{ u.username }}</span>
            </span>
          </RouterLink>
          <span v-if="tab === 'requests'" class="fl__actions">
            <AppButton size="sm" variant="primary" :loading="respond.isPending.value" @click="respond.mutate({ id: u.id, accept: true })">
              {{ t("profile.follow.accept") }}
            </AppButton>
            <AppButton size="sm" variant="outline" :disabled="respond.isPending.value" @click="respond.mutate({ id: u.id, accept: false })">
              {{ t("profile.follow.reject") }}
            </AppButton>
          </span>
        </li>
      </ul>
    </div>
  </AppModal>
</template>

<script setup lang="ts">
import { computed, ref, watch } from "vue"
import { useI18n } from "vue-i18n"
import { useMutation, useQuery, useQueryClient } from "@tanstack/vue-query"
import { isAxiosError } from "axios"
import { followApi } from "@/api/api"
import AppAvatar from "@/components/ui/AppAvatar.vue"
import AppButton from "@/components/ui/AppButton.vue"
import AppErrorState from "@/components/ui/AppErrorState.vue"
import AppModal from "@/components/ui/AppModal.vue"
import AppSpinner from "@/components/ui/AppSpinner.vue"
import AppTabs from "@/components/ui/AppTabs.vue"
import type { FollowUserSummary } from "@/types/follow"

type ListTab = "followers" | "following" | "requests"

const props = defineProps<{ modelValue: boolean; userId: number; initialTab: "followers" | "following"; isOwner: boolean }>()
const emit = defineEmits<{ "update:modelValue": [value: boolean] }>()
const { t } = useI18n()
const queryClient = useQueryClient()

const tab = ref<ListTab>(props.initialTab)
watch(
  () => props.modelValue,
  open => {
    if (open) tab.value = props.initialTab
  },
)

const tabs = computed(() => {
  const base: { key: ListTab; label: string }[] = [
    { key: "followers", label: t("profile.hero.followers") },
    { key: "following", label: t("profile.hero.following") },
  ]
  if (props.isOwner) base.push({ key: "requests", label: t("profile.follow.requests") })
  return base
})

const fetchers: Record<ListTab, () => Promise<FollowUserSummary[]>> = {
  followers: () => followApi.getFollowers(props.userId).then(r => r.data.followers ?? []),
  following: () => followApi.getFollowing(props.userId).then(r => r.data.following ?? []),
  requests: () => followApi.getPendingRequests(props.userId).then(r => r.data.requests ?? []),
}

const query = useQuery({
  queryKey: computed(() => ["follow-list", props.userId, tab.value]),
  queryFn: () => fetchers[tab.value](),
  enabled: computed(() => props.modelValue),
  retry: false,
})
const rows = computed(() => query.data.value ?? [])
const forbidden = computed(() => isAxiosError(query.error.value) && query.error.value.response?.status === 403)

const respond = useMutation({
  mutationFn: ({ id, accept }: { id: number; accept: boolean }) =>
    accept ? followApi.acceptRequest(props.userId, id) : followApi.rejectRequest(props.userId, id),
  onSuccess: () => {
    queryClient.invalidateQueries({ queryKey: ["follow-list", props.userId] })
    queryClient.invalidateQueries({ queryKey: ["user-profile", props.userId] })
  },
})
</script>

<style scoped>
.fl { min-height: 8rem; max-height: 60svh; margin-top: 0.75rem; overflow-y: auto; }
.fl__msg { padding: 2rem 0; font-size: 0.875rem; text-align: center; color: var(--color-text-muted); }
.fl__list { display: flex; flex-direction: column; gap: 0.5rem; padding: 0; list-style: none; }
.fl__row { display: flex; align-items: center; justify-content: space-between; gap: 0.5rem; }
.fl__user { display: flex; align-items: center; gap: 0.625rem; min-width: 0; }
.fl__names { display: flex; flex-direction: column; min-width: 0; }
.fl__name { overflow: hidden; font-size: 0.875rem; font-weight: 600; color: var(--color-text-primary); text-overflow: ellipsis; white-space: nowrap; }
.fl__handle { font-size: 0.75rem; color: var(--color-text-muted); }
.fl__actions { display: flex; gap: 0.375rem; }
</style>
