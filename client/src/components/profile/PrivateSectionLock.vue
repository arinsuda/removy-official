<template>
  <ProfileLockedState
    :can-follow="auth.isLoggedIn"
    :follow-state="followState"
    :busy="follow.isPending.value || unfollow.isPending.value"
    @follow="follow.mutate()"
    @cancel="unfollow.mutate()"
  />
</template>

<script setup lang="ts">
import { computed, watch } from "vue"
import ProfileLockedState from "@/components/profile/ProfileLockedState.vue"
import { useFollowActions, useFollowStatus } from "@/composables/useFollow"
import { useAuthStore } from "@/stores/auth"

/** Locked state for a private profile section (library, achievements) with follow / cancel request. */
const props = defineProps<{ userId: number }>()
const emit = defineEmits<{ reload: [] }>()

const auth = useAuthStore()
const status = useFollowStatus(() => props.userId, () => auth.isLoggedIn)
const { follow, unfollow } = useFollowActions(() => props.userId)
const followState = computed(() => status.data.value?.follow_status)

// Once the owner accepts, the section can be fetched again.
watch(followState, (state) => {
  if (state === "accepted") emit("reload")
})
</script>
