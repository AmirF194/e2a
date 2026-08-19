// Wraps `ServerConfiguration.makeRequestContext`, the one seam that still
// holds the raw path string before `new URL(...)` resolves a ".." segment
// away; a `Middleware.pre()` hook runs too late for that.
import type { BaseServerConfiguration } from "./generated/servers.js";
import type { HttpMethod, RequestContext } from "./generated/http/http.js";
import { E2AValidationError } from "./errors.js";

function dotSegment(endpoint: string): string | undefined {
  return endpoint.split("/").find((segment) => /^\.+$/.test(segment));
}

/** Wraps a `BaseServerConfiguration` to reject any outgoing request whose
 *  path contains a dot-only segment ("." or "..") before it can be resolved
 *  away by URL parsing. */
export function guardDotSegments(inner: BaseServerConfiguration): BaseServerConfiguration {
  return {
    makeRequestContext(endpoint: string, httpMethod: HttpMethod): RequestContext {
      const bad = dotSegment(endpoint);
      if (bad !== undefined) {
        throw new E2AValidationError({
          code: "invalid_request_path",
          message: `refusing to send a request whose path contains a dot-segment ("${bad}"): ${endpoint}`,
          status: 0,
          retryable: false,
        });
      }
      return inner.makeRequestContext(endpoint, httpMethod);
    },
  };
}
