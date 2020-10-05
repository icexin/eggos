package fs

import "github.com/spf13/afero"

var builtinFiles = map[string]string{
	"/etc/resolv.conf":          `nameserver 114.114.114.114`,
	"/proc/sys/kernel/hostname": `eggos`,
}

func etcInit() {
	for name, content := range builtinFiles {
		err := afero.WriteFile(Root, name, []byte(content), 0644)
		if err != nil {
			panic(err)
		}
	}
}
