//
// version.go -- contains srnd version strings
//

package srnd

import "fmt"

const major_version = 3
const minor_version = 0
const patch_verson = 0
const program_name = "srnd"

var GitVersion string

func Version() string {
	return fmt.Sprintf("%s-%d.%d.%d%s", program_name, major_version, minor_version, patch_verson, GitVersion)
}
