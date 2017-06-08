package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/manyminds/api2go/jsonapi"
	"upper.io/db.v3"
	"upper.io/db.v3/lib/sqlbuilder"
)

func doSync(wg *sync.WaitGroup, remote string, timestamp *time.Time) {
	defer wg.Done()
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("Recovered from a panic %v while syncing %s", r, remote)
		}
	}()

	client := &http.Client{Timeout: 10 * time.Second}

	for {
		u, err := url.Parse(remote)
		if err != nil {
			panic(err)
		}

		u.RawQuery = url.Values{
			"filter[Modified][gt]": []string{timestamp.Format(time.RFC3339Nano)},
		}.Encode()

		fmt.Printf("GETing %+v\n", u)
		resp, err := client.Get(u.String())
		if err != nil {
			panic(err)
		}

		data, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			panic(err)
		}

		assets := make([]Asset, 0)

		if err = jsonapi.Unmarshal(data, &assets); err != nil {
			panic(err)
		}

		if len(assets) == 0 {
			return
		}

		err = DB.Tx(DB.Context(), func(tx sqlbuilder.Tx) error {
			var existing Asset
			collection := tx.Collection("asset")

			for _, asset := range assets {
				err := collection.Find(db.Cond{"id": asset.Uri}).One(&existing)

				if err != nil && err == db.ErrNoMoreRows {
					fmt.Printf("Found missing Asset %v", asset)
					_, err = collection.Insert(asset)
					if err != nil {
						return err
					}
					continue
				}

				if err != nil {
					return err
				}

				// TODO Deleted
				if existing.Name == asset.Name && existing.Created == asset.Created {
					continue
				}

				// Senority rules
				if existing.Created.After(asset.Created) {
					fmt.Printf("Found conflicting Asset %v", asset)
					// Bump modification so other servers will see this change
					asset.Modified = time.Now().UTC()
					err = collection.Find(db.Cond{"id": asset.Uri}).Update(asset)
					if err != nil {
						return err
					}
				}

			}

			return nil
		})

		timestamp = &assets[len(assets)-1].Created
		fmt.Printf("New timestamp is %+v\n", timestamp)
	}
}

func SyncAssets(remotes map[string]time.Time) {
	if len(remotes) == 0 {
		fmt.Println("Recieved no remotes, syncing will be disabled")
		return
	}

	fmt.Println("Sync goroutine started\nSyncing to:")
	for k, _ := range remotes {
		fmt.Printf("\t* %s\n", k)
	}

	for now := range time.Tick(10 * time.Second) {
		var wg sync.WaitGroup

		for remote, timestamp := range remotes {
			wg.Add(1)
			fmt.Printf("Syncing %s from %s at %s\n", remote, timestamp, now)
			go doSync(&wg, remote, &timestamp)
		}

		wg.Wait()
		fmt.Printf("Finished sync")
	}
}
