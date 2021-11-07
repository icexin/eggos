/*
 * @Descripttion:
 * @version:
 * @Author: cm.d
 * @Date: 2021-11-07 15:04:45
 * @LastEditors: cm.d
 * @LastEditTime: 2021-11-07 15:15:02
 */
package animal

type PigTemplate struct {
}

func (pt PigTemplate) Get() string {
	pigTemplate := `
     \
      \       
        _//| .-~~~-.
      _/oo  }       }-@
     ('')_  }       |
      '--'| { }--{  }
    	   //_/  /_/ 
  `
	return pigTemplate
}
