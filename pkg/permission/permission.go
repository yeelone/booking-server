package permission

import (
	"booking/util"
	"bufio"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/axgle/mahonia"
	"github.com/spf13/viper"
)

type Resource struct {
	ID      string `json:"id"`
	Checked bool   `json:"checked"`
}

func GetPermissionFieldsFromConf(subject string, permissions map[string]Resource) map[string]map[string]Resource {

	runtimeViper := viper.New()
	runtimeViper.AddConfigPath("conf/permissions") // 如果没有指定配置文件，则解析默认的配置文件
	runtimeViper.SetConfigName("permission")

	runtimeViper.SetConfigType("yaml")                  // 设置配置文件格式为YAML
	if err := runtimeViper.ReadInConfig(); err != nil { // viper解析配置文件
		fmt.Println(err)
		return nil
	}

	keys := make(map[string]map[string]Resource)

	for _, key := range runtimeViper.AllKeys() {
		s := strings.Split(key, ".")
		if len(s) > 0 {
			if _, ok := keys[s[0]]; !ok {
				keys[s[0]] = make(map[string]Resource)
			}

			if _, ok := keys[s[0]][s[1]]; ok {
				continue
			}
			resource := Resource{}
			resource.ID = runtimeViper.GetString(s[0] + "." + s[1] + ".resource")
			str := subject + "," + runtimeViper.GetString(s[0]+"."+s[1]+".object")
			if _, ok := permissions[str]; ok {
				resource.Checked = true
				keys[s[0]][s[1]] = resource
			} else {
				resource.Checked = false
				keys[s[0]][s[1]] = resource
			}
		}
	}
	return keys
}

// map[string]map[string]bool    role => object => checked
func GetRolePermissionFromCSVFile() (permissions map[string]map[string]bool) {
	orgFilename := "conf/permissions/rbac_policy.csv"
	if !util.Exists(orgFilename) {
		return
	}
	file, err := os.Open(orgFilename)
	if err != nil {
	}
	defer file.Close()

	decoder := mahonia.NewDecoder("utf8")            // 把原来ANSI格式的文本文件里的字符，用utf8进行解码。
	reader := csv.NewReader(decoder.NewReader(file)) // 这样，最终返回的字符串就是utf-8了。（go只认utf8）

	permissions = make(map[string]map[string]bool)
	for {
		line, error := reader.Read()
		if error == io.EOF {
			break
		} else if error != nil {
		}

		role := strings.Replace(line[1], " ", "", -1)
		object := strings.Replace(line[2], " ", "", -1)

		if _, ok := permissions[role]; !ok {
			permissions[role] = make(map[string]bool)
		}

		permissions[role][object] = true
	}

	return permissions
}

func SavePermissionsToCSV(permissions map[string]map[string]bool) error {
	file, err := os.Create("conf/permissions/rbac_policy.csv")
	defer file.Close()

	if err != nil {
		return err
	}

	runtimeViper := viper.New()
	runtimeViper.AddConfigPath("conf/permissions") // 如果没有指定配置文件，则解析默认的配置文件
	runtimeViper.SetConfigName("permission")

	runtimeViper.SetConfigType("yaml")                  // 设置配置文件格式为YAML
	if err := runtimeViper.ReadInConfig(); err != nil { // viper解析配置文件
		fmt.Println(err)
		return err
	}

	writer := bufio.NewWriter(file)
	defer writer.Flush()

	for role, perm := range permissions {
		for p := range perm {
			s := "p" + ", " + role + ", " + p + ", read \n"
			writer.Write([]byte(s))
		}
	}
	return nil
}
