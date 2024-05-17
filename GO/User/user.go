package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"

	_ "github.com/go-sql-driver/mysql"
)

func authorize(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// authorization := r.Header.Get("Authorization")
		// if authorization != "user_tertentu" {
		// 	http.Error(w, "Unauthorized (Anda Tidak Punya Akses)", http.StatusUnauthorized)
		// 	return
		// }
		next.ServeHTTP(w, r)
	}
}

func getAllUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	db, err := sql.Open("mysql", "root:@tcp(127.0.0.1:3306)/service_user")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	rows, err := db.Query("SELECT id, username, password, email, role FROM users")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var users []map[string]string
	for rows.Next() {
		var id, username, password, email, role string
		if err := rows.Scan(&id, &username, &password, &email, &role); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		user := map[string]string{
			"id":       id,
			"username": username,
			"password": password,
			"email":    email,
		}
		users = append(users, user)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}

func addUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	// Buka koneksi ke database
	db, err := sql.Open("mysql", "root:@tcp(127.0.0.1:3306)/service_user")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	// Parse data yang diterima dari body request
	var user map[string]string
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Periksa apakah username sudah ada dalam database
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM users WHERE username = ?", user["username"]).Scan(&count)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if count > 0 {
		// Jika username sudah ada, kirimkan response bahwa user tidak ditambahkan
		response := map[string]interface{}{
			"message": "Username already exists. User not added.",
			"status":  "gagal",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		return
	}

	// Eksekusi query untuk menambahkan pengguna ke dalam database
	result, err := db.Exec("INSERT INTO users (username, password, email, role, jenis_kelamin, nomor_telepon, alamat) VALUES (?, ?, ?, ?, ?, ?, ?)",
		user["username"], user["password"], user["email"], user["role"], user["jenis_kelamin"], user["nomor_telepon"], user["alamat"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Ambil ID pengguna yang baru saja ditambahkan
	userID, err := result.LastInsertId()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Kirimkan balasan bahwa pengguna telah ditambahkan
	response := map[string]interface{}{
		"message": "User added successfully",
		"userID":  userID,
		"status":  "berhasil",
	}
	// Set header response sebagai JSON
	w.Header().Set("Content-Type", "application/json")
	// Encode hasil ke format JSON dan kirimkan sebagai response
	json.NewEncoder(w).Encode(response)
}

func checkCredentials(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Buka koneksi ke database
	db, err := sql.Open("mysql", "root:@tcp(127.0.0.1:3306)/service_user")
	if err != nil {
		// Jika terjadi kesalahan saat membuka koneksi ke database
		http.Error(w, "Service sedang bermasalah", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	// Ambil data username, password, dan role dari query parameter
	email := r.URL.Query().Get("email")
	password := r.URL.Query().Get("password")

	// Periksa apakah email dan password cocok dalam database
	var id int      // ID pengguna
	var role string // Peran pengguna
	var username string
	row := db.QueryRow("SELECT id, role, username FROM users WHERE email = ? AND password = ?", email, password)
	err = row.Scan(&id, &role, &username)
	if err != nil {
		// Jika terjadi kesalahan saat mengeksekusi query atau tidak ada baris yang cocok
		response := map[string]interface{}{
			"status":  "failed",
			"message": "email atau password salah",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		return // Menghentikan eksekusi fungsi setelah memberikan respons
	}

	// Jika cocok, kirim status kode 200 (OK) bersama dengan ID pengguna dan role dalam bentuk JSON
	w.Header().Set("Content-Type", "application/json")
	response := map[string]interface{}{
		"status":   "success",
		"id":       id,
		"role":     role,
		"username": username,
	}
	json.NewEncoder(w).Encode(response)
}

func main() {
	var mux = http.NewServeMux()

	// login
	mux.HandleFunc("/check-credentials", checkCredentials)

	// Route untuk mendapatkan semua pengguna
	mux.HandleFunc("/get-all-user", authorize(getAllUser))

	// Route untuk menambahkan pengguna baru
	mux.HandleFunc("/add-user", authorize(addUserHandler()))

	fmt.Println("user server running on port : 9004")

	// Jalankan server HTTP pada port 9004
	http.ListenAndServe(":9004", mux)
}

// addUserHandler merupakan fungsi yang memberikan handler HTTP untuk route /add-user
func addUserHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		addUser(w, r)
	}
}
