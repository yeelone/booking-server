package main

import (
	"booking"
	"booking/config"
	"booking/models"
	"booking/pkg/auth"
	"booking/pkg/formdata"
	"booking/pkg/token"
	v "booking/pkg/version"
	"booking/resolvers"
	"booking/util"
	"bytes"
	"context"
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/handler"
	"github.com/asdine/storm"
	"github.com/gorilla/websocket"
	"github.com/lexkong/log"
	"github.com/parnurzeal/gorequest"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"time"
)

const defaultPort = "8080"

var (
	cfg     = pflag.StringP("config", "c", "", "booking config file path.")
	version = pflag.BoolP("version", "v", false, "show version info.")
)

const DEBUG = true

type HttpResponse struct {
	Code    int               `json:"code"`
	Message string            `json:"message"`
	Data    map[string]string `json:"data"`
}

// Ticket 类型
type Ticket struct {
	Errcode   int    `json:"errcode,omitempty"`
	Errmsg    string `json:"errmsg,omitempty"`
	Ticket    string `json:"ticket,omitempty"`
	ExpiresIn int    `json:"expires_in,omitempty"`
}

// Token 类型
type Token struct {
	AccessToken string `json:"access_token,omitempty"`
	ExpiresIn   int    `json:"expires_in,omitempty"`
}

// Sign 签名类型
type Sign struct {
	AppID     string `json:"app_id,omitempty"`
	Timestamp string `json:"timestamp,omitempty"`
	NonceStr  string `json:"nonce_str,omitempty"`
	Signature string `json:"signature,omitempty"`
}

var (
	//微信公众号
	wxAppID     = "wxca621d166ede0e26"
	wxSecret    = "49aa6b3397311509bdd73ea57ded79f6"
	letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
)

func main() {
	pflag.Parse()
	if *version {
		v1 := v.Get()
		marshalled, err := json.MarshalIndent(&v1, "", "  ")
		if err != nil {
			fmt.Printf("%v\n", err)
			os.Exit(1)
		}

		fmt.Println(string(marshalled))
		return
	}

	// init config
	if err := config.Init(*cfg); err != nil {
		panic(err)
	}
	// init db
	models.DB.Init()
	defer models.DB.Close()

	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	go func() {
		for {
			GetWeixin(wxAppID, wxSecret)
			time.Sleep(time.Duration(7200) * time.Second)
		}
	}()

	e := auth.GetEnforcer("./conf/permissions/rbac_model.conf", "./conf/permissions/rbac_policy.csv")

	c := booking.Config{Resolvers: &resolvers.Resolver{}}

	c.Directives.HasRole = func(ctx context.Context, _ interface{}, next graphql.Resolver, resolver string) (interface{}, error) {

		tokenStr := ctx.Value(models.CURRENTUSERID).(string)
		secret := viper.GetString("jwt_secret")
		c, err := token.Parse(tokenStr, secret)

		if DEBUG {
			return next(ctx)
		}

		if err != nil {
			return nil, fmt.Errorf("Access denied")
		}

		sub := c.Role   // the user that wants to access a resource.
		obj := resolver // the resource that is going to be accessed.
		act := "read"   // the operation that the user performs on the resource.

		if e.Enforce(sub, obj, act) == true {
			// permit alice to read data1
			return next(ctx)
		} else {
			return nil, fmt.Errorf("auth: Access denied")
		}

		return next(ctx)
	}

	c.Directives.NeedLogin = func(ctx context.Context, obj interface{}, next graphql.Resolver, resolver string) (interface{}, error) {

		tokenStr := ctx.Value(models.CURRENTUSERID).(string)
		secret := viper.GetString("jwt_secret")
		c, err := token.Parse(tokenStr, secret)

		ctx = context.WithValue(ctx, "user_id", c.ID)

		log.Infof("resolver %-13s | role %-12s | user id %s ", resolver, c.Role, c.ID)

		if DEBUG {
			return next(ctx)
		}

		if err != nil {
			return nil, fmt.Errorf("login: Access denied")
		}

		//return nil, fmt.Errorf("Need Login ,Request Denied")
		return next(ctx)
	}

	fs := http.FileServer(http.Dir("upload/"))
	http.Handle("/", http.FileServer(http.Dir("download/wechat"))) //存放微信JS接口安全域名验证文件

	fs2 := http.FileServer(http.Dir("assets/"))
	fs3 := http.FileServer(http.Dir("download/"))

	if DEBUG {
		http.Handle("/playground", handler.Playground("GraphQL playground", "/query"))
	}

	http.Handle("/upload/", enableCORS(http.StripPrefix("/upload/", fs)))
	http.Handle("/assets/", enableCORS(http.StripPrefix("/assets/", fs2)))
	http.Handle("/download/", enableCORS(http.StripPrefix("/download/", fs3)))

	//	//http.Handle("/login", Login())
	http.Handle("/query", enableCORS(jwtMiddleware(handler.GraphQL(booking.NewExecutableSchema(c), handler.WebsocketUpgrader(websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	})))))

	http.Handle("/query/wechatToken", enableCORS(signHandler()))

	http.Handle("/upload", enableCORS(Upload()))
	log.Infof("connect to http://localhost:%s/ for GraphQL playground", port)
	log.Infof("", http.ListenAndServe(":"+port, nil))
}

func enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "*")

		next.ServeHTTP(w, r)
	})
}

func jwtMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now().UTC()

		header := r.Header.Get("Authorization")

		var t string
		// Parse the header to get the token part.
		fmt.Sscanf(header, "Bearer %s", &t)
		ctx := context.WithValue(r.Context(), models.CURRENTUSERID, t)

		ip := util.ReadUserIP(r)
		ctx = context.WithValue(ctx, models.CLIENT_IP, ip)

		// and call the next with our new context
		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)

		end := time.Now().UTC()
		latency := end.Sub(start)

		log.Infof("%-13s | %-12s ", latency, ip)
	})
}

func Upload() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			r.ParseMultipartForm(128)

			filepath, err := formdata.GetFormData(r.MultipartForm)

			resp := HttpResponse{}

			if err != nil {
				resp.Code = 501
				resp.Message = "save form data failed:" + err.Error()
			} else {
				resp.Code = 200
				resp.Message = "save form data success"
				resp.Data = make(map[string]string)
				resp.Data["filepath"] = filepath
			}

			w.Header().Set("Content-Type", "application/json")
			js, err := json.Marshal(resp)
			w.Write(js)

		}
	}
}

//signHandler 异步处理微信签名
func signHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		type urlStruct struct {
			Url string
		}

		wxURL := urlStruct{}
		result, err := ioutil.ReadAll(r.Body)
		if err != nil {
		} else {
			err := json.Unmarshal(bytes.NewBuffer(result).Bytes(), &wxURL)
			if err != nil {
				fmt.Println(err.Error())
			}
		}

		if r.Method == "POST" {
			wxNoncestr := RandStringRunes(32)
			fmt.Println("url.QueryUnescape(r.FormValue)", wxURL.Url)

			timestamp, signature := GetCanshu(wxNoncestr, wxURL.Url)

			var u = Sign{
				AppID:     wxAppID,
				Timestamp: timestamp,
				NonceStr:  wxNoncestr,
				Signature: signature,
			}
			w.Header().Add("Access-Control-Allow-Headers", "Content-Type") //header的类型
			w.Header().Set("Content-type", "application/json")             //返回数据格式是json
			b, err := json.Marshal(u)
			if err != nil {
				log.Infof(err.Error())
			}
			w.Write(b)
		} else if r.Method == "GET" {
			fmt.Println("r")
		}
	}
}

//GetCanshu 微信签名算法
func GetCanshu(noncestr, url string) (timestamp, signature string) {
	db, err := storm.Open("db/weixin.db")
	if err != nil {
		log.Infof("Database open err:", err.Error())
	}
	defer db.Close()

	defer func() { //异常处理
		if err := recover(); err != nil {
			time.Sleep(time.Duration(3) * time.Second)
		}
	}()
	var tc Ticket
	if e := db.Get("sessions", "ticket", &tc); e != nil {
		panic(e.Error())
	}

	timestamp = strconv.FormatInt(time.Now().Unix(), 10)
	longstr := "jsapi_ticket=" + tc.Ticket + "&noncestr=" + noncestr + "&timestamp=" + timestamp + "&url=" + url
	fmt.Println("longstr", longstr)
	h := sha1.New()
	if _, e := h.Write([]byte(longstr)); e != nil {
		log.Infof(e.Error())
	}

	signature = fmt.Sprintf("%x", h.Sum(nil))
	return
}

//GetWeixin 得到微信AccessToken和JSTicket
func GetWeixin(appid, secret string) {
	var tk Token
	var tc Ticket
	db, err := storm.Open("db/weixin.db")
	if err != nil {
		log.Infof("Database open err:", err.Error())
	}
	defer db.Close()

	gorequest.New().Get("https://api.weixin.qq.com/cgi-bin/token?grant_type=client_credential&appid=" + appid + "&secret=" + secret).EndStruct(&tk)
	gorequest.New().Get("https://api.weixin.qq.com/cgi-bin/ticket/getticket?access_token=" + tk.AccessToken + "&type=jsapi").EndStruct(&tc)

	if e := db.Set("sessions", "token", &tk); e != nil {
		log.Infof(e.Error())
	}
	if e := db.Set("sessions", "ticket", &tc); e != nil {
		log.Infof(e.Error())
	}
}

//RandStringRunes 生成随机字符串
func RandStringRunes(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func Login() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			decoder := json.NewDecoder(r.Body)
			var u models.User

			err := decoder.Decode(&u)
			if err != nil {
				panic(err)
			}
			// Get the user information by the login username.
			d, err := models.GetUserByName(u.Username)
			if err != nil {
				return
			}

			// Compare the login password with the user password.
			if err := auth.Compare(d.Password, u.Password); err != nil {
				return
			}

			role := ""
			if len(d.Roles) > 0 {
				role = d.Roles[0].Name
			}
			// Sign the json web token.
			t, err := token.Sign(token.Context{ID: d.ID, Username: d.Username, Role: role}, "")
			if err != nil {
				return
			}

			fmt.Println(t)

		}
	}
}
