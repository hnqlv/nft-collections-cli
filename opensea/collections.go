package main

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/urfave/cli/v2"
	"go.uber.org/zap"
)

type Address struct {
	ContractAddress string `json:"address"`
}

type Stats struct {
	SevenDayVolume  float64 `json:"seven_day_volume" bson:"seven_day_volume"`
	ThirtyDayVolume float64 `json:"thirty_day_volume" bson:"thirty_day_volume"`
	TotalVolume     float64 `json:"total_volume" bson:"total_volume"`
}

type ResponseCollections struct {
	Collections []Collections `json:"collections"`
}

type Collections struct {
	Name                  string    `json:"name"`
	Slug                  string    `json:"slug"`
	DiscordUrl            string    `json:"discord_url"`
	TwitterUsername       string    `json:"twitter_username"`
	PrimaryAssetContracts []Address `json:"primary_asset_contracts"`
	Stats                 Stats     `json:"stats"`
	CreatedDate           string    `json:"created_date"`
}

type CollectionItem struct {
	Name            string `bson:"name"`
	Slug            string `bson:"slug"`
	DiscordUrl      string `bson:"discord_url"`
	TwitterUsername string `bson:"twitter_username"`
	Stats           Stats  `bson:"stats"`
}

type Record struct {
	Name        string
	TotalVolume float64
}

var (
	client = http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:    50,
			IdleConnTimeout: 30 * time.Second,
		},
	}
)

func read(c []CollectionItem) [][]string {
	var records [][]string
	for _, row := range c {
		item := []string{
			row.Name,
			fmt.Sprintf("%f", row.Stats.TotalVolume),
			fmt.Sprintf("%f", row.Stats.SevenDayVolume),
			fmt.Sprintf("%f", row.Stats.ThirtyDayVolume),
			row.DiscordUrl,
			row.TwitterUsername,
			row.Slug,
		}
		records = append(records, item)
	}
	return records
}

func retrieveCollections(ctx *cli.Context) error {
	logger, _ := zap.NewProduction()
	defer logger.Sync()
	sugar := logger.Sugar()

	offset := 0
	pagination := 300
	z := 0

	recordFile, err := os.Create("./sample.csv")
	if err != nil {
		fmt.Println("An error encountered ::", err)
	}

	writer := csv.NewWriter(recordFile)
	// heading
	heading := [][]string{
		{"Name", "Total Volume", "Seven Day Volume", "Thirty Day Volume", "Discord", "Twitter", "Slug"},
	}
	writer.WriteAll(heading)

	for {
		if offset > ctx.Int("total") {
			cli.Exit("Bye", 86)
			break
		}
		offset = pagination * z
		sugar.Infof("offset %d", offset)
		collections, err := getCollections(ctx.Context, offset, pagination)
		if err != nil {
			sugar.Fatal(err)
			break
		}
		records := read(collections)
		err = writer.WriteAll(records)
		if err != nil {
			fmt.Println("An error encountered ::", err)
		}
		z++
	}

	return nil
}

func getCollections(ctx context.Context, offset int, limit int) ([]CollectionItem, error) {
	url := fmt.Sprintf("https://api.opensea.io/api/v1/collections?offset=%d&limit=%d", offset, limit)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	req.Header.Add("Accept", "application/json")
	if err != nil {
		return nil, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	var resCollections ResponseCollections
	decoder := json.NewDecoder(resp.Body)
	if err = decoder.Decode(&resCollections); err != nil {
		return nil, err
	}

	var collectionItems []CollectionItem

	for _, cl := range resCollections.Collections {
		item := CollectionItem{
			Name:            cl.Name,
			Slug:            cl.Slug,
			DiscordUrl:      cl.DiscordUrl,
			TwitterUsername: cl.TwitterUsername,
			Stats:           cl.Stats,
		}
		l := len(cl.PrimaryAssetContracts)
		if l > 0 && cl.Stats.SevenDayVolume > 1 {
			collectionItems = append(collectionItems, item)
		}
	}

	return collectionItems, nil
}
