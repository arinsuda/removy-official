import { describe, expect, it } from "vitest"
import { adminErrorKey } from "@/utils/adminError"
import { expectKeysInBothLocales } from "./helpers/realI18n"

const fail = (status: number, error: string, code?: string) => ({ response: { status, data: { error, code } } })

describe("adminErrorKey", () => {
  it.each([
    [fail(409, "cannot demote or deactivate final active admin", "FINAL_ADMIN"), "admin.errors.finalAdmin"],
    [fail(409, "cannot change role of inactive user", "INACTIVE_USER_ROLE_CHANGE"), "admin.errors.inactiveRole"],
    [fail(400, "cannot deactivate your own account", "SELF_DEACTIVATION"), "admin.errors.selfDeactivate"],
    [fail(400, "cannot delete your own account", "SELF_DELETION"), "admin.errors.selfDelete"],
    [fail(400, "user is already in requested status", "ALREADY_IN_STATUS"), "admin.errors.alreadyStatus"],
    [fail(400, "invalid role: must be 'admin' or 'user'", "INVALID_ROLE"), "admin.errors.invalidRole"],
    [fail(400, "invalid user id", "INVALID_INPUT"), "admin.errors.invalidInput"],
    [fail(404, "user not found", "USER_NOT_FOUND"), "admin.errors.notFound"],
    [fail(404, "review not found", "REVIEW_NOT_FOUND"), "admin.errors.notFound"],
    [fail(404, "not found"), "admin.errors.notFound"],
    [fail(403, "forbidden", "FORBIDDEN"), "admin.errors.forbidden"],
    [fail(401, "unauthorized", "UNAUTHORIZED"), "admin.errors.forbidden"],
    [fail(403, "forbidden"), "admin.errors.forbidden"],
    [fail(500, "internal server error", "INTERNAL_ERROR"), "admin.errors.generic"],
    [fail(409, "cannot demote or deactivate final active admin"), "admin.errors.generic"],
    [fail(400, "bad", "toString"), "admin.errors.generic"],
    [new Error("network"), "admin.errors.generic"],
  ])("maps %j", (err, key) => {
    expect(adminErrorKey(err)).toBe(key)
  })

  it("has every key in both locales", () => {
    expectKeysInBothLocales([
      "admin.errors.finalAdmin",
      "admin.errors.inactiveRole",
      "admin.errors.selfDeactivate",
      "admin.errors.selfDelete",
      "admin.errors.alreadyStatus",
      "admin.errors.invalidRole",
      "admin.errors.invalidInput",
      "admin.errors.notFound",
      "admin.errors.forbidden",
      "admin.errors.generic",
    ])
  })
})
