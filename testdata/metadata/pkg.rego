# METADATA
# title: my package
# custom:
#   key: value
# scope: package

package metadata_test

import rego.v1

resp contains "a" if {
    input.path == "/tmp/data"
}
