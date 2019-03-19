package main

import (
	"booking"
	"booking/config"
	"booking/models"
	"booking/pkg/formdata"
	"booking/pkg/auth"
	"booking/pkg/token"
	v "booking/pkg/version"
	"booking/resolvers"
	"context"
	"encoding/json"
	"fmt"
	"github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/handler"
	"github.com/gorilla/websocket"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"log"
	"net/http"
	"os"
)

const defaultPort = "8080"
const CURRENTUSERID = "CURRENTUSERID"
var (
	cfg     = pflag.StringP("config", "c", "", "booking config file path.")
	version = pflag.BoolP("version", "v", false, "show version info.")
)

const DEBUG = true

type HttpResponse struct {
	Code int `json:"code"`
	Message string `json:"message"`
	Data   map[string]string `json:"data"`
}

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

	e := auth.GetEnforcer("./conf/permissions/rbac_model.conf", "./conf/permissions/rbac_policy.csv")

	c := booking.Config{Resolvers: &resolvers.Resolver{}}
	c.Directives.HasRole = func(ctx context.Context, _ interface{}, next graphql.Resolver, resolver string)  (interface{}, error){
		if DEBUG {
			return next(ctx)
		}
		tokenStr := ctx.Value(CURRENTUSERID).(string)
		secret := viper.GetString("jwt_secret")
		c, err := token.Parse(tokenStr,secret)

		if  err != nil {
			return nil,fmt.Errorf("Access denied")
		}

		sub := c.Role          // the user that wants to access a resource.
		obj := resolver // the resource that is going to be accessed.
		act := "read"   // the operation that the user performs on the resource.

		if e.Enforce(sub, obj, act) == true {
			// permit alice to read data1
			return next(ctx)
		} else {
			return nil,fmt.Errorf("Access denied")
		}

		return next(ctx)
	}

	c.Directives.NeedLogin = func(ctx context.Context, obj interface{}, next graphql.Resolver)  (interface{}, error){
		if DEBUG {
			return next(ctx)
		}

		//return nil, fmt.Errorf("Need Login ,Request Denied")
		return next(ctx)
	}

	fs := http.FileServer(http.Dir("upload/"))

	http.Handle("/", handler.Playground("GraphQL playground", "/query"))
	http.Handle("/upload/", enableCORS(http.StripPrefix("/upload/", fs)))
	//	//http.Handle("/login", Login())
	http.Handle("/query", enableCORS(jwtMiddleware(handler.GraphQL(booking.NewExecutableSchema(c),handler.WebsocketUpgrader(websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			fmt.Println("wss")
			return true
		},
	})))))
	http.Handle("/upload", enableCORS(Upload()))
	log.Printf("connect to http://localhost:%s/ for GraphQL playground", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}


func enableCORS(next http.Handler) http.Handler{
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		next.ServeHTTP(w, r)
	})
}

func jwtMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get("Authorization")
		var t string
		// Parse the header to get the token part.
		fmt.Sscanf(header, "Bearer %s", &t)
		ctx := context.WithValue(r.Context(), CURRENTUSERID, t)

		// and call the next with our new context
		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
		log.Println("Executing jwtMiddleware again")
	})
}

func Upload() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			r.ParseMultipartForm(128)

			filepath , err := formdata.GetFormData(r.MultipartForm)

			resp := HttpResponse{}

			if err != nil {
				resp.Code = 501
				resp.Message = "save form data failed:" + err.Error()
			}else{
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