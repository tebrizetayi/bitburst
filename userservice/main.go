package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/tebrizetayi/bitburst/dataservice"
	"github.com/tebrizetayi/bitburst/model"
)

const ()

//BitburstConfig is config structure.
type BitburstConfig struct {
	//Timeperiod  for deleting users in the system
	DeletionTimeSpan time.Duration
	//Context timeout period in the timeout handler.
	TimeoutTimeSpan time.Duration
	//Database connection string
	ConnString string
	//Event queue for go routines
	EventQueueLength int
	//User online status URL
	UserStatusUrl string
}

//Users is fullfilled with Ids.
type Users struct {
	Object_ids []int `json:"object_ids"`
}

var (
	//IGormClient is interface that does database operations.
	IGormClient *dataservice.GormClient
	//Config is a object to set the configurations
	Config BitburstConfig
	//OnlineIds End list of online users
	OnlineIds []int

	secure sync.Mutex
)

//OnlineUsersHandler is handler for  getting users status from external api.
type OnlineUsersHandler struct {
}

func (h OnlineUsersHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	//Reads the requested data.
	body, _ := ioutil.ReadAll(r.Body)
	users := Users{}
	err := json.Unmarshal(body, &users)

	if err != nil {
		fmt.Println(err)
		return
	}

	//chId and chUser are buffered channels for passing data between GetUserStatus and IGormClient.AddorUpdate
	//The buffered channel here is used as create kind of go routines.If chID channel is full it , goroutine will be created after there will be
	//an empty place in buffer.
	chID := make(chan int, Config.EventQueueLength)
	chUser := make(chan model.User, Config.EventQueueLength)

	OnlineIds = []int{}

	for _, v := range users.Object_ids {
		chID <- v
		go GetUserStatus(ctx, <-chID, chUser)
		usr := <-chUser
		go IGormClient.AddorUpdate(ctx, usr)

		//if user is online then the id of a user can be added into the output.
		//Before adding user id , locking is important because of the goroutines.
		secure.Lock()
		if usr.Online {
			OnlineIds = append(OnlineIds, int(usr.ID))
		}
		secure.Unlock()

	}
	close(chID)
	close(chUser)
}

//TimeoutMiddleware is adds timeout context to request.
func timeOutMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//Channel for getting info whether the request  is responded before given timeout period
		done := make(chan bool)

		//Creating context with timeout
		ctx, cancelFunc := context.WithTimeout(r.Context(), Config.TimeoutTimeSpan)
		defer cancelFunc()

		go func() {
			r = r.WithContext(ctx)
			next.ServeHTTP(w, r)
			close(done)
		}()

		encoder := json.NewEncoder(w)
		select {
		//if requested is responded before given timeout period
		case <-done:
			encoder.Encode(OnlineIds)
			return
		//if cancelFunc is run
		case <-ctx.Done():
			encoder.Encode(OnlineIds)
			return
		}
	})
}

func init() {
	// Initialization of configuration
	Config = BitburstConfig{
		ConnString:       "postgres://postgres:secret@localhost:5432/bitburst?sslmode=disable",
		DeletionTimeSpan: 30 * time.Second,
		TimeoutTimeSpan:  2 * time.Second,
		EventQueueLength: 200,
		UserStatusUrl:    "http://localhost:9010/objects/",
	}

	//Creating connection to database
	IGormClient = &dataservice.GormClient{}
	IGormClient.SetupDB(Config.ConnString)

}

func main() {
	//Callback handler
	http.Handle("/callback", timeOutMiddleware(OnlineUsersHandler{}))

	//Assigning job
	go Job(IGormClient.DeleteObsoleteUsers, Config.DeletionTimeSpan)

	//Running server
	log.Fatal(http.ListenAndServe(":9090", nil))
}

//GetUserStatus create and send the request to get online status of given Id.
//After getting online status of given id it puts user into the channel.
func GetUserStatus(ctx context.Context, id int, chUser chan model.User) {

	var err error
	var res *http.Response
	var req *http.Request

	req, err = http.NewRequest("get", Config.UserStatusUrl+strconv.FormatInt(int64(id), 10), nil)
	req.WithContext(ctx)

	client := &http.Client{}
	res, err = client.Do(req)

	if err != nil {
		log.Println(err)
		return
	}
	var respBytes []byte
	respBytes, err = ioutil.ReadAll(res.Body)
	if err != nil {
		log.Println(err)
		return
	}

	usr := model.User{}
	json.Unmarshal(respBytes, &usr)

	chUser <- usr
}

//Job scheduler allows you to define task that need to be automatically executed at a given point in time t.
func Job(task func(time.Duration) error, t time.Duration) {
	ticker := time.Tick(1 * time.Second)
	for range ticker {
		task(t)
	}
}
