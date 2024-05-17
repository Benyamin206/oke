package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"

	_ "github.com/go-sql-driver/mysql"
)

func getAllOrder(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	db, err := sql.Open("mysql", "root:@tcp(127.0.0.1:3306)/service_order")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	rows, err := db.Query("SELECT * FROM orders")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var jadwals []map[string]string
	for rows.Next() {
		var id, user_id, jadwal_id, qty, total_amount, status_pembayaran string
		if err := rows.Scan(&id, &user_id, &jadwal_id, &qty, &total_amount, &status_pembayaran); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		jadwal := map[string]string{
			"id":                id,
			"user_id":           user_id,
			"jadwal_id":         jadwal_id,
			"qty":               qty,
			"total_amount":      total_amount,
			"status_pembayaran": status_pembayaran,
		}
		jadwals = append(jadwals, jadwal)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(jadwals)
}

// func updateStatusPembayaran(w http.ResponseWriter, r *http.Request) {
// 	if r.Method != http.MethodPut {
// 		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
// 		return
// 	}

// 	type RequestBody struct {
// 		OrderID          string `json:"order_id"`
// 		StatusPembayaran string `json:"status_pembayaran"`
// 	}

// 	var reqBody RequestBody
// 	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
// 		http.Error(w, err.Error(), http.StatusBadRequest)
// 		return
// 	}

// 	db, err := sql.Open("mysql", "root:@tcp(127.0.0.1:3306)/service_order")
// 	if err != nil {
// 		http.Error(w, err.Error(), http.StatusInternalServerError)
// 		return
// 	}
// 	defer db.Close()

// 	stmt, err := db.Prepare("UPDATE orders SET status_pembayaran = ? WHERE id = ?")
// 	if err != nil {
// 		http.Error(w, err.Error(), http.StatusInternalServerError)
// 		return
// 	}
// 	defer stmt.Close()

// 	res, err := stmt.Exec(reqBody.StatusPembayaran, reqBody.OrderID)
// 	if err != nil {
// 		http.Error(w, err.Error(), http.StatusInternalServerError)
// 		return
// 	}

// 	rowsAffected, err := res.RowsAffected()
// 	if err != nil {
// 		http.Error(w, err.Error(), http.StatusInternalServerError)
// 		return
// 	}

// 	w.Header().Set("Content-Type", "application/json")
// 	json.NewEncoder(w).Encode(map[string]interface{}{
// 		"message":       "Status pembayaran updated successfully",
// 		"rows_affected": rowsAffected,
// 	})
// }

func createOrder(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	// Parse data kapal dari body request
	var kapal map[string]string
	err := json.NewDecoder(r.Body).Decode(&kapal)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Buka koneksi ke database
	db, err := sql.Open("mysql", "root:@tcp(127.0.0.1:3306)/service_order")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	// Eksekusi query untuk insert data kapal baru
	result, err := db.Exec("INSERT INTO orders (user_id, jadwal_id, qty, total_amount, status_pembayaran) VALUES (?, ?, ?, ?, ?)", kapal["user_id"], kapal["jadwal_id"], kapal["qty"], kapal["total_amount"], kapal["status_pembayaran"])
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

func getOrderByIDuser(w http.ResponseWriter, r *http.Request) {

	// Pastikan method yang digunakan adalah GET
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse parameter user_id dari URL
	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		http.Error(w, "Parameter user_id is required", http.StatusBadRequest)
		return
	}

	// Buka koneksi ke database
	db, err := sql.Open("mysql", "root:@tcp(127.0.0.1:3306)/service_order")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	// Query untuk mendapatkan semua order berdasarkan user_id
	rows, err := db.Query("SELECT * FROM orders WHERE user_id = ?", userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	// Buat slice untuk menyimpan semua data order
	var orders []map[string]string

	// Loop melalui setiap baris hasil query dan tambahkan ke slice orders
	for rows.Next() {
		var (
			orderID, user_id, jadwal_id, qty, total_amount, status_pembayaran string
		)
		err := rows.Scan(&orderID, &user_id, &jadwal_id, &qty, &total_amount, &status_pembayaran)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		order := map[string]string{
			"id":                orderID,
			"user_id":           user_id,
			"jadwal_id":         jadwal_id,
			"qty":               qty,
			"total_amount":      total_amount,
			"status_pembayaran": status_pembayaran,
		}
		orders = append(orders, order)
	}

	// Set header dan kirimkan response JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(orders)
}

func authorize(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
	}
}

func main() {
	var mux = http.NewServeMux()

	// login
	// Route untuk mendapatkan semua pengguna
	mux.HandleFunc("/get-all-order", authorize(getAllOrder))
	// mux.HandleFunc("/create-jadwal", authorize(createOrder)) // Menambahkan handler untuk create-kapal

	mux.HandleFunc("/get-order-by-id-user", authorize(getOrderByIDuser))

	mux.HandleFunc("/update-status-pembayaran", authorize(updateStatusPembayaran))

	mux.HandleFunc("/create-order", authorize(createOrder)) // Menambahkan handler untuk create-kapal

	fmt.Println("order server running on port : 9014")

	// Jalankan server HTTP pada port 9014
	http.ListenAndServe(":9014", mux)

}
