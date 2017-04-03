package srnd

import (
	"log"
	"os/exec"
)

func ExecHook(config *HookConfig, group, msgid, ref string) {
	cmd := exec.Command(config.exec, group, msgid, ref)
	err := cmd.Run()
	if err != nil {
		b, _ := cmd.CombinedOutput()
		log.Println("calling hook", config.name, "failed")
		log.Println(string(b))
	}
}
