package main

import (
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"log"
	"strconv"
	"strings"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

func GetMD5Hash(text string) string {
	hash := md5.Sum([]byte(text))
	return hex.EncodeToString(hash[:])
}

func datosPublicosUsuario(nombre string) Usuario {
	var datosUsuario Usuario

	db := conexionDB("blog")
	err := db.QueryRowx("SELECT nombre, email, avatar, administrador FROM usuarios WHERE nombre=?", nombre).StructScan(&datosUsuario)
	if err != nil {
		log.Fatal(err)
	}
	return datosUsuario
}

func statsUsuario(nombre string) map[string]int {
	var stats map[string]int
	var cantidadReacciones int
	var cantidadComentarios int
	stats = make(map[string]int)

	db := conexionDB("blog")

	strReacciones := ""
	strComentarios := ""
	_ = db.QueryRowx("SELECT reacciones, comentarios FROM usuariosSocial WHERE nombre=?", nombre).Scan(&strReacciones, &strComentarios)

	if strings.TrimSpace(strComentarios) != "" {
		sliceComentarios := strings.Split(strComentarios, ";")
		for _, comentario := range sliceComentarios {
			cantidad, _ := strconv.Atoi(strings.Split(comentario, ":")[1])
			cantidadComentarios += cantidad
		}
	}
	if strings.TrimSpace(strReacciones) != "" {
		sliceReacciones := strings.Split(strReacciones, ";")
		cantidadReacciones = len(sliceReacciones)
	}

	stats["Posts"] = 0
	stats["Reacciones"] = cantidadReacciones
	stats["Comentarios"] = cantidadComentarios

	return stats
}

func datosUsuarioActual(c *gin.Context) map[string]string {
	session := sessions.Default(c)
	datos := make(map[string]string)

	if session.Get("Nombre") != nil {
		conseguido := datosPublicosUsuario(session.Get("Nombre").(string))

		datos["Nombre"] = session.Get("Nombre").(string)
		datos["Email"] = conseguido.Email
		datos["Avatar"] = conseguido.Avatar
		datos["Administrador"] = conseguido.Administrador

	}
	return datos
}

func cifrarImagen(valorImagen string) string {
	return "data:image/png;base64," + base64.StdEncoding.EncodeToString([]byte(valorImagen))
}

func conseguirPosts() []Post {
	db := conexionDB("blog")

	var posts []Post
	err := db.Select(&posts, "SELECT * FROM posts")
	if err != nil {
		// handle the error
		log.Fatal(err)
	}

	return posts
}
