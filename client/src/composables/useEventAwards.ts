import { computed, ref, watch } from "vue"
import { keepPreviousData, useQuery } from "@tanstack/vue-query"
import { eventApi } from "@/api/endpoints/event"
import type { AwardMediaType } from "@/types/event"
import { groupAwards } from "@/utils/awards"
import { currentVotingYear } from "@/utils/events"

/**
 * Awards page state: media type and year filters, the winners query and the year options.
 * When the selected year has no winners, it moves to the newest year that does.
 */
export function useEventAwards() {
  const mediaType = ref<AwardMediaType>("movie")
  const year = ref(currentVotingYear())

  const query = useQuery({
    queryKey: computed(() => ["event-awards", mediaType.value, year.value]),
    queryFn: () =>
      eventApi.getAwards({ media_type: mediaType.value, year: year.value }).then((res) => res.data),
    placeholderData: keepPreviousData,
  })

  watch(
    () => query.data.value,
    (data) => {
      if (!data || data.years.length === 0 || data.years.includes(year.value)) return
      year.value = Math.max(...data.years)
    },
  )

  return {
    mediaType,
    year,
    query,
    years: computed(() => query.data.value?.years ?? []),
    groups: computed(() => groupAwards(query.data.value?.awards ?? [])),
  }
}
