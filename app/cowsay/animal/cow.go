/*
 * @Descripttion:
 * @version:
 * @Author: cm.d
 * @Date: 2021-11-07 00:01:19
 * @LastEditors: cm.d
 * @LastEditTime: 2021-11-07 00:56:32
 */
package animal

type CowTemplate struct {
}

func (cowTemplate CowTemplate) Get() string {
	cowTempalte := `
    \   ^__^
     \  (oo)\_______
        (__)\       )\/\
            ||----w |
            ||     ||
`
	return cowTempalte
}
