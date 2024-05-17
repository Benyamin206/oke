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

func getAllNahkoda(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	db, err := sql.Open("mysql", "root:@tcp(127.0.0.1:3306)/service_nahkoda")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	rows, err := db.Query("SELECT id, nama, nomor_hp, jenis_kelamin FROM nahkodas")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var nahkodas []map[string]string
	for rows.Next() {
		var id, nama, nomor_hp, jenis_kelamin string
		if err := rows.Scan(&id, &nama, &nomor_hp, &jenis_kelamin); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		nahkoda := map[string]string{
			"id":            id,
			"nama":          nama,
			"nomor_hp":      nomor_hp,
			"jenis_kelamin": jenis_kelamin,
		}
		nahkodas = append(nahkodas, nahkoda)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(nahkodas)
}

func getNahkodaByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	if r.Method != http.MethodGet {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	// Parse parameter ID dari URL
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "Parameter ID is required", http.StatusBadRequest)
		return
	}

	// Buka koneksi ke database
	db, err := sql.Open("mysql", "root:@tcp(127.0.0.1:3306)/service_nahkoda")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	// Query untuk mendapatkan kapal berdasarkan ID
	row := db.QueryRow("SELECT id, nama, jenis_kelamin, nomor_hp FROM nahkodas WHERE id = ?", id)

	// Buat variabel untuk menyimpan data kapal
	var (
		nahkodaID, nama, jenis_kelamin, nomor_hp string
	)

	// Scan hasil query ke variabel nahkoda
	err = row.Scan(&nahkodaID, &nama, &jenis_kelamin, &nomor_hp)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "nahkoda not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Buat peta untuk mengirimkan sebagai respons JSON
	nahkoda := map[string]string{
		"id":            nahkodaID,
		"nama":          nama,
		"jenis_kelamin": jenis_kelamin,
		"nomor_hp":      nomor_hp,
	}

	// Set header dan kirimkan response JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(nahkoda)
}

func createNahkoda(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	// Parse data kapal dari body request
	var nahkoda map[string]string
	err := json.NewDecoder(r.Body).Decode(&nahkoda)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Buka koneksi ke database
	db, err := sql.Open("mysql", "root:@tcp(127.0.0.1:3306)/service_nahkoda")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	// Eksekusi query untuk insert data kapal baru
	result, err := db.Exec("INSERT INTO nahkodas (nama, nomor_hp, jenis_kelamin) VALUES (?, ?, ?)", nahkoda["nama"], nahkoda["nomor_hp"], nahkoda["jenis_kelamin"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Ambil ID kapal yang baru saja dibuat
	id, _ := result.LastInsertId()

	// Kirim respons dengan ID kapal yang baru saja dibuat
	w.Header().Set("Content-Type", "application/json")
	response := map[string]int64{"id": id}
	json.NewEncoder(w).Encode(response)
}

func updateNahkoda(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var input struct {
		ID           string `json:"id"`
		Nama         string `json:"nama"`
		NomorHP      string `json:"nomor_hp"`
		JenisKelamin string `json:"jenis_kelamin"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	db, err := sql.Open("mysql", "root:@tcp(127.0.0.1:3306)/service_nahkoda")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	query := "UPDATE nahkodas SET nama = ?, nomor_hp = ?, jenis_kelamin = ? WHERE id = ?"
	stmt, err := db.Prepare(query)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer stmt.Close()

	_, err = stmt.Exec(input.Nama, input.NomorHP, input.JenisKelamin, input.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Nahkoda updated successfully"))
}

func main() {
	var mux = http.NewServeMux()

	// login
	// Route untuk mendapatkan semua pengguna
	mux.HandleFunc("/get-all-nahkoda", authorize(getAllNahkoda))

	// Route untuk mendapatkan kapal berdasarkan ID
	mux.HandleFunc("/get-nahkoda-by-id", authorize(getNahkodaByID))

	mux.HandleFunc("/create-nahkoda", authorize(createNahkoda)) // Menambahkan handler untuk create-kapal

	mux.HandleFunc("/update-nahkoda", authorize(updateNahkoda))

	fmt.Println("nahkoda server running on port : 9008")

	// Jalankan server HTTP pada port 9004
	http.ListenAndServe(":9008", mux)
}
