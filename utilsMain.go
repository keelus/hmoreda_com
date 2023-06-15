package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func conseguirProyectos(id string, c *gin.Context) []Proyecto {
	idioma := idiomaActual(c)

	var proyectosPorIdioma map[string][]Proyecto
	proyectosPorIdioma = make(map[string][]Proyecto)
	proyectosPorIdioma["EN"] = []Proyecto{}
	proyectosPorIdioma["ES"] = []Proyecto{}

	db := conexionDB("main")

	var proyectos []ProyectoCompleto
	err := db.Select(&proyectos, "SELECT * FROM proyectos")
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
