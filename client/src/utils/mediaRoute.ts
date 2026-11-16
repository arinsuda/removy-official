import type { RouteLocationRaw } from "vue-router"
import type { MediaType } from "@/types/common"

/** Detail route name per media type. Route names keep the URL prefixes in one place (router/index.ts). */
const DETAIL_ROUTE_NAME: Record<MediaType, string> = {
  movie: "movie-detail",
  tv: "tv-detail",
  manga: "manga-detail",
}

/** URL prefix per media type, for code that builds path strings. Must match router/index.ts. */
const DETAIL_PATH_PREFIX: Record<MediaType, string> = {
  movie: "/movies",
  tv: "/tv",
  manga: "/manga",
}

/** Route name of a title's detail page, for `:to` objects that also take params. */
export function mediaRouteName(type: MediaType): string {
  return DETAIL_ROUTE_NAME[type]
}

/** Router location of a title's detail page: /movies/:id, /tv/:id or /manga/:id. */
export function mediaRoute(type: MediaType, id: number): RouteLocationRaw {
  return { name: DETAIL_ROUTE_NAME[type], params: { id } }
}

/** Path string of a title's detail page, for links built as strings. */
export function mediaPath(type: MediaType, id: number): string {
  return `${DETAIL_PATH_PREFIX[type]}/${id}`
}

/** i18n key of a media type's display label, e.g. `media.types.manga`. */
export function mediaTypeLabelKey(type: MediaType): string {
  return `media.types.${type}`
}
