package main

import (
	"gx1727.com/xin/framework"
	"gx1727.com/xin/module/cms"
)

func main() {
	framework.RegisterModule(cms.Module())
	framework.Run()
}
