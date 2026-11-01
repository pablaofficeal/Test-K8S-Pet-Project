use crate::proxy::rewrite_path;

#[test]
fn plain_prefix_replaces_the_route_path() {
    assert_eq!(
        rewrite_path("/main/api/v1/profile", "/api/v1/backend/profile", "/main/api/v1/profile"),
        "/api/v1/backend/profile"
    );
}

#[test]
fn plain_prefix_keeps_the_rest_of_the_path() {
    assert_eq!(rewrite_path("/gw/a", "/up", "/gw/a/tail/more"), "/up/tail/more");
}

#[test]
fn unmatched_request_path_is_kept_whole() {
    assert_eq!(rewrite_path("/gw/a", "/up", "/other/path"), "/up/other/path");
}

#[test]
fn empty_prefix_strips_the_route_path_only() {
    assert_eq!(rewrite_path("/gw/a", "", "/gw/a/x"), "/x");
}

#[test]
fn single_parameter_is_substituted() {
    assert_eq!(
        rewrite_path("/main/api/v1/get/my/image/:key", "/api/v1/main/backend/get/my/image/:key", "/main/api/v1/get/my/image/photo.png"),
        "/api/v1/main/backend/get/my/image/photo.png"
    );
}

#[test]
fn parameter_in_the_middle_of_the_path() {
    assert_eq!(
        rewrite_path("/gw/boards/:id/node", "/todo/create/my/board/:id/node", "/gw/boards/42/node"),
        "/todo/create/my/board/42/node"
    );
}

#[test]
fn several_parameters_are_substituted() {
    assert_eq!(rewrite_path("/gw/:a/x/:b", "/up/:b/:a", "/gw/1/x/2"), "/up/2/1");
}

#[test]
fn parameter_value_is_not_decoded_or_altered() {
    assert_eq!(rewrite_path("/gw/:id", "/up/:id", "/gw/a%20b"), "/up/a%20b");
}

#[test]
fn parameter_route_with_a_plain_prefix_appends_the_whole_request_path() {
    // route_path с параметром не совпадает с запросом буквально, поэтому strip_prefix ничего не отрезает
    assert_eq!(rewrite_path("/gw/:id", "/up/fixed", "/gw/7"), "/up/fixed/gw/7");
}
