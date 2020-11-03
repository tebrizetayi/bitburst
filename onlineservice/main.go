package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"time"
)

func main() {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	go func() {
		client := &http.Client{Timeout: 6 * time.Second}

		for {
			time.Sleep(5 * time.Second)

			ids := make([]string, rng.Int31n(200))
			for i := range ids {
				ids[i] = strconv.Itoa(rng.Int() % 100)
			}
			body := bytes.NewBuffer([]byte(fmt.Sprintf(`{"object_ids":[%s]}`, strings.Join(ids, ","))))
			resp, err := client.Post("http://localhost:9090/callback", "application/json", body)
			if err != nil {
				fmt.Println(err)
				continue
			}

			var bytes []byte
			bytes, err = ioutil.ReadAll(resp.Body)
			if err != nil {
				fmt.Println(err)
			} else {
				log.Println(string(bytes))
			}
			_ = resp.Body.Close()

		}
	}()

	http.HandleFunc("/objects/", func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(time.Duration(rng.Int63n(4000)+300) * time.Millisecond)

		idRaw := strings.TrimPrefix(r.URL.Path, "/objects/")
		id, err := strconv.Atoi(idRaw)
		if err != nil {
			http.Error(w, "invalid id", http.StatusBadRequest)
			return
		}

		w.Write([]byte(fmt.Sprintf(`{"id":%d,"online":%v}`, id, rng.Int()%2 == 0)))
	})
	go func() { _ = http.ListenAndServe(":9010", nil) }()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, os.Kill)
	<-sig

	fmt.Println("closing")
}
