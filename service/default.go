package service

import (
	"fmt"
	"github.com/gin-contrib/sessions"
	"net/http"
	database "todolist.go/db"

	"github.com/gin-gonic/gin"
)

// Home renders index.html
func Home(ctx *gin.Context) {
	// Get DB connection
	db, err := database.GetConnection()
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}

	userID := sessions.Default(ctx).Get("user")

	var user database.User
	err = db.Get(&user, "SELECT id, name, password FROM users WHERE id = ?", userID)

	ctx.HTML(http.StatusOK, "index.html", gin.H{"Title": "HOME", "userName": user.Name})
}

// NotImplemented renders error.html with 501 Not Implemented
func NotImplemented(ctx *gin.Context) {
	msg := fmt.Sprintf("%s access to %s is not implemented yet", ctx.Request.Method, ctx.Request.URL)
	ctx.Header("Cache-Contrl", "no-cache")
	Error(http.StatusNotImplemented, msg)(ctx)
}

// Error returns a handler which renders error.html
func Error(code int, message string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.HTML(code, "error.html", gin.H{"Code": code, "Error": message})
	}
}
