package authz

allow {
    input.user == "alice"
}

allow {
    input.role == "admin"
}
