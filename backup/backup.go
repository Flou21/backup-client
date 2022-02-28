package backup

import (
	"fmt"
	"os"
	"path"
	"strconv"
	"sync"
	"time"

	"github.com/Coflnet/db-backup/client/api"
	"github.com/Coflnet/db-backup/client/model"
	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func StartBackup() {

	targets, err := api.LoadBackupTargets()
	if err != nil {
		log.Fatal().Err(err).Msgf("could not load backup targets")
	}

	if len(targets) == 0 {
		log.Info().Msgf("no targets to backup, stopping application..")
		return
	}

	backups := make(chan *model.Backup)

	go func() {
		defer close(backups)
		var wg sync.WaitGroup
		for _, target := range targets {
			wg.Add(1)
			go func(target *model.Target) {
				defer wg.Done()
				log.Info().Msgf("starting backup of %s, type: %s", target.Name, target.Type)
				err := BackupTarget(target, backups)
				if err != nil {
					log.Error().Err(err).Msgf("failed to backup target %s", target.Name)
					return
				}
				log.Info().Msgf("backup of target %s finished")
			}(target)
		}
		wg.Wait()
	}()

	for backup := range backups {
		log.Info().Msgf("created a backup of %s on timestamp %s with size %s", backup.Target.Name, backup.Timestamp, backup.Size)
	}
}

func BackupTarget(target *model.Target, backups chan *model.Backup) error {

	if target.Type == model.TARGET_MYSQL {
		return fmt.Errorf("mysql backups are not implemented yet")
	}

	if target.Type == model.TARGET_MONGO {
		err := BackupMongo(target, backups)
		if err != nil {
			log.Error().Err(err).Msgf("failed to create mongo backup")
			return err
		}
		return nil
	}

	if target.Type == model.TARGET_CASSANDRA {
		err := BackupCassandra(target, backups)
		if err != nil {
			log.Error().Err(err).Msgf("error when creating cassandra backup")
			return err
		}
		return nil
	}

	return fmt.Errorf("backup target is unknown, target: %s", target.Type)
}

func BasePath() (string, error) {
	basePath := os.Getenv("BACKUP_BASE_PATH")

	if basePath == "" {
		log.Error().Msgf("BACKUP_BASE_PATH is not set")
		return "", fmt.Errorf("BACKUP_BASE_PATH not set")
	}

	return basePath, nil
}

func BackupName(target *model.Target, tableOrCollection string) (string, error) {

	basePath, err := BasePath()
	if err != nil {
		log.Error().Err(err).Msgf("could not get base path for backup target: %s", target.Name)
		return "", err
	}
	baseName := time.Now().Format(time.RFC3339)

	p := path.Join(basePath, target.Name, tableOrCollection, baseName)

	return p, nil
}

func CreateBackupInstance(target *model.Target, backupPath string) (*model.Backup, error) {

	f, err := os.Stat(backupPath)
	if err != nil {
		log.Error().Err(err).Msgf("could not open file %s to measure size", backupPath)
		return nil, err
	}

	size := f.Size()

	backup := &model.Backup{
		ID:        primitive.NewObjectID(),
		Timestamp: time.Now(),
		Path:      backupPath,
		Size:      size,
		Target: &model.Target{
			ID:                     target.ID,
			Name:                   target.Name,
			Type:                   target.Type,
			Port:                   target.Port,
			Username:               target.Username,
			Password:               strconv.Itoa(len(target.Password)),
			Ip:                     target.Ip,
			Interval:               target.Interval,
			Database:               target.Database,
			AuthenticationDatabase: target.AuthenticationDatabase,
		},
	}

	return backup, nil
}

func createBackupPathIfNecessary(name string) error {
	path := path.Dir(name)
	perm := os.FileMode(777)

	log.Info().Msgf("creating path %s", path)

	err := os.MkdirAll(path, perm)
	if err != nil {
		log.Error().Err(err).Msgf("an error occurred creating path %s", path)
		return err
	}

	return nil
}

func SaveBackupOfCollectionOrTable(target *model.Target, backupPath string) (*model.Backup, error) {

	// create instance
	log.Info().Msgf("start saving backup to the api..")
	backup, err := CreateBackupInstance(target, backupPath)

	if err != nil {
		log.Error().Err(err).Msgf("could not create backup of target %s to the api", target)
		return nil, err
	}
	log.Info().Msgf("backup of target %s with timestamp %s was created, file: %s", target.Name, backup.Timestamp, backup.Path)

	// upload instance
	api.CreateBackup(backup)
	log.Info().Msgf("uploaded backup %s", target.Name)

	return backup, nil
}
