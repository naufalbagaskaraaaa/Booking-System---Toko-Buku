package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	_ "github.com/lib/pq"
)

type Book struct {
	ID_Buku int    `json:"id_buku"`
	Judul   string `json:"judul"`
	Author  string `json:"author"`
	Stock   int    `json:"stock"`
}

type History struct {
	ID           int    `json:"id"`
	ID_Buku      int    `json:"id_buku"`
	JudulBuku    string `json:"judul_buku"`
	BorrowerName string `json:"borrower_name"`
	BorrowDate   string `json:"borrow_date"`
	Status       string `json:"status"`
}

func withCORS(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		handler(w, r)
	}
}

func main() {
	connStr := "user=webadmin password=MGDcoc25159 dbname=bookingSystem host=node73839-bukusukamaju.user.cloudjkt01.com port=5432 sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal("Gagal buka koneksi:", err)
	}
	defer db.Close()
	fmt.Println("Sukses connect db PAK")

	http.HandleFunc("/api/books", withCORS(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			rows, err := db.Query("SELECT id_buku, judul, author, stock FROM books ORDER BY id_buku DESC")
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			defer rows.Close()

			var books []Book
			for rows.Next() {
				var b Book
				if err := rows.Scan(&b.ID_Buku, &b.Judul, &b.Author, &b.Stock); err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				books = append(books, b)
			}
			json.NewEncoder(w).Encode(books)
			return
		}

		if r.Method == "POST" {
			var newBook Book
			if err := json.NewDecoder(r.Body).Decode(&newBook); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			var existingID int
			err := db.QueryRow(
				"SELECT id_buku FROM books WHERE LOWER(judul) = LOWER($1)",
				newBook.Judul,
			).Scan(&existingID)

			if err == nil {
				http.Error(w, "Buku dengan judul ini sudah ada di database!", http.StatusConflict)
				return
			}

			err = db.QueryRow(
				"INSERT INTO books (judul, author, stock) VALUES ($1, $2, $3) RETURNING id_buku",
				newBook.Judul, newBook.Author, newBook.Stock,
			).Scan(&newBook.ID_Buku)

			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(newBook)
			return
		}
	}))

	http.HandleFunc("/api/book", withCORS(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var request struct {
			ID_Buku      int    `json:"id_buku"`
			BorrowerName string `json:"borrower_name"`
		}
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		result, err := db.Exec(
			"UPDATE books SET stock = stock - 1 WHERE id_buku = $1 AND stock > 0",
			request.ID_Buku,
		)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		rowsAffected, _ := result.RowsAffected()
		if rowsAffected == 0 {
			http.Error(w, "Gagal meminjam: Stok habis atau buku tidak ditemukan", http.StatusBadRequest)
			return
		}

		_, err = db.Exec(
			"INSERT INTO borrow_history (id_buku, borrower_name, status) VALUES ($1, $2, 'Dipinjam')",
			request.ID_Buku, request.BorrowerName,
		)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		json.NewEncoder(w).Encode(map[string]string{"message": "Buku berhasil dipinjam!"})
	}))

	http.HandleFunc("/api/return", withCORS(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var request struct {
			ID_Buku int `json:"id_buku"`
		}
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		var activeHistoryID int
		err := db.QueryRow(
			"SELECT id FROM borrow_history WHERE id_buku = $1 AND status = 'Dipinjam' ORDER BY borrow_date ASC LIMIT 1",
			request.ID_Buku,
		).Scan(&activeHistoryID)

		if err == sql.ErrNoRows {
			http.Error(w, "Buku ini tidak sedang dipinjam!", http.StatusBadRequest)
			return
		}
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		_, err = db.Exec(
			"UPDATE books SET stock = stock + 1 WHERE id_buku = $1",
			request.ID_Buku,
		)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		_, err = db.Exec(
			"UPDATE borrow_history SET status = 'Dikembalikan' WHERE id = $1",
			activeHistoryID,
		)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		json.NewEncoder(w).Encode(map[string]string{"message": "Buku berhasil dikembalikan!"})
	}))

	http.HandleFunc("/api/book/delete", withCORS(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var request struct {
			ID_Buku int `json:"id_buku"`
		}
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		db.Exec("DELETE FROM borrow_history WHERE id_buku = $1", request.ID_Buku)

		_, err := db.Exec("DELETE FROM books WHERE id_buku = $1", request.ID_Buku)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		json.NewEncoder(w).Encode(map[string]string{"message": "Buku berhasil dihapus!"})
	}))

	http.HandleFunc("/api/book/update", withCORS(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var b Book
		if err := json.NewDecoder(r.Body).Decode(&b); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		_, err := db.Exec(
			"UPDATE books SET judul = $1, author = $2, stock = $3 WHERE id_buku = $4",
			b.Judul, b.Author, b.Stock, b.ID_Buku,
		)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		json.NewEncoder(w).Encode(map[string]string{"message": "Buku berhasil diperbarui!"})
	}))

	http.HandleFunc("/api/history", withCORS(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		query := `
			SELECT h.id, h.id_buku, b.judul, h.borrower_name, h.borrow_date, h.status 
			FROM borrow_history h 
			JOIN books b ON h.id_buku = b.id_buku 
			ORDER BY h.id DESC
		`
		rows, err := db.Query(query)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var histories []History
		for rows.Next() {
			var h History
			if err := rows.Scan(&h.ID, &h.ID_Buku, &h.JudulBuku, &h.BorrowerName, &h.BorrowDate, &h.Status); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			histories = append(histories, h)
		}
		json.NewEncoder(w).Encode(histories)
	}))

	fs := http.FileServer(http.Dir("./frontend/dist"))
	http.Handle("/", fs)
	fmt.Println("Server jalan di http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}