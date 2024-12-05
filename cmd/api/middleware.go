package main

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/golang-jwt/jwt/v4"
)

// ฟังก์ชันสำหรับการเปิดใช้งาน CORS
func (app *application) enableCORS(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:5173")

		if r.Method == "OPTIONS" {
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Access-Control-Allow-Methods", "GET,POST,PUT,PATCH,DELETE,OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, X-CSRF-Token, Authorization")
			return
		} else {
			h.ServeHTTP(w, r)
		}
	})
}

// ฟังก์ชันสำหรับการตรวจสอบการ Auth
func (app *application) authRequired(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _, err := app.auth.GetTokenFromHeaderAndVerify(w, r)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// ฟังก์ชันสำหรับการตรวจสอบ Token ด้วย JWT
func (app *application) jwtMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// ดึง token จาก Authorization header
		tokenString := r.Header.Get("Authorization")

		// ตรวจสอบว่ามีการส่ง Token มาหรือไม่
		if tokenString == "" {
			http.Error(w, "Authorization header is required", http.StatusUnauthorized)
			return
		}

		// ตรวจสอบรูปแบบของ token ว่าเป็น Bearer token หรือไม่
		if !strings.HasPrefix(tokenString, "Bearer ") {
			http.Error(w, "Invalid token format", http.StatusUnauthorized)
			return
		}

		// ตัดคำว่า "Bearer " ออก เพื่อให้เหลือเฉพาะ token
		tokenString = strings.TrimPrefix(tokenString, "Bearer ")

		// ตรวจสอบความถูกต้องของ token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// ตรวจสอบว่ามีการใช้ signing method ที่ถูกต้องหรือไม่
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			// คืนค่า secret key ที่ใช้ในการตรวจสอบ
			return []byte(os.Getenv("JWT_SECRET")), nil
		})

		if err != nil || !token.Valid {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		// หาก token ถูกต้อง ให้ส่งคำขอไปยัง handler ถัดไป
		next.ServeHTTP(w, r)
	})
}
