package main

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-contrib/multitemplate"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"modernc.org/sqlite"
)

type Usuario struct {
	Nombre        string `db:"nombre"`
	Email         string `db:"email"`
	Avatar        string `db:"avatar"`
	Administrador string `db:"administrador"`
}

type Post struct {
	ID         string `db:"id"`
	Titulo     string `db:"titulo"`
	CoverImg   string `db:"coverImg"`
	Contenido  string `db:"contenido"`
	Fecha      string `db:"fecha"`
	Reacciones string `db:"reacciones"`
}

func renderBlog() multitemplate.Renderer {
	r := multitemplate.NewRenderer()
	r.AddFromFiles("index", "blog/templates/index.html", "blog/templates/navbar.html")
	r.AddFromFiles("login", "blog/templates/login.html", "blog/templates/navbarCompacto.html")
	r.AddFromFiles("register", "blog/templates/register.html", "blog/templates/navbarCompacto.html")
	r.AddFromFiles("profile", "blog/templates/profile.html", "blog/templates/navbar.html")
	r.AddFromFiles("poster", "blog/templates/poster.html", "blog/templates/navbarCompacto.html")
	r.AddFromFiles("posts", "blog/templates/posts.html", "blog/templates/navbar.html")
	return r
}

func cargarRutasBlog(routerBlog *gin.Engine) {
	store := cookie.NewStore([]byte("secret"))
	store.Options(sessions.Options{
		Path:     "/",
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   3600 * 24 * 30, // 30 dias
	})
	routerBlog.Use(sessions.Sessions("sesion", store))

	routerBlog.HTMLRender = renderBlog()
	routerBlog.Use(gin.Recovery())

	routerBlog.Static("/static", "blog/static")
	routerBlog.GET("/", func(c *gin.Context) {

		// iniciado := c.Query("iniciado")
		c.HTML(http.StatusOK, "index", gin.H{
			"titulo":        "Inicio Blog 🚀",
			"UsuarioActual": datosUsuarioActual(c),
		})
	})
	routerBlog.GET("/login", func(c *gin.Context) {
		session := sessions.Default(c)
		if session.Get("Nombre") != nil {
			c.Redirect(http.StatusTemporaryRedirect, "/")
		}

		c.HTML(http.StatusOK, "login", gin.H{
			"titulo": "Login Blog 🚀",
			// "valor": 5,
		})
	})
	routerBlog.GET("/register", func(c *gin.Context) {
		session := sessions.Default(c)
		if session.Get("Nombre") != nil {
			c.Redirect(http.StatusTemporaryRedirect, "/")
		}

		c.HTML(http.StatusOK, "register", gin.H{
			"titulo": "Register Blog 🚀",
			// "valor": 5,
		})
	})
	routerBlog.GET("/profile/:nombre", func(c *gin.Context) {
		c.HTML(http.StatusOK, "profile", gin.H{
			"titulo":        "Register Blog 🚀",
			"datosPerfil":   datosPublicosUsuario(c.Param("nombre")),
			"Estadisticas":  statsUsuario(c.Param("nombre")),
			"UsuarioActual": datosUsuarioActual(c),
			// "valor": 5,
		})
	})

	routerBlog.GET("/logout", func(c *gin.Context) {
		session := sessions.Default(c)
		session.Clear()
		session.Save()

		c.Redirect(http.StatusTemporaryRedirect, "/")
	})

	routerBlog.GET("/forzarLogin", func(c *gin.Context) {
		session := sessions.Default(c)
		usuario := "keelus"

		var datosUsuario Usuario

		db := conexionDB("blog")
		err := db.QueryRowx("SELECT nombre, email, avatar FROM usuarios WHERE nombre=?", usuario).StructScan(&datosUsuario)
		if err != nil {
			log.Fatal(err)
		}

		session.Set("Nombre", datosUsuario.Nombre)
		session.Set("Email", datosUsuario.Email)

		session.Save()
		c.Redirect(http.StatusTemporaryRedirect, "/")

	})

	routerBlog.GET("/verValores", func(c *gin.Context) {
		session := sessions.Default(c)
		// Retrieve values from session
		nombre := session.Get("Nombre")
		email := session.Get("Email")

		// Print the values
		c.JSON(http.StatusOK, gin.H{
			"Nombre": nombre,
			"Email":  email,
		})
	})

	routerBlog.POST("/backend/login", func(c *gin.Context) {
		usuarioIntroducido := c.PostForm("Usuario")
		contrasenyaIntroducida := c.PostForm("Contrasenya")
		contrasenyaMD5 := GetMD5Hash(contrasenyaIntroducida)

		db := conexionDB("blog")

		var existe bool
		err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM usuarios WHERE contrasenya=? AND (email=? OR nombre=?))", contrasenyaMD5, usuarioIntroducido, usuarioIntroducido).Scan(&existe)

		if err != nil {
			// Error
			c.JSON(http.StatusOK, gin.H{"status": "error"})
			return
		}

		if !existe {
			// No existe
			c.JSON(http.StatusOK, gin.H{"status": "dataError", "error": "The user does not exist, or the combination of user-password is wrong."})
			return
		}
		// Existe

		session := sessions.Default(c)
		var datosUsuario Usuario

		err = db.QueryRowx("SELECT nombre, email, avatar FROM usuarios WHERE (email=? OR nombre=?)", usuarioIntroducido, usuarioIntroducido).StructScan(&datosUsuario)
		if err != nil {
			log.Fatal(err)
		}

		session.Set("Nombre", datosUsuario.Nombre)
		session.Set("Email", datosUsuario.Email)

		session.Save()

		c.JSON(http.StatusOK, gin.H{
			"status": "OK",
		})

		// c.JSON(http.StatusOK, gin.H{
		// 	"UsuarioIntroducido":     usuarioIntroducido,
		// 	"contrasenyaIntroducida": contrasenyaIntroducida,
		// 	"contrasenyaMD5":         contrasenyaMD5,
		// })
	})

	routerBlog.POST("/backend/publishPost", func(c *gin.Context) {
		db := conexionDB("blog")

		titulo := c.PostForm("Titulo")
		if len(strings.TrimSpace(titulo)) == 0 {
			c.JSON(http.StatusOK, gin.H{
				"status": "error",
				"error": gin.H{
					"code":  0,
					"texto": "The title can't be empty.",
				},
			})
			return
		}
		link := c.PostForm("Link")
		if len(strings.TrimSpace(link)) == 0 {
			c.JSON(http.StatusOK, gin.H{
				"status": "error",
				"error": gin.H{
					"code":  1,
					"texto": "The link-id can't be empty.",
				},
			})
			return
		}
		imgCover := c.PostForm("ImgCover")
		contenido := c.PostForm("Contenido")
		if len(strings.TrimSpace(contenido)) == 0 {
			contenido = "[empty post]"
		}

		fecha := time.Now().Unix()
		_, err := db.Exec("INSERT INTO posts (id, titulo, coverImg, contenido, fecha) values (?, ?, ?, ?, ?)", link, titulo, imgCover, contenido, fecha)
		if err != nil {
			queryErr, _ := err.(*sqlite.Error)
			if queryErr.Code() == 1555 {
				c.JSON(http.StatusOK, gin.H{
					"status": "error",
					"error": gin.H{
						"code":  2,
						"texto": "A post already exists with that title-id.",
					},
				})
				return
			} else {
				c.JSON(http.StatusOK, gin.H{
					"status": "error",
					"error": gin.H{
						"code":  -1,
						"texto": err.Error(),
					},
				})
				return
			}
		}
		c.JSON(http.StatusOK, gin.H{
			"status": "OK",
		})
	})

	routerBlog.GET("/utils/avatar/:usuario", func(c *gin.Context) {
		usuario := c.Param("usuario")

		var datosUsuario Usuario
		var datosImagen []byte

		db := conexionDB("blog")
		err := db.QueryRowx("SELECT nombre, email, avatar FROM usuarios WHERE nombre=?", usuario).StructScan(&datosUsuario)
		if err != nil {
			if err == sql.ErrNoRows {
				c.Writer.WriteHeader(http.StatusNotFound)
				return
			}
			log.Fatal(err)
		}
		if datosUsuario.Nombre == "" {
			c.Writer.WriteHeader(http.StatusNotFound)
		}

		datosImagen = []byte(datosUsuario.Avatar)

		if datosUsuario.Avatar == "default" {

			contenido, err := os.Open(abs_path() + "/blog/static/imagenes/defaultImagen.png")
			if err != nil {
				fmt.Println("Error opening file:", err)
				return
			}
			defer contenido.Close()

			datosImagen, err = ioutil.ReadAll(contenido)
		}

		c.Data(http.StatusOK, "image/png", datosImagen)
	})

	routerBlog.GET("/poster", func(c *gin.Context) {
		c.HTML(http.StatusOK, "poster", gin.H{})
	})
	routerBlog.GET("/posts", func(c *gin.Context) {
		c.HTML(http.StatusOK, "posts", gin.H{
			"Posts":         conseguirPosts(),
			"UsuarioActual": datosUsuarioActual(c),
		})
	})
}
