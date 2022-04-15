package app

import (
	"github.com/digitalcircle-com-br/compass/lib/gw"
	"github.com/digitalcircle-com-br/service"
)

func Run() error {
	service.Init("compass")
	go gw.Setup()
	service.LockMainRoutine()
	return nil
}
