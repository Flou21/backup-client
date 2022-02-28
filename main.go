package main

import (
	"github.com/Coflnet/db-backup/client/api"
	"github.com/Coflnet/db-backup/client/backup"
	"github.com/rs/zerolog/log"
)

func main() {
	log.Info().Msgf("starting backup process..")

	err := api.Init()
	if err != nil {
		log.Fatal().Err(err).Msgf("could not init api")
	}

	backup.StartBackup()
}
