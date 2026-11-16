import { computed, ref, watch } from "vue"
import { useAuthStore } from "@/stores/auth"
import { isAdultByBirth, readAdultAck, writeAdultAck } from "@/components/manga/mangaAdult"

/**
 * Adult-title access for the current viewer. `eligible` comes from the profile date of birth;
 * `canShowAdult` also needs the per-user 18+ acknowledgement from the confirm modal.
 */
export function useMangaAdult() {
  const auth = useAuthStore()
  const userId = computed(() => auth.user?.id ?? null)
  const eligible = computed(() => isAdultByBirth(auth.user?.date_of_birth, new Date()))

  const acknowledged = ref(readAdultAck(userId.value))
  watch(userId, (id) => (acknowledged.value = readAdultAck(id)))

  const canShowAdult = computed(() => eligible.value && acknowledged.value)

  function acknowledge() {
    if (userId.value === null) return
    writeAdultAck(userId.value)
    acknowledged.value = true
  }

  return { eligible, acknowledged, canShowAdult, acknowledge }
}
