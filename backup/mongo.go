package backup

import (
	"context"
	"fmt"
	"os/exec"
	"sync"

	"github.com/Coflnet/db-backup/client/model"
	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func BackupMongo(target *model.Target, backups chan<- *model.Backup) error {

	log.Info().Msgf("collecting collections for target %s ...", target.Name)

	collections, err := CollectionsFromDatabase(target)
	if err != nil {
		log.Error().Err(err).Msgf("error trying to list collections for target %s", target.Name)
		return err
	}

	log.Info().Msgf("backup for database %s is running..", target.Name)
	ScheduleBackups(target, collections, backups)
	return nil
}

func ScheduleBackups(target *model.Target, collections []string, backupChannel chan<- *model.Backup) {
	var wg sync.WaitGroup
	for _, collection := range collections {
		wg.Add(1)
		go func(target *model.Target, collection string, backupChannel chan<- *model.Backup) {
			defer wg.Done()
			err := BackupMongoCollection(target, collection, backupChannel)
			if err != nil {
				log.Error().Err(err).Msgf("there was an error when backing up collection %s", collection)
				return
			}
		}(target, collection, backupChannel)
	}
	wg.Wait()
}

func BackupMongoCollection(target *model.Target, collection string, backupChannel chan<- *model.Backup) error {
	backupName, err := BackupName(target, collection)
	if err != nil {
		log.Error().Err(err).Msgf("could not get name of backup")
		return err
	}

	backupName = fmt.Sprintf("%s-%s.json", backupName, collection)
	err = createBackupPathIfNecessary(backupName)
	if err != nil {
		log.Error().Err(err).Msgf("could not create backup path")
		return err
	}

	log.Info().Msgf("starting backup for collection %s", collection)
	_, err = exec.Command(
		"mongoexport",
		"--host",
		target.Ip,
		"--port",
		fmt.Sprint(target.Port),
		"--password",
		target.Password,
		"--username",
		target.Username,
		"--authenticationDatabase",
		target.AuthenticationDatabase,
		"--db",
		target.Database,
		"--collection",
		collection,
		"--out",
		backupName,
	).CombinedOutput()

	if err != nil {
		log.Error().Err(err).Msgf("command failed")
		return err
	}

	backup, err := SaveBackupOfCollectionOrTable(target, backupName)
	if err != nil {
		log.Error().Err(err).Msgf("error saving backup")
		return err
	}

	backupChannel <- backup
	return nil
}

func CollectionsFromDatabase(target *model.Target) ([]string, error) {
	uri := fmt.Sprintf("mongodb://%s:%s@%s:%d", target.Username, target.Password, target.Ip, target.Port)

	log.Info().Msgf("connecting to database with uri: %s", uri)

	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(uri))
	if err != nil {
		log.Error().Err(err).Msgf("error connecting to backup database")
		return nil, err
	}

	collections, err := client.Database(target.Database).ListCollectionNames(context.TODO(), bson.D{})
	if err != nil {
		log.Error().Err(err).Msgf("error getting collection names for backup database")
		return nil, err
	}

	return collections, nil
}
