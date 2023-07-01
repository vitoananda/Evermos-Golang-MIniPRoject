package main

import (
	"fmt"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"
	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
	"github.com/gosimple/slug"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

var db *gorm.DB
var err error


type User struct {
	ID           int        `json:"id"`
	Nama         string     `json:"nama"`
	KataSandi    string     `json:"kata_sandi"`
	NoTelp       string     `json:"no_telp"`
	TanggalLahir time.Time  `json:"tanggal_lahir"`
	JenisKelamin string     `json:"jenis_kelamin"`
	Tentang      string     `json:"tentang"`
	Pekerjaan    string     `json:"pekerjaan"`
	Email        string     `json:"email"`
	IDProvinsi   string     `json:"id_provinsi"`
	IDKota       string     `json:"id_kota"`
	IsAdmin      bool       `json:"is_admin"`
	UpdatedAt    time.Time  `json:"updated_at"`
	CreatedAt    time.Time  `json:"created_at"`
	Toko         []Toko     `json:"toko"`
	Alamat       []Alamat   `json:"alamat"`
}

type Toko struct {
	ID        int       `json:"id"`
	IDUser    int       `json:"id_user"`
	NamaToko  string    `json:"nama_toko"`
	URLToko   string    `json:"url_toko"`
	UpdatedAt time.Time `json:"updated_at"`
	CreatedAt time.Time `json:"created_at"`
	Produk    []Produk  `json:"produk"`
}

type Alamat struct {
	ID          int       `json:"id"`
	IDUser      int       `json:"id_user"`
	JudulAlamat string    `json:"judul_alamat"`
	NamaPenerima string    `json:"nama_penerima"`
	NoTelp      string    `json:"no_telp"`
	DetailAlamat string    `json:"detail_alamat"`
	UpdatedAt   time.Time `json:"updated_at"`
	CreatedAt   time.Time `json:"created_at"`
}

type Kategori struct {
	ID        int       `json:"id"`
	NamaKategori string    `json:"nama_kategori"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Produk struct {
	ID             int       `json:"id"`
	IDToko         int       `json:"id_toko"`
	IDCategory     int       `json:"id_category"`
	NamaProduk     string    `json:"nama_produk"`
	Slug           string    `json:"slug"`
	HargaReseller  int       `json:"harga_reseller"`
	HargaKonsumen  int       `json:"harga_konsumen"`
	Stok           int       `json:"stok"`
	Deskripsi      string    `json:"deskripsi"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type Transaksi struct {
	ID              int       `json:"id"`
	IDUser          int       `json:"id_user"`
	IDAlamat        int       `json:"id_alamat"`
	HargaTotal      int       `json:"harga_total"`
	KodeInvoice     string    `json:"kode_invoice"`
	MetodePembayaran string    `json:"metode_pembayaran"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}


func authenticate(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tokenString := r.Header.Get("Authorization")
		if tokenString == "" {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{"error": "Missing authorization token"})
			return
		}

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return []byte("secretKey"), nil
		})

		if err != nil || !token.Valid {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{"error": "Invalid authorization token"})
			return
		}

		next.ServeHTTP(w, r)
	}
}


func register(w http.ResponseWriter, r *http.Request) {
	var user User
	json.NewDecoder(r.Body).Decode(&user)

	// Check if email is already used
	if isEmailUsed(user.Email) {
		response := map[string]string{"error": "Email already in use"}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"email": user.Email,
		"exp":   time.Now().Add(time.Hour * 24).Unix(),
	})

	tokenString, err := token.SignedString([]byte("secretKey"))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	db.Create(&user)
	if db.Error != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	store := Toko{
		IDUser:   user.ID,
		NamaToko: "Default Store",
		URLToko:  "https://example.com",
	}

	db.Create(&store)
	if db.Error != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"token": tokenString,
		"user":  user,
		"store": store,
	}
	json.NewEncoder(w).Encode(response)
}

func login(w http.ResponseWriter, r *http.Request) {
	var user User
	json.NewDecoder(r.Body).Decode(&user)

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"email": user.Email,
		"exp":   time.Now().Add(time.Hour * 24).Unix(),
	})

	tokenString, err := token.SignedString([]byte("secretKey"))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	response := map[string]string{"token": tokenString}
	json.NewEncoder(w).Encode(response)
}


func createStore(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID, _ := strconv.Atoi(vars["id"])

	var user User
	if err := db.First(&user, userID).Error; err != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "User not found"})
		return
	}

	store := Toko{
		IDUser:   user.ID,
		NamaToko: r.FormValue("nama_toko"),
		URLToko:  r.FormValue("url_toko"),
	}

	db.Create(&store)

	json.NewEncoder(w).Encode(store)
}

func getStore(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID, _ := strconv.Atoi(vars["id"])
	var user User
	if err := db.First(&user, userID).Error; err != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "User not found"})
		return
	}
	var store Toko
	if err := db.Where("id_user = ?", userID).First(&store).Error; err != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "Store not found"})
		return
	}

	json.NewEncoder(w).Encode(store)
}


func createAddress(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID, _ := strconv.Atoi(vars["id"])
	var user User
	if err := db.First(&user, userID).Error; err != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "User not found"})
		return
	}

	address := Alamat{
		IDUser:      user.ID,
		JudulAlamat: r.FormValue("judul_alamat"),
		NamaPenerima: r.FormValue("nama_penerima"),
		NoTelp:      r.FormValue("no_telp"),
		DetailAlamat: r.FormValue("detail_alamat"),
	}

	db.Create(&address)

	json.NewEncoder(w).Encode(address)
}

func getAddress(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID, _ := strconv.Atoi(vars["id"])
	var user User
	if err := db.First(&user, userID).Error; err != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "User not found"})
		return
	}
	var addresses []Alamat
	db.Where("id_user = ?", userID).Find(&addresses)

	json.NewEncoder(w).Encode(addresses)
}


func createCategory(w http.ResponseWriter, r *http.Request) {
	category := Kategori{
		NamaKategori: r.FormValue("nama_kategori"),
	}
	db.Create(&category)

	json.NewEncoder(w).Encode(category)
}

func getCategories(w http.ResponseWriter, r *http.Request) {
	var categories []Kategori
	db.Find(&categories)

	json.NewEncoder(w).Encode(categories)
}


func createProduct(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	storeID, _ := strconv.Atoi(vars["id"])
	var store Toko
	if err := db.First(&store, storeID).Error; err != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "Store not found"})
		return
	}
	idKategori, _ := strconv.Atoi(r.FormValue("id_category"))
	hargaReseller, _ := strconv.Atoi(r.FormValue("harga_reseller"))
	hargaKonsumen, _ := strconv.Atoi(r.FormValue("harga_konsumen"))
	stok, _ := strconv.Atoi(r.FormValue("stok"))

	product := Produk{
		IDToko:         storeID,
		IDCategory:     idKategori,
		NamaProduk:     r.FormValue("nama_produk"),
		Slug:           slug.Make(r.FormValue("nama_produk")),
		HargaReseller:  hargaReseller,
		HargaKonsumen:  hargaKonsumen,
		Stok:           stok,
		Deskripsi:      r.FormValue("deskripsi"),
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	db.Create(&product)

	json.NewEncoder(w).Encode(product)
}

func getProducts(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	storeID, _ := strconv.Atoi(vars["id"])
	var store Toko
	if err := db.First(&store, storeID).Error; err != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "Store not found"})
		return
	}
	var products []Produk
	db.Where("id_toko = ?", storeID).Find(&products)

	json.NewEncoder(w).Encode(products)
}


func createTransaction(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID, _ := strconv.Atoi(vars["id"])
	var user User
	if err := db.First(&user, userID).Error; err != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "User not found"})
		return
	}

	idAlamat, _ := strconv.Atoi(r.FormValue("id_alamat"))
	hargaTotal, _ := strconv.Atoi(r.FormValue("harga_total"))

	transaction := Transaksi{
		IDUser:          user.ID,
		IDAlamat:        idAlamat,
		HargaTotal:      hargaTotal,
		KodeInvoice:     generateInvoiceCode(),
		MetodePembayaran: r.FormValue("metode_pembayaran"),
	}

	db.Create(&transaction)

	json.NewEncoder(w).Encode(transaction)
}

func getTransactions(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID, _ := strconv.Atoi(vars["id"])
	var user User
	if err := db.First(&user, userID).Error; err != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "User not found"})
		return
	}
	var transactions []Transaksi
	db.Where("id_user = ?", userID).Find(&transactions)

	json.NewEncoder(w).Encode(transactions)
}


func isEmailUsed(email string) bool {
	var count int
	if db != nil {
		db.Model(&User{}).Where("email = ?", email).Count(&count)
	}
	return count > 0
}

func generateInvoiceCode() string {
	return ""
}


func main() {
	db, err := gorm.Open("mysql", "root:vand0507@tcp(127.0.0.1:3306)/evermos?parseTime=true")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	if err != nil {
		log.Println("Connection Failed to Open")
	} else {
		log.Println("Connection Established")
	}

	db.AutoMigrate(&User{}, &Toko{}, &Alamat{}, &Kategori{}, &Produk{}, &Transaksi{})

	router := mux.NewRouter()

	router.HandleFunc("/", homePage).Methods("GET")
	router.HandleFunc("/register", register).Methods("POST")
	router.HandleFunc("/login", login).Methods("POST")

	router.HandleFunc("/users/{id}/store", authenticate(createStore)).Methods("POST")
	router.HandleFunc("/users/{id}/store", getStore).Methods("GET")

	router.HandleFunc("/users/{id}/address", authenticate(createAddress)).Methods("POST")
	router.HandleFunc("/users/{id}/address", getAddress).Methods("GET")

	router.HandleFunc("/categories", authenticate(createCategory)).Methods("POST")
	router.HandleFunc("/categories", getCategories).Methods("GET")

	router.HandleFunc("/stores/{id}/products", authenticate(createProduct)).Methods("POST")
	router.HandleFunc("/stores/{id}/products", getProducts).Methods("GET")

	router.HandleFunc("/users/{id}/transactions", authenticate(createTransaction)).Methods("POST")
	router.HandleFunc("/users/{id}/transactions", getTransactions).Methods("GET")

	handleRequests(router)
}

func handleRequests(router *mux.Router) {
	addr := "0.0.0.0:8000"
	log.Printf("Server running on http://%s", addr)
	log.Fatal(http.ListenAndServe(addr, router))
}

func homePage(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Welcome to the homepage!")
}
