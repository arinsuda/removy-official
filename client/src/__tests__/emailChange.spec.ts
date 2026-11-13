import { describe, expect, it } from "vitest"
import { EMAIL_ERROR_KEYS, EMAIL_SEND_FAILED_KEY, emailErrorKey } from "@/components/settings/emailChange"

const withCode = (code?: string, error?: string) => ({ response: { data: { code, error } } })

describe("emailErrorKey", () => {
  it("maps each server code to its own message key", () => {
    expect(emailErrorKey(withCode("EMAIL_IN_USE"))).toBe(EMAIL_ERROR_KEYS.EMAIL_IN_USE)
    expect(emailErrorKey(withCode("EMAIL_INVALID"))).toBe(EMAIL_ERROR_KEYS.EMAIL_INVALID)
    expect(emailErrorKey(withCode("EMAIL_SAME"))).toBe(EMAIL_ERROR_KEYS.EMAIL_SAME)
  })

  it("falls back to the generic message for unknown or missing codes", () => {
    expect(emailErrorKey(withCode("SOMETHING_ELSE"))).toBe(EMAIL_SEND_FAILED_KEY)
    expect(emailErrorKey(withCode(undefined, "email already in use"))).toBe(EMAIL_SEND_FAILED_KEY)
    expect(emailErrorKey(new Error("network"))).toBe(EMAIL_SEND_FAILED_KEY)
    expect(emailErrorKey(undefined)).toBe(EMAIL_SEND_FAILED_KEY)
  })
})
