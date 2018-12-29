package upgrade

import (
	"crypto/md5"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path"
	"runtime"
	"time"
)

type Hook interface {
	// 获取发布包路径
	GetReleasePath() string
	// 发布前操作，保存数据，终端服务等
	BeforeRelease()
	// 发布完成后操作，还原数据，恢复服务，检测新程序是否运行正常等
	AfterRelease()
}

type Options struct {
	TargetPath string      //文件下载路径
	TargetMode os.FileMode //文件权限
	oldPath    string      //历史版本路径
	curPath    string      //当前版本路径
}

func Execute(h Hook, ops *Options) {
	time.Sleep(time.Duration(3) * time.Second)
	if h == nil {
		println("Hook must be initialized")
		return
	}
	if runtime.GOOS != "linux" {
		println("the package only support linux")
		return
	}
	ch := make(chan string)
	go func() {
		ch <- h.GetReleasePath()
	}()
	//1、阻塞获取发布版本包
	url := <-ch
	//2、检查配置项
	checkOptions(ops)
	//3、下载发布包
	ret := download(url, ops)
	if !ret {
		println("download failed")
		return
	}
	//4、发布前操作
	h.BeforeRelease()
	//5、发布
	ret = release(ops)
	println("release ", ret)
	//6、发布后操作
	if ret {
		h.AfterRelease()
		clean()
	}
}

func checkOptions(ops *Options) {
	ops.curPath = os.Args[0]
	ops.oldPath = ops.curPath + ".old"
	if ops.TargetPath == "" || ops.TargetPath == ops.curPath {
		ops.TargetPath = ops.curPath + ".new"
	}
	if ops.TargetMode == 0 {
		ops.TargetMode = 0755
	}
}

func download(url string, ops *Options) bool {
	println("download " + url)
	res, err := http.Get(url)
	if res.StatusCode != 200 {
		println("url file is not exists")
		return false
	}
	if err != nil {
		println(err.Error())
		return false
	}
	fp, err := os.OpenFile(ops.TargetPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, ops.TargetMode)
	if err != nil {
		println(err.Error())
		return false
	}
	defer fp.Close()
	count, err := io.Copy(fp, res.Body)
	if err != nil {
		println(err.Error())
		return false
	}
	if count != res.ContentLength {
		println("download ContentLength error")
		return false
	}
	//判断下载文件和当前文件是否一样
	f, err := os.Open(ops.curPath)
	if err != nil {
		println(err.Error())
		return false
	}
	defer f.Close()
	md5Cur, md5Target := md5.New(), md5.New()
	io.Copy(md5Cur, f)
	io.Copy(md5Target, fp)
	if string(md5Cur.Sum(nil)) == string(md5Target.Sum(nil)) {
		println("download file is the same as current file")
		return false
	}
	return true
}

func release(ops *Options) bool {
	println("release", ops.TargetPath)
	// 移除历史版本
	if exists(ops.oldPath) {
		println("remove oldPath")
		err := os.Remove(ops.oldPath)
		if err != nil {
			println(err.Error())
			return false
		}
	}
	// 将当前版本备份
	println("Rename curPath")
	err := os.Rename(ops.curPath, ops.oldPath)
	if err != nil {
		println(err.Error())
		return false
	}
	// 将下载版本替换为当前版本
	println("Rename TargetPath")
	err = os.Rename(ops.TargetPath, ops.curPath)
	if err != nil {
		println(err.Error())
		// 替换失败，回滚文件
		err = os.Rename(ops.oldPath, ops.curPath)
		if err != nil {
			println("rollback fail")
		}
		return false
	}
	// 启动新版本
	println("exec Command ", path.Base(ops.curPath))
	cmd := exec.Command("bash", "-c", `./`+path.Base(ops.curPath))
	err = cmd.Start()
	if err != nil {
		// 替换失败，回滚文件
		err = os.Rename(ops.oldPath, ops.curPath)
		if err != nil {
			println("rollback fail")
		}
		return false
	}
	return true
}

func clean() {
	os.Exit(1)
}

func exists(path string) bool {
	_, err := os.Stat(path) //os.Stat获取文件信息
	if err != nil {
		if os.IsExist(err) {
			return true
		}
		return false
	}
	return true
}
