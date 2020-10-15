package api

import(
  "github.com/go-gremlin/gremlin"
)
var DB *gremlin.Client

func InitDB() {
  auth := gremlin.OptAuthUserPass("root", "root")
  var err error
	DB, err = gremlin.NewClient("IP:80/gremlin", auth)
	_, err = DB.ExecQuery(`g.V()`)
	if err != nil {
		panic(err)
	}
}
