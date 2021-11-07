/*
 * @Descripttion:
 * @version:
 * @Author: cm.d
 * @Date: 2021-11-07 02:52:15
 * @LastEditors: cm.d
 * @LastEditTime: 2021-11-07 02:55:45
 */
package animal

type SheepTemplate struct {
}

func (st SheepTemplate) Get() string {
	sheepTemplate := `
    \
     \  __     
       UooU\.'@@@@@@'.
       \__/(@@@@@@@@@@)
            (@@@@@@@@)
            'YY~~~~YY'
             ||    ||
 `
	return sheepTemplate
}
