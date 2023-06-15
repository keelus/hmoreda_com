package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"html/template"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/gin-contrib/multitemplate"
	"github.com/gin-gonic/gin"
)

type ProyectoCompleto struct {
	ID            string          `db:"id"`
	NombreES      string          `db:"nombre_es"`
	NombreEN      string          `db:"nombre_en"`
	DescripcionES string          `db:"descripcion_es"`
	DescripcionEN string          `db:"descripcion_en"`
	Imagenes      SliceImagenes   `db:"imagenes"`
	Tecnologias   MapaTecnologias `db:"tecnologias"`
	Enlaces       MapaEnlaces     `db:"enlaces"`
	HTMLES        string          `db:"html_es"`
	HTMLEN        string          `db:"html_en"`
	Tipo          string          `db:"tipo"`
	Visitas       int             `db:"visitas"`
}

type Proyecto struct {
	ID          string
	Nombre      string
	Descripcion string
	Imagenes    SliceImagenes
	Tecnologias MapaTecnologias
	Enlaces     MapaEnlaces
	HTML        string
	Tipo        string
	Visitas     int
}

type SliceImagenes []string
type MapaTecnologias map[string][]string
type MapaEnlaces map[string]string

func (s *SliceImagenes) Scan(value interface{}) error {
	return json.Unmarshal([]byte(value.(string)), s)
}

func (m *MapaTecnologias) Scan(value interface{}) error {
	return json.Unmarshal([]byte(value.(string)), m)
}

func (m *MapaEnlaces) Scan(value interface{}) error {
	return json.Unmarshal([]byte(value.(string)), m)
}

func renderMain() multitemplate.Renderer {
	r := multitemplate.NewRenderer()

	// ## LANG = EN
	r.AddFromFiles("indexEN", "main/templates/EN/index.html", "main/templates/EN/navbar.html")
	r.AddFromFiles("proyectosEN", "main/templates/EN/proyectos.html", "main/templates/EN/navbar.html")
	r.AddFromFilesFuncs("proyectoEN", template.FuncMap{
		"isLast": func(index int, length int) bool {
			return index+1 == length
		},
		"replaceStr": func(variable string, search string, replace string) string {
			return strings.Replace(variable, search, replace, 1)
		},
	}, "main/templates/EN/proyecto.html", "main/templates/EN/navbar.html")
	r.AddFromFiles("contactoEN", "main/templates/EN/contacto.html", "main/templates/EN/navbar.html")
	r.AddFromFiles("errorEN", "main/templates/EN/error.html", "main/templates/EN/navbar.html")

	// ## LANG = ES
	r.AddFromFiles("indexES", "main/templates/ES/index.html", "main/templates/ES/navbar.html")
	r.AddFromFiles("proyectosES", "main/templates/ES/proyectos.html", "main/templates/ES/navbar.html")
	r.AddFromFilesFuncs("proyectoES", template.FuncMap{
		"isLast": func(index int, length int) bool {
			return index+1 == length
		},
		"replaceStr": func(variable string, search string, replace string) string {
			return strings.Replace(variable, search, replace, 1)
		},
	}, "main/templates/ES/proyecto.html", "main/templates/ES/navbar.html")
	r.AddFromFiles("contactoES", "main/templates/ES/contacto.html", "main/templates/ES/navbar.html")
	r.AddFromFiles("errorES", "main/templates/ES/error.html", "main/templates/ES/navbar.html")

	return r
}
func cargarRutasMain(routerMain *gin.Engine) {
	routerMain.HTMLRender = renderMain()
	routerMain.Use(gin.Recovery())

	routerMain.Static("/static", "main/static")

	// Error handling middleware
	routerMain.Use(func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				forzarError(http.StatusInternalServerError, c, "")
			}
		}()
		c.Next()
	})

	routerMain.NoRoute(func(c *gin.Context) {
		forzarError(http.StatusNotFound, c, "")
	})

	routerMain.GET("/", func(c *gin.Context) {
		iniciado := true
		_, err := c.Cookie("age")
		if err == http.ErrNoCookie {
			iniciado = false
		}
		var dosProyectosAleatorios [2]Proyecto
		todosProyectos := conseguirProyectos("", c)

		rand.Seed(time.Now().UnixNano())

		aleatorioPrimero := rand.Intn(len(todosProyectos))
		dosProyectosAleatorios[0] = todosProyectos[aleatorioPrimero]

		aleatorioSegundo := rand.Intn(len(todosProyectos))
		for aleatorioSegundo == aleatorioPrimero {
			aleatorioSegundo = rand.Intn(len(todosProyectos))
		}
		dosProyectosAleatorios[1] = todosProyectos[aleatorioSegundo]

		c.HTML(http.StatusOK, fmt.Sprintf("index%s", idiomaActual(c)), gin.H{
			"titulo":       "Pagina Main",
			"valor":        [3]string{},
			"Iniciado":     iniciado,
			"Activo":       "index",
			"Proyectos":    dosProyectosAleatorios,
			"IdiomaActual": idiomaActual(c),
			// "valor": 5,
		})
	})

	routerMain.GET("/proyectos", func(c *gin.Context) {
		c.HTML(http.StatusOK, fmt.Sprintf("proyectos%s", idiomaActual(c)), gin.H{
			"Proyectos":    conseguirProyectos("", c),
			"Activo":       "proyectos",
			"IdiomaActual": idiomaActual(c),
		})
	})

	routerMain.GET("/contacto", func(c *gin.Context) {
		c.HTML(http.StatusOK, fmt.Sprintf("contacto%s", idiomaActual(c)), gin.H{
			"Activo":       "contacto",
			"IdiomaActual": idiomaActual(c),
		})
	})

	routerMain.GET("/proyectos/:id", func(c *gin.Context) {
		proyectoEspecifico := conseguirProyectos(c.Param("id"), c)
		if len(proyectoEspecifico) == 0 {
			forzarError(http.StatusNotFound, c, fmt.Sprintf("Proyecto con id '%s' no encontrado.", c.Param("id")))
			return
		} else {
			var cookieProyectosVisitados []byte
			cookieProyectosVisitadosB64, err := c.Cookie("proyectosVisitados")

			if err != nil {
				if err == http.ErrNoCookie {
					fmt.Println("Cookie no existe! []")
					sliceVacio, _ := json.Marshal([]string{})
					sliceVacioB64 := base64.StdEncoding.EncodeToString(sliceVacio)
					c.SetCookie("proyectosVisitados", sliceVacioB64, MAX_COOKIE_DUR, "/", CUR_DOMAIN, false, true)

					cookieProyectosVisitadosB64, _ = c.Cookie("proyectosVisitados")
				} else {
					forzarError(505, c, "Ha ocurrido un error desconocido al extraer la cookie 'proyectosVisitados'.")
					return
				}
			}

			fmt.Println(cookieProyectosVisitadosB64)

			cookieProyectosVisitados, err = base64.StdEncoding.DecodeString(cookieProyectosVisitadosB64)
			if err != nil {
				forzarError(505, c, "Ha ocurrido un error desconocido al descifrar la cookie desde Base 64")
				return
			}

			var proyectosVisitados []string
			json.Unmarshal(cookieProyectosVisitados, &proyectosVisitados)
			fmt.Println(proyectosVisitados)
			var proyectoVisitado bool

			for _, IDproyecto := range proyectosVisitados {
				if IDproyecto == proyectoEspecifico[0].ID {
					proyectoVisitado = true
				}
			}

			if !proyectoVisitado {
				fmt.Println("Proyecto no visitado! +1 visita y SQL!")
				proyectosVisitados = append(proyectosVisitados, proyectoEspecifico[0].ID)
				cookieProyectosVisitadosNueva, err := json.Marshal(proyectosVisitados)
				if err != nil {
					forzarError(500, c, "Ha ocurrido un error al convertir el slice a un []byte de JSON.")
					return
				} else {
					cookieProyectosVisitadosNuevaB64 := base64.StdEncoding.EncodeToString(cookieProyectosVisitadosNueva)
					c.SetCookie("proyectosVisitados", cookieProyectosVisitadosNuevaB64, MAX_COOKIE_DUR, "/", CUR_DOMAIN, false, true)
				}

				db := conexionDB("main")
				_, err = db.Exec("UPDATE proyectos SET visitas = visitas + 1 WHERE id= ?", proyectoEspecifico[0].ID)
				if err != nil {
					forzarError(http.StatusInternalServerError, c, "Ha ocurrido un error al ejecutar un comando en la base de datos.")
					return
				}

				// Sumamos lo visual
				proyectoEspecifico[0].Visitas += 1
			} else {
				fmt.Println("Proyecto ya visitado :)")
				fmt.Println(cookieProyectosVisitados)
				fmt.Println(proyectosVisitados)
			}

			// set
			// 	c.SetCookie("age", "30", 3600, "/", "localhost", false, true)
			// c.String(200, "Cookie all set!")

			// get
			// 	valor, err := c.Cookie("age")
			// 	final := "The cookie value is: " + valor
			// 	if err == http.ErrNoCookie {
			// 		final = "The cookie is empty :("
			// 	}
			// 	c.String(200, final)

			// Remove
			// c.SetCookie("age", "", -1, "/", "localhost", false, true)
			// c.String(200, "The cookie has been deleted :(")

			c.HTML(http.StatusOK, fmt.Sprintf("proyecto%s", idiomaActual(c)), gin.H{
				"Activo":       "proyecto",
				"Proyecto":     proyectoEspecifico[0],
				"ProyectoHTML": template.HTML(proyectoEspecifico[0].HTML),
				"IdiomaActual": idiomaActual(c),
			})
		}
	})
}
