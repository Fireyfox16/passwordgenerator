package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strings"
	"time"

	_ "github.com/lib/pq"
)

var conf Configuration
var db *sql.DB

type Configuration struct {
	Host           string
	Dbname         string
	User           string
	Password       string
	Port           int
	MinSpecialChar int
	MinNum         int
	MinUpperCase   int
	PasswordLength int
}

var (
	lowerCharSet   = "abcdedfghijklmnopqrst"
	upperCharSet   = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	specialCharSet = "!@#$%&*"
	numberSet      = "0123456789"
	allCharSet     = lowerCharSet + upperCharSet + specialCharSet + numberSet
)

func init() {
	file, err := os.Open("conf.json")
	if err != nil {
		log.Fatalf("Error opening conf.json: %s", err)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	err = decoder.Decode(&conf)
	if err != nil {
		log.Fatalf("Error decoding conf.json: %s", err)
	}

	connstr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", conf.Host, conf.Port, conf.User, conf.Password, conf.Dbname)
	database, err := sql.Open("postgres", connstr)
	if err != nil {
		log.Fatal(err)
	}
	db = database
}

func main() {
	var password string
	for {
		password = genPassword()
		if err := createTable(); err != nil {
			log.Fatalf("Error creating Table: %s", err)
		}
		exists, err := checkDB(password)
		if err != nil {
			log.Fatalf("Error checking password: %s", err)
		}
		if !exists {
			if err := insertPassword(password); err != nil {
				log.Fatalf("Error inserting password into DB: %s", err)
			}
			fmt.Printf("This is your new password: %s", password)
			break
		}
	}
}

func genPassword() string {
	var password strings.Builder

	rand.Seed(time.Now().UnixNano())
	minSpecialChar := conf.MinSpecialChar
	minNum := conf.MinNum
	minUpperCase := conf.MinUpperCase
	passwordLength := conf.PasswordLength

	for i := 0; i < minSpecialChar; i++ {
		random := rand.Intn(len(specialCharSet))
		password.WriteString(string(specialCharSet[random]))
	}

	for i := 0; i < minNum; i++ {
		random := rand.Intn(len(numberSet))
		password.WriteString(string(numberSet[random]))
	}

	for i := 0; i < minUpperCase; i++ {
		random := rand.Intn(len(upperCharSet))
		password.WriteString(string(upperCharSet[random]))
	}

	remainingLength := passwordLength - minSpecialChar - minNum - minUpperCase
	for i := 0; i < remainingLength; i++ {
		random := rand.Intn(len(allCharSet))
		password.WriteString(string(allCharSet[random]))
	}
	shuffle := []rune(password.String())
	rand.Shuffle(len(shuffle), func(i, j int) {
		shuffle[i], shuffle[j] = shuffle[j], shuffle[i]
	})

	return string(shuffle)
}

func createTable() error {
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS passwords (Password varchar(255));`)
	return err
}

func checkDB(password string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS( SELECT * FROM passwords WHERE Password = $1);`
	err := db.QueryRow(query, password).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

func insertPassword(password string) error {
	_, err := db.Exec(`INSERT INTO passwords (Password) VALUES ($1);`, password)
	return err
}
