package system.authz

allow if {
    input.user == "admin"
}

allow if {
    input.role == "developer"
}

