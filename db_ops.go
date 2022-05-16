package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v4"
	"log"
)

type UserWebsocketChannel struct {
	Channel string `json:"channel_name"` // table column name -> "md5"
}

func GetDataFromPGSQLTable() ([]UserWebsocketChannel, error) {
	defer elapsed("__FUNC__: GetDataFromPGSQLTable")()
	errorMessage := ""
	outputData := make([]UserWebsocketChannel, 0)
	postgresConn := fmt.Sprintf("postgres://%v:%v@%v:5432/%v", PGUSER, PGPASS, PGHOST, PGDB)
	db, err := pgx.Connect(context.Background(), postgresConn)
	if err != nil {
		errorMessage = fmt.Sprintf("unable to connect to database: %v", err.Error())
		log.Printf(errorMessage)
		return outputData, errors.New(errorMessage)
	}
	defer func() {
		_ = db.Close(context.Background())
	}()

	///////// READ MULTIPLE ROWS ////////////

	log.Printf("QUERYING ALL ROWS...")

	rows, err := db.Query(context.Background(), "select channel_name from user_websocket_channels")
	if err != nil {
		errorMessage = fmt.Sprintf("could not perform a select from the table : %v", err.Error())
		log.Printf(errorMessage)
		return outputData, errors.New(errorMessage)
	}
	defer func() {
		rows.Close()
	}()

	counter := 0

	for rows.Next() {
		var rowData UserWebsocketChannel
		err = rows.Scan(
			&rowData.Channel,
		)
		if err != nil {
			errorMessage = fmt.Sprintf("got an error during row scan : %v", err.Error())
			log.Printf(errorMessage)
			return outputData, errors.New(errorMessage)
		}

		/* *********** pretty print the output **************** */
		vmDataByteArray, _ := json.MarshalIndent(rowData, "", "    ")
		fmt.Println(string(vmDataByteArray))

		outputData = append(outputData, rowData)

		counter++
	}

	// get any error encountered during iteration
	err = rows.Err()
	if err != nil {
		errorMessage = fmt.Sprintf("got an error during row iteration : %v", err.Error())
		log.Printf(errorMessage)
		return outputData, errors.New(errorMessage)
	}

	log.Printf("after scanning all the rows, counter => %v", counter)

	return outputData, nil
}
