package system.authz

allow {
    input.user == "admin"
}

allow {
    input.role == "developer"
}

