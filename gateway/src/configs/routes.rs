use serde::Deserialize;
use std::collections::HashMap;

#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum RouteAuth {
    Public,
    Private,
}

#[derive(Debug, Clone, Copy, PartialEq, Eq, Deserialize)]
#[serde(rename_all = "lowercase")]
pub enum LocalHandler {
    Hello,
    Healthz,
    Readyz,
    Metrics,
    Docs,
    Openapi,
}

#[derive(Debug, Clone)]
pub struct RateLimit {
    pub limit: u32,
    pub window_secs: u64,
}

#[derive(Debug, Clone)]
pub struct Route {
    pub method: String,
    pub path: String,
    pub upstream: String,
    pub prefix: String,
    pub timeout_secs: u64,
    pub auth: RouteAuth,
    pub websocket: bool,
    pub handler: Option<LocalHandler>,
    pub rate: Option<RateLimit>,
}

#[derive(Debug, Clone)]
pub struct GatewayConfig {
    pub listen_addr: String,
    pub routes: Vec<Route>,
}

#[derive(Debug, Deserialize)]
struct FileConfig {
    #[serde(default)]
    server: ServerConfig,
    #[serde(default)]
    microservices_timeouts: HashMap<String, String>,
    #[serde(default)]
    microservices_hosts: HashMap<String, HostDef>,
    #[serde(default)]
    public_routes: Vec<RouteDef>,
    #[serde(default)]
    private_routes: Vec<RouteDef>,
    #[serde(default)]
    rate_limits: String,
}

#[derive(Debug, Deserialize)]
struct ServerConfig {
    #[serde(default = "default_port")]
    port: u16,
    #[serde(default = "default_host")]
    gateway_host: String,
    #[serde(default = "default_microservices_host")]
    microservices_host: String,
}

impl Default for ServerConfig {
    fn default() -> Self {
        Self {
            port: default_port(),
            gateway_host: default_host(),
            microservices_host: default_microservices_host(),
        }
    }
}

#[derive(Debug, Deserialize)]
struct HostDef {
    host: String,
    #[serde(default)]
    timeout: String,
    #[serde(default)]
    rate_limit: Option<RateLimitVal>,
}

#[derive(Debug, Deserialize)]
struct RouteDef {
    path: String,
    #[serde(default = "default_method")]
    method: String,
    target: String,
    #[serde(default)]
    timeout: String,
    #[serde(default)]
    prefix: String,
    #[serde(default)]
    websocket: bool,
    #[serde(default)]
    handler: Option<LocalHandler>,
    #[serde(default)]
    rate_limit: Option<RateLimitVal>,
}

#[derive(Debug, Clone, Deserialize)]
#[serde(untagged)]
enum RateLimitVal {
    Text(String),
    Rules(Vec<Rule>),
}

#[derive(Debug, Clone, Deserialize)]
struct Rule {
    #[serde(default)]
    by: String,
    #[serde(default)]
    limit: i64,
    #[serde(default)]
    window: String,
}

#[derive(Debug, Deserialize, Default)]
struct LimitsFile {
    #[serde(default)]
    policies: HashMap<String, Vec<Rule>>,
    #[serde(default)]
    microservices: HashMap<String, Vec<Rule>>,
    #[serde(default)]
    routes: Vec<LimitedRoute>,
}

#[derive(Debug, Deserialize)]
struct LimitedRoute {
    path: String,
    #[serde(default)]
    method: String,
    #[serde(default)]
    rate_limit: Option<RateLimitVal>,
}

fn default_port() -> u16 {
    8080
}

fn default_host() -> String {
    "0.0.0.0".to_owned()
}

fn default_microservices_host() -> String {
    "127.0.0.1".to_owned()
}

fn default_method() -> String {
    "ANY".to_owned()
}

fn parse_duration_secs(raw: &str) -> Result<u64, String> {
    let s = raw.trim();
    let (num, mult) = if let Some(v) = s.strip_suffix('s') {
        (v, 1)
    } else if let Some(v) = s.strip_suffix('m') {
        (v, 60)
    } else if let Some(v) = s.strip_suffix('h') {
        (v, 3600)
    } else {
        (s, 1)
    };
    num.trim()
        .parse::<u64>()
        .map(|n| n * mult)
        .map_err(|_| format!("invalid duration {s:?}"))
}

struct Ctx {
    microservices_host: String,
    timeouts: HashMap<String, String>,
    hosts: HashMap<String, ResolvedHost>,
    limits: Option<LimitsFile>,
}

struct ResolvedHost {
    host: String,
    timeout: String,
    rate_limit: Option<RateLimitVal>,
}

fn resolve_once(raw: &str, ctx: &Ctx) -> String {
    let mut out = raw.replace("${microservices_host}", &ctx.microservices_host);
    for (k, v) in &ctx.timeouts {
        out = out.replace(&format!("${{microservices_timeouts.{k}}}"), v);
    }
    for (k, h) in &ctx.hosts {
        out = out.replace(&format!("${{microservices_hosts.{k}}}"), &h.host);
        out = out.replace(&format!("${{microservices_hosts.{k}.timeout}}"), &h.timeout);
    }
    out
}

fn resolve_str(raw: &str, ctx: &Ctx) -> Result<String, String> {
    let mut out = raw.to_owned();
    for _ in 0..10 {
        let next = resolve_once(&out, ctx);
        if next == out {
            break;
        }
        out = next;
    }
    if out.contains("${") {
        return Err(format!("unresolved placeholder in {raw:?}"));
    }
    Ok(out)
}

fn target_service_name(target: &str) -> String {
    const PREFIX: &str = "${microservices_hosts.";
    if !target.starts_with(PREFIX) {
        return String::new();
    }
    let rest = target.trim_start_matches(PREFIX).trim_end_matches('}');
    rest.split('.').next().unwrap_or("").to_owned()
}

fn method_matches(want: &str, got: &str) -> bool {
    let want = want.trim().to_uppercase();
    want.is_empty() || want == "ANY" || want == "*" || want == got.trim().to_uppercase()
}

fn rules_to_limit(rules: &[Rule]) -> Result<Option<RateLimit>, String> {
    if rules.is_empty() {
        return Ok(None);
    }
    let rule = rules
        .iter()
        .find(|r| r.by.trim().eq_ignore_ascii_case("ip"))
        .or(rules.first())
        .expect("non-empty rules");
    if rule.by.trim().is_empty() {
        return Err("rate limit rule: by is required".to_owned());
    }
    if rule.limit <= 0 {
        return Err("rate limit rule: limit must be > 0".to_owned());
    }
    Ok(Some(RateLimit {
        limit: rule.limit as u32,
        window_secs: parse_duration_secs(&rule.window)?,
    }))
}

fn resolve_ref(name: &str, ctx: &Ctx) -> Result<Option<RateLimit>, String> {
    let key = name
        .trim()
        .trim_start_matches("${")
        .trim_end_matches('}')
        .trim();
    if let Some(rest) = key.strip_prefix("rate_limits.") {
        if let Some(limits) = &ctx.limits {
            if let Some(policy) = rest.strip_prefix("policies.") {
                return limits
                    .policies
                    .get(policy)
                    .map(|r| rules_to_limit(r))
                    .unwrap_or(Err(format!("unknown placeholder ${{{name}}}")));
            }
            if let Some(service) = rest.strip_prefix("microservices.") {
                return limits
                    .microservices
                    .get(service)
                    .map(|r| rules_to_limit(r))
                    .unwrap_or(Err(format!("unknown placeholder ${{{name}}}")));
            }
            if let Some(rules) = limits.microservices.get(rest) {
                return rules_to_limit(rules);
            }
            if let Some(rules) = limits.policies.get(rest) {
                return rules_to_limit(rules);
            }
        }
        return Err(format!("unknown placeholder ${{{name}}}"));
    }
    const HOSTS_PREFIX: &str = "microservices_hosts.";
    if let Some(rest) = key.strip_prefix(HOSTS_PREFIX)
        && let Some(service) = rest.strip_suffix(".rate_limit")
    {
        let host = ctx
            .hosts
            .get(service)
            .ok_or_else(|| format!("unknown placeholder ${{{name}}}"))?;
        if let Some(val) = &host.rate_limit {
            return resolve_val(val, service, ctx);
        }
        if let Some(limits) = &ctx.limits
            && let Some(rules) = limits.microservices.get(service)
        {
            return rules_to_limit(rules);
        }
        return Ok(None);
    }
    Err(format!("unresolved placeholder ${{{name}}}"))
}

fn resolve_val(val: &RateLimitVal, service: &str, ctx: &Ctx) -> Result<Option<RateLimit>, String> {
    match val {
        RateLimitVal::Rules(rules) => rules_to_limit(rules),
        RateLimitVal::Text(text) => {
            let low = text.trim().to_lowercase();
            if ["off", "false", "none", "disable", "disabled"].contains(&low.as_str()) {
                return Ok(None);
            }
            if low.starts_with("${") && low.ends_with('}') {
                return resolve_ref(text.trim(), ctx);
            }
            let _ = service;
            Err(format!("invalid rate_limit {text:?}"))
        }
    }
}

fn effective_rate(route: &RouteDef, service: &str, ctx: &Ctx) -> Result<Option<RateLimit>, String> {
    if let Some(val) = &route.rate_limit {
        return resolve_val(val, service, ctx);
    }
    if let Some(limits) = &ctx.limits {
        for item in &limits.routes {
            if item.path == route.path
                && method_matches(&item.method, &route.method)
                && let Some(val) = &item.rate_limit
            {
                return resolve_val(val, service, ctx);
            }
        }
    }
    if let Some(host) = ctx.hosts.get(service)
        && let Some(val) = &host.rate_limit
    {
        return resolve_val(val, service, ctx);
    }
    if let Some(limits) = &ctx.limits
        && let Some(rules) = limits.microservices.get(service)
    {
        return rules_to_limit(rules);
    }
    Ok(None)
}

const GATEWAY_PREFIX: &str = "/main/api/v1/";

fn build_route(
    def: &RouteDef,
    auth: RouteAuth,
    ctx: &Ctx,
    routes_path: &str,
) -> Result<Route, String> {
    if !def.path.starts_with(GATEWAY_PREFIX) {
        return Err(format!(
            "path {:?} must start with {GATEWAY_PREFIX}",
            def.path
        ));
    }
    let local = def.target == "local";
    if local && def.handler.is_none() {
        return Err(format!(
            "route {:?}: target local requires a handler",
            def.path
        ));
    }
    if !local && def.handler.is_some() {
        return Err(format!(
            "route {:?}: handler requires target local",
            def.path
        ));
    }
    let service = target_service_name(&def.target);
    let resolved = resolve_str(&def.target, ctx).map_err(|e| format!("{routes_path}: {e}"))?;
    let trimmed = resolved.trim_end_matches('/').to_owned();
    let upstream = if local || trimmed.contains("://") {
        trimmed
    } else {
        format!("http://{trimmed}")
    };
    let timeout_raw = resolve_str(&def.timeout, ctx).map_err(|e| format!("{routes_path}: {e}"))?;
    let timeout_secs = if timeout_raw.trim().is_empty() {
        0
    } else {
        parse_duration_secs(&timeout_raw)?
    };
    Ok(Route {
        method: def.method.to_uppercase(),
        path: def.path.clone(),
        upstream,
        prefix: if def.prefix.is_empty() {
            def.path.clone()
        } else {
            def.prefix.clone()
        },
        timeout_secs,
        auth,
        websocket: def.websocket,
        handler: def.handler,
        rate: effective_rate(def, &service, ctx)?,
    })
}

fn limits_path(routes_path: &str, file: &FileConfig) -> Option<String> {
    let raw = file.rate_limits.trim();
    if raw.is_empty() {
        return None;
    }
    if std::path::Path::new(raw).is_absolute() {
        return Some(raw.to_owned());
    }
    if let Some(dir) = std::path::Path::new(routes_path).parent() {
        let candidate = dir.join(raw);
        if candidate.exists() {
            return Some(candidate.to_string_lossy().into_owned());
        }
    }
    Some(raw.to_owned())
}

pub fn load(path: &str) -> Result<GatewayConfig, String> {
    let text = std::fs::read_to_string(path).map_err(|e| e.to_string())?;
    let file: FileConfig = serde_yaml::from_str(&text).map_err(|e| e.to_string())?;

    let limits = match limits_path(path, &file) {
        None => None,
        Some(p) => {
            let data = std::fs::read_to_string(&p).map_err(|e| e.to_string())?;
            Some(serde_yaml::from_str::<LimitsFile>(&data).map_err(|e| e.to_string())?)
        }
    };

    let mut hosts = HashMap::new();
    for (name, def) in &file.microservices_hosts {
        hosts.insert(
            name.clone(),
            ResolvedHost {
                host: def.host.clone(),
                timeout: def.timeout.clone(),
                rate_limit: def.rate_limit.clone(),
            },
        );
    }
    let ctx = Ctx {
        microservices_host: file.server.microservices_host.clone(),
        timeouts: file.microservices_timeouts.clone(),
        hosts,
        limits,
    };

    let mut routes = Vec::new();
    for def in &file.public_routes {
        routes.push(build_route(def, RouteAuth::Public, &ctx, path)?);
    }
    for def in &file.private_routes {
        routes.push(build_route(def, RouteAuth::Private, &ctx, path)?);
    }
    for (i, a) in routes.iter().enumerate() {
        for b in &routes[..i] {
            if a.method == b.method && a.path == b.path {
                return Err(format!("duplicate route {} {}", a.method, a.path));
            }
        }
    }

    Ok(GatewayConfig {
        listen_addr: format!("{}:{}", file.server.gateway_host, file.server.port),
        routes,
    })
}