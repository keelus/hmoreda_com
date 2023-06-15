package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"

	_ "modernc.org/sqlite"
)

const CUR_DOMAIN = "hmoreda.com"
const MAX_COOKIE_DUR = 3600 * 24 * 400 // 400 dias
const DEF_LANG = "EN"

func abs_path() string {
	path, _ := os.Getwd()
	return path
}

func main() {
	rMain := gin.Default()
	rBlog := gin.Default()

	cargarRutasMain(rMain)
	cargarRutasBlog(rBlog)

	_ = http.ListenAndServe(":9990",
		NewSubdomainHandler().
			On("blog.hmoreda.com:9990", rBlog).
			On("blog.localhost:9990", rBlog).
			Default(rMain),
	)
}

func conexionDB(nombreDB string) *sqlx.DB {
	rutaSQLite := abs_path() + fmt.Sprintf("/%s/databases/%s.sqlite", nombreDB, nombreDB)

	conexion, err := sqlx.Open("sqlite", rutaSQLite)
	if err != nil {
		log.Fatal(err)
	}

	return conexion
}
