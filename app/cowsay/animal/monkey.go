/*
 * @Descripttion:
 * @version:
 * @Author: cm.d
 * @Date: 2021-11-07 15:00:52
 * @LastEditors: cm.d
 * @LastEditTime: 2021-11-07 15:15:46
 */

package animal

type MonkeyTemplate struct {
}

func (mt MonkeyTemplate) Get() string {
	monkeyTemplate := `
   \   
    \  
     \
         .="=.
       _/.-.-.\_     _
      ( ( o o ) )    ))
       |/  "  \|    //
        \'---'/    //
        /'"""'\\  ((
       / /_,_\ \\  \\
       \_\_'__/  \  ))
       /'  /'~\   |//
      /   /    \  /
 ,--',--'\/\    /
 '-- "--'  '--'
`
	return monkeyTemplate
}
