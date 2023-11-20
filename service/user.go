package service

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	database "todolist.go/db"
)

func NewUserForm(ctx *gin.Context) {
	ctx.HTML(http.StatusOK, "new_user_form.html", gin.H{"Title": "Register user"})
}

func LoginForm(ctx *gin.Context) {
	ctx.HTML(http.StatusOK, "login.html", gin.H{"Title": "Login Service"})
}

func hash(pw string) []byte {
	const salt = "todolist.go#"
	h := sha256.New()
	h.Write([]byte(salt))
	h.Write([]byte(pw))
	return h.Sum(nil)
}

func RegisterUser(ctx *gin.Context) {
	// フォームデータの受け取り
	username := ctx.PostForm("username")
	password := ctx.PostForm("password")
	if username == "" || password == "" {
		Error(http.StatusBadRequest, "Empty parameter")(ctx)
		return
	}

	// DB 接続
	db, err := database.GetConnection()
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}

	// DB への保存
	result, err := db.Exec("INSERT INTO users(name, password) VALUES (?, ?)", username, hash(password))
	if err != nil {
		ctx.HTML(http.StatusOK, "new_user_form.html", gin.H{"Title": "Register user", "ErrorMessage": "このユーザーアカウントは作成できません。別のユーザ名を試してください。"})
		return
	}

	// 保存状態の確認
	id, _ := result.LastInsertId()
	var user database.User
	err = db.Get(&user, "SELECT id, name, password FROM users WHERE id = ?", id)
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}

	session := sessions.Default(ctx)
	session.Set(userkey, user.ID)
	session.Save()

	ctx.Redirect(http.StatusFound, "/")
}

const userkey = "user"

func Login(ctx *gin.Context) {
	username := ctx.PostForm("username")
	password := ctx.PostForm("password")

	db, err := database.GetConnection()
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}

	// ユーザの取得
	var user database.User
	err = db.Get(&user, "SELECT id, name, password FROM users WHERE deleted = 0 AND name = ?", username)
	if err != nil {
		ctx.HTML(http.StatusBadRequest, "login.html", gin.H{"Title": "Login", "Username": username, "Error": "No such user"})
		return
	}

	// パスワードの照合
	if hex.EncodeToString(user.Password) != hex.EncodeToString(hash(password)) {
		ctx.HTML(http.StatusBadRequest, "login.html", gin.H{"Title": "Login", "Username": username, "Error": "Incorrect password"})
		return
	}

	// セッションの保存
	session := sessions.Default(ctx)
	session.Set(userkey, user.ID)
	session.Save()

	ctx.Redirect(http.StatusFound, "/")
}

func Logout(ctx *gin.Context) {
	session := sessions.Default(ctx)
	session.Clear()
	session.Options(sessions.Options{MaxAge: -1})
	session.Save()
	ctx.Redirect(http.StatusFound, "/")
}

func LoginCheck(ctx *gin.Context) {
	if sessions.Default(ctx).Get(userkey) == nil {
		ctx.Redirect(http.StatusFound, "/login")
		ctx.Abort()
	} else {
		ctx.Next()
	}
}

func EditUserDataForm(ctx *gin.Context) {
	userID := sessions.Default(ctx).Get("user")

	// DB 接続
	db, err := database.GetConnection()
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}

	var user database.User
	err = db.Get(&user, "SELECT id, name, password FROM users WHERE id = ?", userID)
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}

	ctx.HTML(http.StatusOK, "form_edit_user.html", gin.H{"Title": "Edit User Data", "userName": user.Name})
}
func UpdateUserData(ctx *gin.Context) {
	userID := sessions.Default(ctx).Get("user")

	// フォームデータの受け取り
	username := ctx.PostForm("new_username")
	password := ctx.PostForm("new_password")
	if username == "" || password == "" {
		Error(http.StatusBadRequest, "Empty parameter")(ctx)
		return
	}

	// DB 接続
	db, err := database.GetConnection()
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}

	var user database.User
	err = db.Get(&user, "SELECT id, name, password FROM users WHERE id = ?", userID)
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}

	_, err = db.Exec("UPDATE users SET name=?, password=? WHERE id=?", username, hash(password), userID)
	if err != nil {
		ctx.HTML(http.StatusOK, "form_edit_user.html", gin.H{"Title": "Register user", "ErrorMessage": "このユーザー名に変更できません。別のIDを試してください。", "userName": user.Name})
		return
	}

	// 保存状態の確認
	err = db.Get(&user, "SELECT id, name, password FROM users WHERE id = ?", userID)
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}
	ctx.Redirect(http.StatusFound, "/")
}

func DeleteUserForm(ctx *gin.Context) {
	userID := sessions.Default(ctx).Get("user")

	// DB 接続
	db, err := database.GetConnection()
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}

	var user database.User
	err = db.Get(&user, "SELECT id, name, password FROM users WHERE id = ?", userID)
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}

	ctx.HTML(http.StatusOK, "form_delete_user.html", gin.H{"Title": "Delete User", "userName": user.Name})
}

func DeleteUser(ctx *gin.Context) {
	userID := sessions.Default(ctx).Get("user")

	// DB 接続
	db, err := database.GetConnection()
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}

	_, err = db.Exec("UPDATE users SET deleted=1 WHERE id=?", userID)
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}

	// 保存状態の確認
	var user database.User
	err = db.Get(&user, "SELECT id, name, password FROM users WHERE id = ?", userID)
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}

	ctx.Redirect(http.StatusFound, "/logout")
}
