import { readFileSync } from "node:fs"
import { join } from "node:path"
import { describe, expect, it } from "vitest"

/**
 * Files touched by the manga social work. Routing and labels must go through
 * mediaRoute / mediaPath / mediaTypeLabelKey so manga is never sent to a movie or series page.
 */
const CHANGED_FILES = [
  "components/home/homeMedia.ts",
  "components/home/HomeFriendsStrip.vue",
  "components/feed/FeedItem.vue",
  "components/profile/tabs/ReviewTabCard.vue",
  "components/profile/tabs/WatchedTimeline.vue",
  "components/profile/components/EditReview.vue",
  "components/profile/tabs/profileMedia.ts",
  "components/bmol/BmolCard.vue",
  "components/bmol/BmolBoard.vue",
  "components/bmol/BmolAddSheet.vue",
  "components/movie/PopupCard.vue",
  "components/events/EventNomineeCard.vue",
  "components/events/awards/AwardWinnerCard.vue",
  "components/share/shareCards.ts",
  "components/share/ReviewShareTemplate.vue",
  "components/share/cards/BolSquareCard.vue",
  "components/share/cards/BolStoryCard.vue",
  "components/share/cards/PosterStoryCard.vue",
  "components/share/cards/SquareCard.vue",
  "components/share/cards/StickerCard.vue",
  "composables/useHeaderSearch.ts",
  "composables/useHomePopup.ts",
  "composables/useShareImage.ts",
  "utils/feedActivity.ts",
  "views/admin/AdminReviewsView.vue",
  "views/library/components/LibraryAddSheet.vue",
  "api/endpoints/social.ts",
  "components/manga/MangaActionBar.vue",
  "views/manga/MangaDetailView.vue",
]

/** A two-way movie/tv switch: `x === "tv" ? ... : ...` or `x === "movie" ? ... : ...`. */
const TWO_WAY_SWITCH = /===\s*["'](tv|movie)["']\s*\?/

describe("manga-aware routing in changed files", () => {
  it.each(CHANGED_FILES)("%s has no movie/tv two-way switch", (file) => {
    const src = readFileSync(join(process.cwd(), "src", file), "utf8")
    const offending = src
      .split("\n")
      .filter((line) => TWO_WAY_SWITCH.test(line) && !line.includes("tmdb-only"))
    expect(offending).toEqual([])
  })
})
