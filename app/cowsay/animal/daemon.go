/*
 * @Descripttion:
 * @version:
 * @Author: cm.d
 * @Date: 2021-11-07 15:09:19
 * @LastEditors: cm.d
 * @LastEditTime: 2021-11-07 15:11:55
 */
/*
 * @Descripttion:
 * @version:
 * @Author: cm.d
 * @Date: 2021-11-07 15:00:52
 * @LastEditors: cm.d
 * @LastEditTime: 2021-11-07 15:04:29
 */

package animal

type DaemonTemplate struct {
}

func (dt DaemonTemplate) Get() string {
	daemonTemplate := `
   \         ,        ,
    \       /(        )'
     \      \\ \\___   / |
            /- _  '-/  '
           (/\\/ \\ \\   /\\
           / /   | '    \
           O O   ) /    |
           '-^--''<     '
          (_.)  _  )   /
           '.___/'    /
             '-----' /
<----.     __ / __   \\
<----|====O)))==) \\) /====
<----'    '--' '.__,' \\
             |        |
              \\       /
        ______( (_  / \\______
      ,'  ,-----'   |        \\
      '--{__________)        \\
 `
	return daemonTemplate
}
