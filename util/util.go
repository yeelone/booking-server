package util

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"github.com/teris-io/shortid"
	"io"
	"math"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"time"
)

func GenShortId() (string, error) {
	return shortid.Generate()
}

func Uint2Str(i uint64) string {
	return fmt.Sprintf("%v", i)
}

const (
	empty = ""
	tab   = "\t"
)

func PrettyJson(data interface{}) string {
	buffer := new(bytes.Buffer)
	encoder := json.NewEncoder(buffer)
	encoder.SetIndent(empty, tab)

	err := encoder.Encode(data)
	if err != nil {
		return empty
	}
	return buffer.String()
}

func Exists(path string) bool {
	_, err := os.Stat(path) //os.Stat获取文件信息
	if err != nil {
		if os.IsExist(err) {
			return true
		}
		return false
	}
	return true
}

func MoveFile(orgFile, desFile string) error {
	err := os.Rename(orgFile, desFile)
	return err
}

//ConvertToNumberingScheme : 将数字转化成ASCII字符，符合excel
func ConvertToNumberingScheme(number int) string {
	baseChar := []rune("A")[0]
	letters := ""

	for number > 0 {
		number -= 1
		letters = string(int(baseChar)+(number%26)) + letters
		number = (number / 26) >> 0 // quick `floor`
	}
	return letters
}

func ExtractFileName(filename string) (name, subffix string) {

	var filenameWithSuffix string
	filenameWithSuffix = path.Base(filename)

	var fileSuffix string
	fileSuffix = path.Ext(filenameWithSuffix) //获取文件后缀

	var filenameOnly string
	filenameOnly = strings.TrimSuffix(filenameWithSuffix, fileSuffix) //获取文件名

	return filenameOnly, fileSuffix

}

//LastMonth 根据传入的日期取得上个月
//https://play.golang.org/p/lnVudDwFYXK
func LastMonth(year, month string) (string, string) {
	y, _ := strconv.Atoi(year)
	m, _ := strconv.Atoi(month)

	if (m - 1) == 0 { //即1月份
		m = 12
		y = y - 1
	} else {
		m -= 1
	}

	y2 := strconv.Itoa(y)
	m2 := ""
	if m < 10 && m > 0 {
		m2 = "0" + strconv.Itoa(m)
	} else {
		m2 = strconv.Itoa(m)
	}

	return y2, m2
}

func Decimal(value float64) float64 {
	//value, _ = strconv.ParseFloat(fmt.Sprintf("%.2f", value), 64)
	value = math.Round(value*100) / 100
	return value
}

func ArrayToString(a []uint64, delim string) string {
	return strings.Trim(strings.Replace(fmt.Sprint(a), " ", delim, -1), "[]")
}

func Strip(str string) string {
	//先去除所有的空格
	str = strings.Replace(str, " ", "", -1)
	//同时去除换行符
	str = strings.Replace(str, "\n", "", -1)
	return str
}

// StringSliceEqualBCE 对比两个slice是否相等
func StringSliceEqualBCE(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	if (a == nil) != (b == nil) {
		return false
	}

	b = b[:len(a)]
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}

	return true
}

// Read a whole file into the memory and store it as array of lines
func ReadLines(path string) (lines []string, err error) {
	var (
		file   *os.File
		part   []byte
		prefix bool
	)
	if file, err = os.Open(path); err != nil {
		return
	}
	defer file.Close()

	reader := bufio.NewReader(file)
	buffer := bytes.NewBuffer(make([]byte, 0))
	for {
		if part, prefix, err = reader.ReadLine(); err != nil {
			break
		}
		buffer.Write(part)
		if !prefix {
			lines = append(lines, buffer.String())
			buffer.Reset()
		}
	}
	if err == io.EOF {
		err = nil
	}
	return
}

func WriteLines(lines []string, path string) (err error) {
	var (
		file *os.File
	)

	if file, err = os.Create(path); err != nil {
		return
	}
	defer file.Close()

	for _, item := range lines {
		_, err := file.WriteString(strings.TrimSpace(item) + "\n")
		if err != nil {
			fmt.Println(err)
			break
		}
	}
	return
}

func StructToMap(model interface{}) map[string]interface{} {
	var sMap map[string]interface{}
	j, _ := json.Marshal(model)
	json.Unmarshal(j, &sMap)
	return sMap
}

//FindUpdatedField : 通过对比新旧两个model来找出变化的字段
func FindUpdatedField(oldModel interface{}, newModel interface{}) (result map[string]map[string]interface{}) {
	s1 := StructToMap(oldModel)
	s2 := StructToMap(newModel)
	result = make(map[string]map[string]interface{}) // 字段 --》 旧值 新值
	for k := range s1 {
		if s1[k] != s2[k] {
			result[k] = make(map[string]interface{})
			result[k]["from"] = s1[k]
			result[k]["to"] = s2[k]
		}
	}
	return result
}

// CountDays : 根据年份月份计算天数
func CountDays(year int, month int) (days int) {
	if month != 2 {
		if month == 4 || month == 6 || month == 9 || month == 11 {
			days = 30

		} else {
			days = 31
			fmt.Fprintln(os.Stdout, "The month has 31 days")
		}
	} else {
		if ((year%4) == 0 && (year%100) != 0) || (year%400) == 0 {
			days = 29
		} else {
			days = 28
		}
	}
	fmt.Fprintf(os.Stdout, "The %d-%d has %d days.\n", year, month, days)
	return
}

// Abs returns the absolute value of x.
func Abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func ReadUserIP(r *http.Request) string {
	IPAddress := r.Header.Get("X-Forwarded-For")
	if IPAddress == "" {
		IPAddress = r.Header.Get("X-Real-Ip")
	}
	if IPAddress == "" {
		IPAddress = r.RemoteAddr
	}
	return IPAddress
}

//如果是7:00 ,要将之转化为07:00
func FillTimeFormat(timeStr string) string {
	hourMinute := strings.Split(timeStr, ":")
	hour := hourMinute[0]
	minute := hourMinute[1]

	if len(hour) == 1 {
		hour = "0" + hour
	}

	if len(minute) == 1 {
		minute = "0" + minute
	}

	//t := time.Now().Format("2006-01-02")
	return " " + hour + ":" + minute + ":00"

}

// 检查当前时间是不是在指定的区间里
// during = 7:00-19:00
func CheckTimeRange(during string) (bool, error) {
	t := strings.Split(during, "-")

	if len(t) < 2 {
		return false, errors.New("format error")
	}

	now := ChangeToMinute(time.Now().Format("15:04"))

	startTime := ChangeToMinute(t[0])
	endTime := ChangeToMinute(t[1])

	if startTime <= now && now <= endTime {
		return true, nil
	}

	return false, nil

}

// 将7：00这样的时间格式转化为分钟数
func ChangeToMinute(t string) int {
	ts := strings.Split(t, ":")

	if len(ts) < 2 {
		return -1
	}

	hour, err := strconv.Atoi(ts[0])
	if err != nil {
		return -1
	}

	minute, err := strconv.Atoi(ts[1])
	if err != nil {
		return -1
	}

	return hour*60 + minute

}
