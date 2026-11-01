use std::sync::atomic::Ordering;

use crate::metrics::Metrics;

#[test]
fn snapshot_starts_at_zero() {
    let snapshot = Metrics::default().snapshot();
    for key in ["total_requests", "proxied_requests", "auth_failures", "upstream_errors", "rate_limited"] {
        assert_eq!(snapshot[key], 0, "{key}");
    }
}

#[test]
fn snapshot_reflects_the_counters() {
    let metrics = Metrics::default();
    metrics.total_requests.fetch_add(5, Ordering::Relaxed);
    metrics.proxied_requests.fetch_add(3, Ordering::Relaxed);
    metrics.auth_failures.fetch_add(2, Ordering::Relaxed);
    metrics.upstream_errors.fetch_add(1, Ordering::Relaxed);
    metrics.rate_limited.fetch_add(4, Ordering::Relaxed);
    let snapshot = metrics.snapshot();
    assert_eq!(snapshot["total_requests"], 5);
    assert_eq!(snapshot["proxied_requests"], 3);
    assert_eq!(snapshot["auth_failures"], 2);
    assert_eq!(snapshot["upstream_errors"], 1);
    assert_eq!(snapshot["rate_limited"], 4);
}
