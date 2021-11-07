/*
 * @Descripttion:cowsay
 * @version:
 * @Author: cm.d
 * @Date: 2021-11-06 19:49:21
 * @LastEditors: cm.d
 * @LastEditTime: 2021-11-07 15:12:36
 */

package cowsay

import (
	"errors"
	"fmt"

	"github.com/dj456119/go-cowsay/gocowsay"

	"github.com/icexin/eggos/app"
	"github.com/icexin/eggos/app/cowsay/animal"
)

var animalTempalteMap = make(map[string]AnimalTemplate)

type AnimalTemplate interface {
	Get() string
}

func cowsay(ctx *app.Context) error {
	err := ctx.ParseFlags()
	if err != nil {
		return err
	}

	if ctx.Flag().NArg() == 0 {
		return errors.New("no input")
	}
	info := ctx.Flag().Arg(0)
	animal := ctx.Flag().Arg(1)
	animalTemplate, err := GetAnimal(animal)
	if err != nil {
		return err
	}
	fmt.Print(gocowsay.Format(animalTemplate.Get(), info))
	return nil
}

func GetAnimal(animalType string) (AnimalTemplate, error) {
	if animalTemplate, ok := animalTempalteMap[animalType]; !ok {
		return nil, errors.New("no support animal " + animalType)
	} else {
		return animalTemplate, nil
	}
}

func RegisterAnimalTemplate(animalType string, animalTemplate AnimalTemplate) {
	animalTempalteMap[animalType] = animalTemplate
}

func init() {
	RegisterAnimalTemplate("", animal.CowTemplate{})
	RegisterAnimalTemplate("cow", animal.CowTemplate{})
	RegisterAnimalTemplate("sheep", animal.SheepTemplate{})
	RegisterAnimalTemplate("deamon", animal.DaemonTemplate{})
	RegisterAnimalTemplate("pig", animal.PigTemplate{})
	RegisterAnimalTemplate("monkey", animal.MonkeyTemplate{})
	app.Register("cowsay", cowsay)
}
