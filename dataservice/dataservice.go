package dataservice

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/tebrizetayi/bitburst/model"

	// Drivers

	_ "github.com/lib/pq"
)

//IGormClient is an interface that does database operations.
type IGormClient interface {
	UpdateUser(ctx context.Context, userData model.User) (model.User, error)
	AddUser(ctx context.Context, userData model.User) (model.User, error)
	QueryUser(ctx context.Context, userID int) (model.User, error)
	DeleteObsoleteUsers(t time.Duration) error
	SetupDB(addr string)
	Check() bool
	Close()
	AddorUpdate(ctx context.Context, userData model.User)
}

type GormClient struct {
	crDB *gorm.DB
}

func (gc *GormClient) Check() bool {
	return gc.crDB != nil
}

func (gc *GormClient) Close() {
	log.Println("Closing connection to database")
	gc.crDB.Close()
}

// AddUser adds the user into the datatable
func (gc *GormClient) AddUser(ctx context.Context, userData model.User) (model.User, error) {

	if gc.crDB == nil {
		return model.User{}, fmt.Errorf("Connection to DB not established!")
	}

	result := gc.crDB.Create(&userData)
	if result.Error != nil {
		log.Fatalf("Error creating UserData: %v \n", result.Error.Error())
		return model.User{}, result.Error
	}
	log.Printf("Successfully created UserData instance %v+ \n", userData)
	return userData, nil
}

// UpdateUser updates the user
func (gc *GormClient) UpdateUser(ctx context.Context, userData model.User) (model.User, error) {

	if gc.crDB == nil {
		return model.User{}, fmt.Errorf("Connection to DB not established!")
	}
	result := gc.crDB.Save(&userData)
	if result.Error != nil {
		log.Fatalf("Error updating UserData: %v \n", result.Error.Error())
		return model.User{}, result.Error
	}

	log.Printf("Successfully updated UserData instance %v+", userData)
	return userData, nil
}

//QueryUser gets the user with given id
func (gc *GormClient) QueryUser(ctx context.Context, userId int) (model.User, error) {

	if gc.crDB == nil {
		return model.User{}, fmt.Errorf("connection to DB not established!")
	}

	usr := model.User{}
	result := gc.crDB.Where("user_id=?", userId).First(&usr)

	if result.Error == gorm.ErrRecordNotFound {
		return model.User{}, gorm.ErrRecordNotFound
	}

	if result.Error != nil {
		return usr, result.Error
	}

	return usr, nil
}

//SetupDB creates the connection object and creates the user table.
func (gc *GormClient) SetupDB(addr string) {
	log.Println("Connecting with connection string: %v", addr)
	var err error
	gc.crDB, err = gorm.Open("postgres", addr)
	if err != nil {
		panic("failed to connect database: " + err.Error())
	}
	// Migrate the schema
	gc.crDB.AutoMigrate(&model.User{})
}

//DeleteObsoleteUsers deletes the obsolete users before a given time periood.
func (gc *GormClient) DeleteObsoleteUsers(t time.Duration) error {
	calculatedTime := time.Now().Add(-t)
	log.Printf("Deleting non active users in %f seconds\n ", t.Seconds())
	db := gc.crDB.Exec("delete from users where last_seen_time<'" + calculatedTime.Format(time.RFC3339) + "'")
	return db.Error
}

func (gc *GormClient) AddorUpdate(ctx context.Context, userData model.User) {

	userData.ID = userData.ID
	userData.UserId = int(userData.ID)
	userData.ID = 0

	if userData.Online {
		userData.LastSeenTime = time.Now()
	}

	foundUser, err := gc.QueryUser(ctx, userData.UserId)
	// if the user is not in the database , it should be created
	if err == gorm.ErrRecordNotFound {
		_, err = gc.AddUser(ctx, userData)
		if err != nil {
			log.Println(err)
		}
		return
	}

	//Checking err different than ErrRecordNotFound
	if err != nil {
		log.Println(err)
		return
	}

	//Checking user id exist in the database, then online status and lastseen time should be updated according to the status
	foundUser.Online = userData.Online
	if foundUser.Online {
		foundUser.LastSeenTime = time.Now()
	}

	//Updating existing user
	_, err = gc.UpdateUser(ctx, foundUser)
	if err != nil {
		log.Println(err)
	}
}
