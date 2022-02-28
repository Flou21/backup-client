package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/Coflnet/db-backup/client/model"
	"github.com/rs/zerolog/log"
)

var api_host = ""

func Init() error {
	api_host = os.Getenv("API_HOST")

	if api_host == "" {
		return fmt.Errorf("API_HOST must be set")
	}

	return nil
}

func LoadBackupTargets() ([]*model.Target, error) {

	url := fmt.Sprintf("%s/targetsWithPendingBackups", api_host)
	log.Info().Msgf("requesting backup targets from %s", url)
	resp, err := http.Get(url)
	if err != nil {
		log.Error().Err(err).Msgf("could not load backup targets")
		return nil, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Error().Err(err).Msgf("could not read body")
		return nil, err
	}

	targets := []*model.Target{}
	err = json.Unmarshal(body, &targets)
	if err != nil {
		log.Error().Err(err).Msgf("could not parse body")
		return nil, err
	}

	if len(targets) == 0 {
		log.Info().Msgf("loading targets was successful, but no targets were found")
	}

	return targets, nil
}

func CreateBackup(backup *model.Backup) error {

	url := fmt.Sprintf("%s/backup", api_host)
	log.Info().Msgf("requesting backup targets from %s", url)

	json_data, err := json.Marshal(backup)
	if err != nil {
		log.Error().Err(err).Msgf("there was an error marshalling the backup %s", backup.Target.Name)
		return err
	}

	_, err = http.Post(url, "application/json", bytes.NewBuffer(json_data))
	if err != nil {
		log.Error().Err(err).Msgf("there was an error posting the backup %s", backup.Target.Name)
		return err
	}

	log.Info().Msgf("backup was uploaded successfully")

	return nil
}
