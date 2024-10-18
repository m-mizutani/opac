# METADATA
# title: conflict1
# custom:
#   key: value
#   tags: ["foo", "bar"]

package metadata_test

import rego.v1

allow if {
    input.path == "/tmp/data"
}
