package database

import (
	"database/sql"
	"fmt"

	"github.com/bikedataproject/go-bike-data-lib/dbmodel"

	"github.com/lib/pq"
	// Import postgres backend for database/sql module
	_ "github.com/lib/pq"
	log "github.com/sirupsen/logrus"
)

// Database : Struct to hold the database connection
type Database struct {
	PostgresHost       string
	PostgresUser       string
	PostgresPassword   string
	PostgresPort       int64
	PostgresDb         string
	PostgresRequireSSL string
}

// getDBConnectionString : Generate connectionstring
func (db Database) getDBConnectionString() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%v", db.PostgresHost, db.PostgresPort, db.PostgresUser, db.PostgresPassword, db.PostgresDb, db.PostgresRequireSSL)
}

// checkConnection : Check if the database can be reached
func (db Database) checkConnection() bool {
	connection, err := sql.Open("postgres", db.getDBConnectionString())
	if err != nil {
		return false
	}
	err = connection.Ping()
	return err == nil
}

// Connect : Connect to Postgres
func (db Database) Connect() (err error) {
	if db.checkConnection() {
		log.Info("Database is reachable")
	} else {
		log.Fatal("Database is unreachable")
	}
	return
}

// GetUserData : Request a user token for the ID
func (db Database) GetUserData(userID string) (usr dbmodel.User, err error) {
	connection, err := sql.Open("postgres", db.getDBConnectionString())
	if err != nil {
		defer connection.Close()
		return
	}
	err = connection.QueryRow(`
	SELECT "Id", "UserIdentifier", "Provider", "ProviderUser", "AccessToken", "RefreshToken", "TokenCreationDate", "ExpiresAt", "ExpiresIn", "IsHistoryFetched"
	FROM "Users"
	WHERE "ProviderUser"=$1;
	`, userID).Scan(&usr.ID, &usr.UserIdentifier, &usr.Provider, &usr.ProviderUser, &usr.AccessToken, &usr.RefreshToken, &usr.TokenCreationDate, &usr.ExpiresAt, &usr.ExpiresIn, &usr.IsHistoryFetched)

	defer connection.Close()
	return
}

// AddUser : Create new user in the database
func (db Database) AddUser(user *dbmodel.User) (newUser dbmodel.User, err error) {
	// Connect to database
	connection, err := sql.Open("postgres", db.getDBConnectionString())
	if err != nil {
		defer connection.Close()
		err = fmt.Errorf("Could not create database connection: %v", err)
		return
	}

	query := `
	INSERT INTO "Users"
	("UserIdentifier", "Provider", "ProviderUser", "TokenCreationDate", "ExpiresAt", "ExpiresIn", "IsHistoryFetched")
	VALUES ($1, $2, $3, $4, $5, $6, $7)
	RETURNING "Id";
	`

	// Create new user & fetch ID
	response := connection.QueryRow(query, &user.UserIdentifier, &user.Provider, &user.ProviderUser, &user.TokenCreationDate, &user.ExpiresAt, &user.ExpiresIn, &user.IsHistoryFetched)

	response.Scan(&newUser.ID)
	if err != nil {
		return
	}

	newUser.ExpiresAt = user.ExpiresAt
	newUser.ExpiresIn = user.ExpiresIn
	newUser.ProviderUser = user.ProviderUser
	newUser.Provider = user.Provider
	newUser.UserIdentifier = user.UserIdentifier

	return
}

// AddContribution : Create new user contribution
func (db Database) AddContribution(contribution *dbmodel.Contribution, user *dbmodel.User) (err error) {
	// Connect to database
	connection, err := sql.Open("postgres", db.getDBConnectionString())
	if err != nil {
		defer connection.Close()
		return fmt.Errorf("Could not create database connection: %v", err)
	}

	// Write Contribution
	query := `
	INSERT INTO "Contributions"
	("UserAgent", "Distance", "TimeStampStart", "TimeStampStop", "Duration", "PointsGeom", "PointsTime")
	VALUES ($1, $2, $3, $4, $5, $6, $7)
	RETURNING "ContributionId";
	`
	response := connection.QueryRow(query, contribution.UserAgent, contribution.Distance, contribution.TimeStampStart, contribution.TimeStampStop, contribution.Duration, contribution.PointsGeom.ToWKT(), pq.Array(contribution.PointsTime))
	defer connection.Close()

	// Create contributions
	userContrib := dbmodel.UserContribution{
		UserID: user.ID,
	}
	response.Scan(&userContrib.ContributionID)

	// Write WriteUserContribution
	query = `
	INSERT INTO "UserContributions"
	("UserId", "ContributionId")
	VALUES ($1, $2);
	`
	if _, err = connection.Exec(query, userContrib.UserID, &userContrib.ContributionID); err != nil {
		defer connection.Close()
		return fmt.Errorf("Could not insert value into contributions: %s", err)
	}

	defer connection.Close()
	return
}
