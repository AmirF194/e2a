import { describe, it, expect, vi } from "vitest";
import { guardDotSegments } from "../../src/v1/dot-segment-guard.js";
import { E2AValidationError } from "../../src/v1/errors.js";
import { HttpMethod } from "../../src/v1/generated/http/http.js";
import type { BaseServerConfiguration } from "../../src/v1/generated/servers.js";
import type { RequestContext } from "../../src/v1/generated/http/http.js";

function fakeInner() {
  const sentinel = {} as RequestContext;
  const inner: BaseServerConfiguration = {
    makeRequestContext: vi.fn(() => sentinel),
  };
  return { inner, sentinel };
}

describe("guardDotSegments", () => {
  it.each([
    ["/v1/account/suppressions/..", ".."],
    ["/v1/account/api-keys/..", ".."],
    ["/v1/agents/support%40acme.com/suppressions/..", ".."],
    ["/v1/agents/support%40acme.com/contacts/..", ".."],
    ["/v1/account/suppressions/.", "."],
    ["/v1/contacts/...", "..."],
  ])("rejects %s before it reaches the inner server config", (endpoint, bad) => {
    const { inner } = fakeInner();
    const guarded = guardDotSegments(inner);
    let caught: unknown;
    try {
      guarded.makeRequestContext(endpoint, HttpMethod.DELETE);
    } catch (e) {
      caught = e;
    }
    expect(caught).toBeInstanceOf(E2AValidationError);
    expect((caught as E2AValidationError).code).toBe("invalid_request_path");
    expect((caught as E2AValidationError).message).toContain(`"${bad}"`);
    expect(inner.makeRequestContext).not.toHaveBeenCalled();
  });

  it.each([
    "/v1/account/suppressions/person%40example.net",
    "/v1/account/api-keys/apk_1",
    "/v1/agents/support%40acme.com/suppressions/person%40example.net",
    "/v1/domains/acme.com",
    "/v1",
    "/v1/account",
  ])("passes %s straight through to the inner server config", (endpoint) => {
    const { inner, sentinel } = fakeInner();
    const guarded = guardDotSegments(inner);
    const ctx = guarded.makeRequestContext(endpoint, HttpMethod.GET);
    expect(ctx).toBe(sentinel);
    expect(inner.makeRequestContext).toHaveBeenCalledWith(endpoint, HttpMethod.GET);
  });
});
