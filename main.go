package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-contrib/multitemplate"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"

	_ "modernc.org/sqlite"
)

const CUR_DOMAIN = "localhost"
const MAX_COOKIE_DUR = 3600 * 24 * 400 // 400 dias
const DEF_LANG = "EN"

var ABS_PATH string

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

func abs_path() string {
	path, _ := os.Getwd()
	return path
}

func renderMain() multitemplate.Renderer {
	r := multitemplate.NewRenderer()
	absPath := abs_path()

	// ## LANG = EN
	r.AddFromFiles("indexEN", absPath+"/main/templates/EN/index.html", absPath+"/main/templates/EN/navbar.html")
	r.AddFromFiles("proyectosEN", absPath+"/main/templates/EN/proyectos.html", absPath+"/main/templates/EN/navbar.html")
	r.AddFromFilesFuncs("proyectoEN", template.FuncMap{
		"isLast": func(index int, length int) bool {
			return index+1 == length
		},
		"replaceStr": func(variable string, search string, replace string) string {
			return strings.Replace(variable, search, replace, 1)
		},
	}, absPath+"/main/templates/EN/proyecto.html", absPath+"/main/templates/EN/navbar.html")
	r.AddFromFiles("contactoEN", absPath+"/main/templates/EN/contacto.html", absPath+"/main/templates/EN/navbar.html")
	r.AddFromFiles("errorEN", absPath+"/main/templates/EN/error.html", absPath+"/main/templates/EN/navbar.html")

	// ## LANG = ES
	r.AddFromFiles("indexES", absPath+"/main/templates/ES/index.html", absPath+"/main/templates/ES/navbar.html")
	r.AddFromFiles("proyectosES", absPath+"/main/templates/ES/proyectos.html", absPath+"/main/templates/ES/navbar.html")
	r.AddFromFilesFuncs("proyectoES", template.FuncMap{
		"isLast": func(index int, length int) bool {
			return index+1 == length
		},
		"replaceStr": func(variable string, search string, replace string) string {
			return strings.Replace(variable, search, replace, 1)
		},
	}, absPath+"/main/templates/ES/proyecto.html", absPath+"/main/templates/ES/navbar.html")
	r.AddFromFiles("contactoES", absPath+"/main/templates/ES/contacto.html", absPath+"/main/templates/ES/navbar.html")
	r.AddFromFiles("errorES", absPath+"/main/templates/ES/error.html", absPath+"/main/templates/ES/navbar.html")

	return r
}

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

func conseguirProyectos(id string, c *gin.Context) []Proyecto {
	idioma := idiomaActual(c)

	rutaSQLite := abs_path() + "/main/databases/main.sqlite"
	var proyectosPorIdioma map[string][]Proyecto
	proyectosPorIdioma = make(map[string][]Proyecto)
	proyectosPorIdioma["EN"] = []Proyecto{}
	proyectosPorIdioma["ES"] = []Proyecto{}

	db, err := sqlx.Open("sqlite", rutaSQLite)
	if err != nil {
		log.Fatal(err)
	}

	var proyectos []ProyectoCompleto
	err = db.Select(&proyectos, "SELECT * FROM proyectos")
	if err != nil {
		// handle the error
		log.Fatal(err)
	}
	// for _, proyecto := range proyectos {
	// 	fmt.Println(proyecto)
	// }
	for _, proyecto := range proyectos {
		proyectoEN := Proyecto{ID: proyecto.ID, Nombre: proyecto.NombreEN, Descripcion: proyecto.DescripcionEN, Imagenes: proyecto.Imagenes, Tecnologias: proyecto.Tecnologias, Enlaces: proyecto.Enlaces, HTML: proyecto.HTMLEN, Tipo: proyecto.Tipo, Visitas: proyecto.Visitas}
		proyectoES := Proyecto{ID: proyecto.ID, Nombre: proyecto.NombreES, Descripcion: proyecto.DescripcionES, Imagenes: proyecto.Imagenes, Tecnologias: proyecto.Tecnologias, Enlaces: proyecto.Enlaces, HTML: proyecto.HTMLES, Tipo: proyecto.Tipo, Visitas: proyecto.Visitas}

		proyectosPorIdioma["EN"] = append(proyectosPorIdioma["EN"], proyectoEN)
		proyectosPorIdioma["ES"] = append(proyectosPorIdioma["ES"], proyectoES)
	}

	if id == "" {
		return proyectosPorIdioma[idioma] // esto por conseguirIdioma()
	} else { // Solo queremos un proyecto especifico
		for _, valorProyecto := range proyectosPorIdioma[idioma] {
			if valorProyecto.ID == id {
				return []Proyecto{valorProyecto}
			}
		}
		return []Proyecto{}
	}
}

func forzarError(codigoErrorHttp int, c *gin.Context, razonVisual string) {
	c.HTML(codigoErrorHttp, fmt.Sprintf("error%s", idiomaActual(c)), gin.H{
		"codigoError":  codigoErrorHttp,
		"textoError":   http.StatusText(codigoErrorHttp),
		"IdiomaActual": idiomaActual(c),
	})

	fmt.Println(fmt.Sprintf("[ERROR|'%s'|%s] - %s", c.FullPath(), time.Now().Format("2006-01-02"), razonVisual))
}

func idiomaActual(c *gin.Context) string {
	idioma, err := c.Cookie("CUR_LANG")
	if err != nil || !(idioma == "EN" || idioma == "ES") {
		// Idioma no valido
		c.SetCookie("CUR_LANG", DEF_LANG, MAX_COOKIE_DUR, "/", CUR_DOMAIN, false, true)
		idioma = DEF_LANG
	}

	return idioma
}

func main() {
	ABS_PATH = abs_path()

	// ####### MAIN WEBSITE #######
	rMain := gin.Default()
	rMain.HTMLRender = renderMain()
	rMain.Use(gin.Recovery())

	rMain.Static("/static", ABS_PATH+"/main/static")

	// Error handling middleware
	rMain.Use(func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				forzarError(http.StatusInternalServerError, c, "")
			}
		}()
		c.Next()
	})

	rMain.NoRoute(func(c *gin.Context) {
		forzarError(http.StatusNotFound, c, "")
	})

	rMain.GET("/", func(c *gin.Context) {
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

	rMain.GET("/proyectos", func(c *gin.Context) {
		c.HTML(http.StatusOK, fmt.Sprintf("proyectos%s", idiomaActual(c)), gin.H{
			"Proyectos":    conseguirProyectos("", c),
			"Activo":       "proyectos",
			"IdiomaActual": idiomaActual(c),
		})
	})

	rMain.GET("/contacto", func(c *gin.Context) {
		c.HTML(http.StatusOK, fmt.Sprintf("contacto%s", idiomaActual(c)), gin.H{
			"Activo":       "contacto",
			"IdiomaActual": idiomaActual(c),
		})
	})

	rMain.GET("/proyectos/:id", func(c *gin.Context) {
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
					c.SetCookie("proyectosVisitados", cookieProyectosVisitadosNuevaB64, MAX_COOKIE_DUR, "/", CUR_DOMAIN, true, true)
				}

				rutaSQLite := ABS_PATH + "/main/databases/main.sqlite"
				db, err := sqlx.Open("sqlite", rutaSQLite)
				if err != nil {
					log.Fatal(err)
				}
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
	// rMain.Run("192.168.0.70:3000")
	// rMain.Run("localhost:3000")
	rMain.Run(":9990")
}
