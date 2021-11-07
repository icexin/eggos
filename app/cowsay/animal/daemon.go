/*
 * @Descripttion:
 * @version:
 * @Author: cm.d
 * @Date: 2021-11-07 15:00:52
 * @LastEditors: cm.d
 * @LastEditTime: 2021-11-07 15:37:45
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
