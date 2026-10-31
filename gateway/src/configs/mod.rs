pub mod cors;
pub mod routes;

pub fn load_dotenv() {
    for path in ["../.env", ".env"] {
        if let Ok(text) = std::fs::read_to_string(path) {
            for raw_line in text.lines() {
                let line = raw_line
                    .trim()
                    .strip_prefix("export ")
                    .unwrap_or(raw_line.trim());
                if line.is_empty() || line.starts_with('#') {
                    continue;
                }
                let Some((key, value)) = line.split_once('=') else {
                    continue;
                };
                let key = key.trim();
                if key.is_empty() || std::env::var(key).is_ok() {
                    continue;
                }
                unsafe {
                    std::env::set_var(key, unquote(value.trim()));
                }
            }
            break;
        }
    }
}

fn unquote(value: &str) -> String {
    let bytes = value.as_bytes();
    if bytes.len() >= 2
        && ((bytes[0] == b'"' && bytes[bytes.len() - 1] == b'"')
            || (bytes[0] == b'\'' && bytes[bytes.len() - 1] == b'\''))
    {
        return value[1..value.len() - 1].to_owned();
    }
    value.to_owned()
}

pub fn jwt_public_key() -> Result<String, String> {
    match std::env::var("JWT_PUBLIC_KEY") {
        Ok(key) if !key.trim().is_empty() => Ok(key.replace("\\n", "\n")),
        _ => Err("JWT_PUBLIC_KEY".to_owned()),
    }
}