GO程序自动升级
-
```
package main

import (
	"github.com/wuzhix/go-auto-upgrade"
	"time"
)

type release struct {
	
}

func (*release) GetReleasePath() string {
	println("GetReleasePath")
	//time.Sleep(time.Duration(3) * time.Second)
	return ""
}

func (*release) BeforeRelease() {
	println("BeforeRelease")
}

func (*release) AfterRelease() {
	println("AfterRelease")
}

func main()  {
	go upgrade.Execute(&release{}, &upgrade.Options{})
	time.Sleep(time.Duration(100) * time.Second)
}
```