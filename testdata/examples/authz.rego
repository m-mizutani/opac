package authz

allow if {
    input.user == "alice"
}

allow if {
    input.role == "admin"
}
