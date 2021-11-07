/*
 * @Descripttion:
 * @version:
 * @Author: cm.d
 * @Date: 2021-11-07 15:09:19
 * @LastEditors: cm.d
 * @LastEditTime: 2021-11-07 15:27:36
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

type DemonTemplate struct {
}

func (dt DemonTemplate) Get() string {
	demonTemplate := `
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
	return demonTemplate
}
