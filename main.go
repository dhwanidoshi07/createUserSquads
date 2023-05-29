package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type User struct {
    ID                 int64   `json:"id"`
    FirstName          string  `json:"first_name"`
    LastName           string  `json:"last_name"`
    Email              *string `json:"email"`
    Mobile             string  `json:"mobile"`
    LastLogin          string  `json:"last_login"`
    SentOtp            string  `json:"sent_otp"`
    IsNewuser          int64   `json:"is_newuser"`
    IsSignup           int64   `json:"is_signup"`
    OtpVerified        int64   `json:"otp_verified"`
    IsWaitlisted       int64   `json:"is_waitlisted"`
    IsCustomer         int64   `json:"is_customer"`
    FirebaseID         string  `json:"firebase_id"`
    TempID             string  `json:"temp_id"`
    PasswordResetToken string  `json:"password_reset_token"`
    InnerCircle        int64   `json:"inner_circle"`
    IsActive           int64   `json:"is_active"`
    IsKyc              int64   `json:"is_kyc"`
    Gender             string  `json:"gender"`
    Dob                string  `json:"dob"`
    RoleID             int64   `json:"role_id"`
    AdminTypeID        string  `json:"admin_type_id"`
    Username           string  `json:"username"`
    UserProfileImage   string  `json:"user_profile_image"`
    Password           string  `json:"password"`
    InvitedBy          string  `json:"invited_by"`
    InvitedCount       int64   `json:"invited_count"`
    IsNotified         int64   `json:"is_notified"`
    AppVersion         string  `json:"app_version"`
    Platform           string  `json:"platform"`
    LoginAt            *string `json:"login_at"`
    CreatedAt          string  `json:"created_at"`
    UpdatedAt          string  `json:"updated_at"`
    Age                int64   `json:"age"`
    Height             string  `json:"height"`
    Weight             string  `json:"weight"`
    CustomerID         int64   `json:"customer_id"`
    OrderID            *int64  `json:"order_id"`
    ServiceID          *int64  `json:"service_id"`
    SessionType        *string `json:"session_type"`
    Type               *string `json:"type"`
    Status             *string `json:"status"`
    CategoryIDs        string  `gorm:"column:category_ids"`
    // CategoryIDs    []string `gorm:"-"`
    // CategoryIDsStr string   `gorm:"column:category_ids"`
}


type Squad struct {
	ID    uint
	Admin uint
}

type SquadResponse struct {
	Flag    int
	Message string
	SquadID uint
}

type SquadListResponse struct {
	Flag   int     `json:"flag"`
	Squads []Squad `json:"squads"`
}

type ErrorLog struct {
	ClassOrigin         string
	SubclassOrigin      string
	ReturnedSQLState    string
	MessageText         string
	MySQLErrno          uint
	ConstraintCatalog   string
	ConstraintSchema    string
	ConstraintName      string
	CatalogName         string
	SchemaName          string
	TableName           string
	ColumnName          string
	CursorName          string
	ProcedureName       string
	Data                string
}

type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data"`
	Error   string      `json:"error"`
}

type UserIDInput struct {
	UserID int `json:"user_id"`
}

func connectToDB() (*gorm.DB, error) {
	err := godotenv.Load()
	if err != nil {
		return nil, err
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		os.Getenv("DOPAMINE_DB_USER"),
		os.Getenv("DOPAMINE_DB_PASS"),
		os.Getenv("DOPAMINE_DB_HOST"),
		os.Getenv("DOPAMINE_DB_NAME"),
	)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	return db, nil
}

func createUserSquad(w http.ResponseWriter, r *http.Request) {
	
	var input UserIDInput
	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		sendErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	

	db, err := connectToDB()
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "Failed to connect to database")
		return
	}

	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			log.Printf("Recovered from panic: %v", r)
		}
	}()

	var vCheck int64
	if err := tx.Model(&User{}).Where("id = ?", input.UserID).Count(&vCheck).Error; err != nil {
		tx.Rollback()
		sendErrorResponse(w, http.StatusInternalServerError, "Failed to check user existence")
		return
	}

	if vCheck != 0 {
		newSquad := Squad{
			Admin: uint(input.UserID),
		}
		if err := tx.Create(&newSquad).Error; err != nil {
			tx.Rollback()
			sendErrorResponse(w, http.StatusInternalServerError, "Failed to create squad")
			return
		}

		var squads []Squad
		if err := tx.Find(&squads).Error; err != nil {
			tx.Rollback()
			sendErrorResponse(w, http.StatusInternalServerError, "Failed to fetch squads")
			return
		}

		squadResponse := SquadListResponse{
			Flag:   1,
			Squads: squads,
		}

		sendJSONResponse(w, http.StatusOK, squadResponse)
	} else {
		squadResponse := SquadResponse{
			Flag:    0,
			Message: "User doesn't exist",
			SquadID: 0,
		}

		sendJSONResponse(w, http.StatusOK, squadResponse)
	}

	if err := tx.Commit().Error; err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "Failed to commit transaction")
		return
	}
}

func sendJSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := APIResponse{
		Success: true,
		Data:    data,
	}

	json.NewEncoder(w).Encode(response)
}

func sendErrorResponse(w http.ResponseWriter, statusCode int, errorMessage string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := APIResponse{
		Success: false,
		Error:   errorMessage,
	}

	json.NewEncoder(w).Encode(response)
}

func main() {
	router := mux.NewRouter()

	router.HandleFunc("/squad", createUserSquad).Methods("POST")

	log.Println("Server listening on port 8080...")
	log.Fatal(http.ListenAndServe(":8080", router))
}
