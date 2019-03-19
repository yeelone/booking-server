package formdata

import (
	"booking/util"
	"io"
	"mime/multipart"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/lexkong/log"
)

func GetFormData(form *multipart.Form) (filepath string, err error) {

	//获取 multi-part/form中的文件数据

	for k, v := range form.File {
		for i := 0; i < len(v); i++ {

			f, _ := v[i].Open()

			//buf,_:= ioutil.ReadAll(f)
			filename, _ := renameFile(v[i].Filename)

			dir := "upload/" + k + "/"

			if !util.Exists(dir) {
				err := os.Mkdir(dir, os.ModePerm)
				if err != nil {
					log.Error("创建文件夹出现错误:", err)
					return "", err
				}
			}
			file, err := os.OpenFile(dir+filename, os.O_WRONLY|os.O_CREATE, 0666)
			if err != nil {
				log.Error("创建文件:"+dir+filename+"出现错误:", err)
				return "", err
			}
			defer f.Close()

			io.Copy(file, f)
			return dir + filename, nil
		}
	}

	return "", nil
}

func renameFile(filename string) (name, subffix string) {
	t := strconv.FormatInt(time.Now().UnixNano(), 10)

	var filenameWithSuffix string
	filenameWithSuffix = path.Base(filename)

	var fileSuffix string
	fileSuffix = path.Ext(filenameWithSuffix) //获取文件后缀

	var filenameOnly string
	filenameOnly = strings.TrimSuffix(filenameWithSuffix, fileSuffix) //获取文件名

	return filenameOnly + t + fileSuffix, fileSuffix

}
