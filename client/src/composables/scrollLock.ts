// Ref-counted body scroll lock: the page stays locked until every holder releases.
//
// Hazard: legacy modals still write `document.body.style.overflow` directly, and
// their cleanup sets it to "" even while an AppModal is open. Nothing re-applies
// the lock after that, so a legacy writer nested with an AppModal can unlock the
// page early. Conversely, if a legacy writer locks before an AppModal opens and
// clears overflow to "" before the AppModal closes, unlockScroll() only restores
// `previousOverflow` when `overflow` is still "hidden", avoiding re-locking the
// page to the legacy "hidden" value. Migrate such writers to lockScroll()/unlockScroll()
// when the flow that nests them with an AppModal is converted.
let locks = 0
let previousOverflow = ""

export function lockScroll() {
  if (locks++ === 0) {
    previousOverflow = document.body.style.overflow
    document.body.style.overflow = "hidden"
  }
}

export function unlockScroll() {
  if (locks > 0 && --locks === 0 && document.body.style.overflow === "hidden") {
    document.body.style.overflow = previousOverflow
  }
}
