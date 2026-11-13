<template>
  <ProfileLockedState
    :can-follow="true"
    :is-guest="!auth.isLoggedIn"
    :follow-state="followState"
    :busy="follow.isPending.value || unfollow.isPending.value"
    @follow="handleFollow"
    @cancel="unfollow.mutate()"
  />
</template>

<script setup lang="ts">
import { computed, watch } from "vue"
import ProfileLockedState from "@/components/profile/ProfileLockedState.vue"
import { useAuthGate } from "@/composables/useAuthGate"
import { useFollowActions, useFollowStatus } from "@/composables/useFollow"
import { useAuthStore } from "@/stores/auth"

/** Locked state for a private profile section (library, achievements) with follow / cancel request. */
const props = defineProps<{ userId: number }>()
const emit = defineEmits<{ reload: [] }>()

const auth = useAuthStore()
const { requireAuth } = useAuthGate()
const status = useFollowStatus(() => props.userId, () => auth.isLoggedIn)
const { follow, unfollow } = useFollowActions(() => props.userId)
const followState = computed(() => status.data.value?.follow_status)

function handleFollow() {
  if (!requireAuth("follow")) return
  follow.mutate()
}

// Once the owner accepts, the section can be fetched again.
watch(followState, (state) => {
  if (state === "accepted") emit("reload")
})
</script>
