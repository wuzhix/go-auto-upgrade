GO程序自动升级
-
```
package main

import (
	"os"
	"time"
)

import (
	"github.com/wuzhix/go-auto-upgrade"
)

type release struct {
	
}

func (*release) GetReleasePath() string {
	println("GetReleasePath")
	time.Sleep(time.Duration(3) * time.Second)
	return ""
}

func (*release) BeforeRelease() {
	println("BeforeRelease")
}

func (*release) AfterRelease(status bool, info string) {
	if status {
		println("release success")
	} else {
		println(info)
	}
}

func main()  {
	println(os.Getpid())
	go upgrade.Execute(&release{}, &upgrade.Options{})
	time.Sleep(time.Duration(100) * time.Second)
}
```