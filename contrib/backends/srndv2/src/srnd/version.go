//
// version.go -- contains srnd version strings
//

package srnd

import "fmt"

const major_version = 5
const minor_version = 0
const program_name = "srnd"

var GitVersion string

func Version() string {
	return fmt.Sprintf("%s-2.%d.%d%s", program_name, major_version, minor_version, GitVersion)
}
