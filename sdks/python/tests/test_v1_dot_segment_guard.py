"""Regression tests: a dot-segment path parameter must never reach httpx.

No HTTP mock responses are registered on purpose, mirroring
test_v1_client_side_validation.py: the point is that no request is sent."""

import pytest

from e2a.v1 import AsyncE2AClient, E2AClient
from e2a.v1._dot_segment_guard import reject_dot_segment_path_params
from e2a.v1.errors import E2AError, E2AValidationError

BASE = "http://test.local"


@pytest.fixture(autouse=True)
def _clear_env(monkeypatch):
    for v in ("E2A_API_KEY", "E2A_API_URL", "E2A_BASE_URL", "E2A_AGENT_EMAIL"):
        monkeypatch.delenv(v, raising=False)


def _assert_typed(err: E2AValidationError) -> None:
    assert isinstance(err, E2AError)
    assert err.code == "invalid_request_path"
    assert err.status == 0  # pre-flight: no HTTP round-trip happened
    assert err.request_id is None
    assert err.retryable is False


@pytest.mark.parametrize("bad", ["..", ".", "..."])
def test_reject_dot_segment_path_params_rejects_all_dots(bad):
    with pytest.raises(E2AValidationError) as ei:
        reject_dot_segment_path_params({"email": bad})
    _assert_typed(ei.value)
    assert bad in str(ei.value)


@pytest.mark.parametrize(
    "path_params",
    [None, {}, {"email": "person@example.net"}, {"id": "apk_1"}],
)
def test_reject_dot_segment_path_params_passes_normal_values(path_params):
    reject_dot_segment_path_params(path_params)  # must not raise


@pytest.mark.anyio
async def test_async_suppression_delete_rejects_dot_segment(httpx_mock):
    async with AsyncE2AClient(api_key="e2a_test", base_url=BASE) as c:
        with pytest.raises(E2AValidationError) as ei:
            await c.account.suppressions.delete("..")
    _assert_typed(ei.value)


@pytest.mark.anyio
async def test_async_api_key_delete_rejects_dot_segment(httpx_mock):
    async with AsyncE2AClient(api_key="e2a_test", base_url=BASE) as c:
        with pytest.raises(E2AValidationError):
            await c.account.api_keys.delete("..")


@pytest.mark.anyio
async def test_async_agent_suppression_delete_rejects_dot_segment_on_either_arg(httpx_mock):
    # The FIRST path param (the agent email) collapses onto /v1/agents just
    # as readily as the second (the suppressed address) does.
    async with AsyncE2AClient(api_key="e2a_test", base_url=BASE) as c:
        with pytest.raises(E2AValidationError):
            await c.agents.delete_suppression("..", "person@example.net")
        with pytest.raises(E2AValidationError):
            await c.agents.delete_suppression("sender@example.com", "..")


def test_sync_suppression_delete_rejects_dot_segment(httpx_mock):
    with E2AClient(api_key="e2a_test", base_url=BASE) as c:
        with pytest.raises(E2AValidationError) as ei:
            c.account.suppressions.delete("..")
    _assert_typed(ei.value)
