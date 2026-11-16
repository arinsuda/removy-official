import { MAX_PAGES } from "@/components/browse/browseConfig"
import type { MangaCountry, MangaFormat, MangaSort, MangaStatus } from "@/types/manga"

/** Country chips: JP is manga, KR manhwa, CN manhua. */
export const MANGA_COUNTRIES: MangaCountry[] = ["JP", "KR", "CN", "all"]
export const MANGA_FORMATS: MangaFormat[] = ["all", "manga", "one_shot", "novel"]
export const MANGA_SORTS: MangaSort[] = ["popular", "top_rated", "trending", "newest"]
export const MANGA_STATUSES: MangaStatus[] = ["all", "releasing", "finished"]

export const MANGA_DEFAULT_COUNTRY: MangaCountry = "JP"
export const MANGA_DEFAULT_FORMAT: MangaFormat = "all"
export const MANGA_DEFAULT_SORT: MangaSort = "popular"
export const MANGA_DEFAULT_STATUS: MangaStatus = "all"

/** Longest genre name accepted from the URL; longer values are dropped as garbage. */
export const MANGA_GENRE_MAX_LENGTH = 40
/** Browse pages are capped the same way as the other browse lists. */
export const MANGA_PAGE_CAP = MAX_PAGES
/** Titles shown in the "similar" rail on the detail page. */
export const MANGA_SIMILAR_SIZE = 12
export const MANGA_LIST_STALE_MS = 60_000
export const MANGA_DETAIL_STALE_MS = 5 * 60_000
/** AniList scores are integers from 0 to 100. */
export const MANGA_ANILIST_SCALE = 100
export const MANGA_ANILIST_URL = "https://anilist.co"

/** Adult titles need a signed-in account whose date of birth makes it at least this old. */
export const MANGA_ADULT_MIN_AGE = 18
/** The age rule counts calendar days in Asia/Bangkok, as the server does. Thailand has no DST: a fixed UTC+7. */
export const MANGA_AGE_TIME_OFFSET_MS = 7 * 60 * 60 * 1000
/** localStorage key prefix; the signed-in user id is appended. Stores the 18+ acknowledgement. */
export const MANGA_ADULT_ACK_KEY_PREFIX = "popcorn-manga-adult-ack:"
/** Error code the detail endpoint returns for an adult title the viewer may not open. */
export const MANGA_AGE_RESTRICTED_CODE = "AGE_RESTRICTED"
