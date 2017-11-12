package srnd

import (
	"log"
	"os/exec"
)

type HookConfig struct {
	name   string
	exec   string
	enable bool
}

func (config *HookConfig) Exec(group, msgid, ref string) {
	if config.enable {
		cmd := exec.Command(config.exec, group, msgid, ref)
		err := cmd.Run()
		if err != nil {
			b, _ := cmd.CombinedOutput()
			log.Println("calling hook", config.name, "failed")
			log.Println(string(b))
		}
	}
}
