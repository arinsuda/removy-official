import { describe, expect, it } from "vitest"
import { adminErrorKey } from "@/utils/adminError"
import { expectKeysInBothLocales } from "./helpers/realI18n"

const fail = (status: number, error: string, code?: string) => ({ response: { status, data: { error, code } } })

describe("adminErrorKey", () => {
  it.each([
    [fail(409, "cannot demote or deactivate final active admin"), "admin.errors.finalAdmin"],
    [fail(409, "x", "FINAL_ADMIN"), "admin.errors.finalAdmin"],
    [fail(409, "cannot change role of inactive user"), "admin.errors.inactiveRole"],
    [fail(400, "cannot deactivate your own account"), "admin.errors.selfDeactivate"],
    [fail(400, "cannot delete your own account"), "admin.errors.selfDelete"],
    [fail(400, "user is already in requested status"), "admin.errors.alreadyStatus"],
    [fail(400, "invalid role: must be 'admin' or 'user'"), "admin.errors.invalidRole"],
    [fail(404, "user not found"), "admin.errors.notFound"],
    [fail(403, "forbidden"), "admin.errors.forbidden"],
    [fail(500, "internal server error"), "admin.errors.generic"],
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
      "admin.errors.notFound",
      "admin.errors.forbidden",
      "admin.errors.generic",
    ])
  })
})
