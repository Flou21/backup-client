package backup

import (
	"fmt"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/Coflnet/db-backup/client/model"
	"github.com/gocql/gocql"
	"github.com/rs/zerolog/log"
)

var (
	session *gocql.Session
	cluster *gocql.ClusterConfig
)

func createCassandraSession(target *model.Target) error {
	log.Info().Msgf("using host %s for cassandra backup", target.Ip)
	ips := strings.Split(target.Ip, ",")
	cluster = gocql.NewCluster(ips...)
	cluster.Keyspace = target.Database
	cluster.Authenticator = gocql.PasswordAuthenticator{
		Username: target.Username,
		Password: target.Password,
	}
	cluster.ConnectTimeout = time.Second * 30

	var err error
	session, err = cluster.CreateSession()
	if err != nil {
		log.Error().Err(err).Msgf("there was an error when creating session to cassandra")
		return err
	}

	log.Info().Msgf("created a session to cassandra")

	return nil
}

func BackupCassandra(target *model.Target, backups chan<- *model.Backup) error {
	log.Info().Msgf("starting backup for cassandra target %s", target.Name)
	err := createCassandraSession(target)
	if err != nil {
		log.Error().Err(err).Msgf("there was an error when connecting to cassandra")
		return err
	}

	tables, err := listCassandraTables(target)
	if err != nil {
		log.Error().Err(err).Msgf("there was an error when listing tables for target %s", target.Name)
		return err
	}

	var wg sync.WaitGroup
	for _, table := range tables {
		wg.Add(1)
		go func(table string) {
			defer wg.Done()

			backup, err := backupCassandraTable(target, table)
			if err != nil {
				log.Error().Err(err).Msgf("there was an error when creating database for table %s in target %s", table, target.Name)
				return
			}

			SaveBackupOfCollectionOrTable(target, backup.Path)

			backups <- backup
		}(table)
	}

	wg.Wait()
	log.Info().Msgf("created %d backups for cassandra target %s", len(tables), target.Name)

	return nil
}

func backupCassandraTable(target *model.Target, table string) (*model.Backup, error) {

	// get backupName / Path
	backupName, err := BackupName(target, table)
	if err != nil {
		log.Error().Err(err).Msgf("there was an error creating backup name")
		return nil, err
	}

	// prepare path and filename
	createBackupPathIfNecessary(backupName)
	backupName = fmt.Sprintf("%s.csv", backupName)

	// create actual backup
	log.Info().Msgf("creating a backup, file: %s, table: %s", backupName, table)
	err = executeBackup(target, backupName, table)
	if err != nil {
		log.Error().Err(err).Msgf("there was an error creating cassandra backup for table %s", table)
		return nil, err
	}

	log.Info().Msgf("backup was created successfully")

	// create backup instance to return
	backup, err := CreateBackupInstance(target, backupName)
	if err != nil {
		log.Error().Err(err).Msgf("there was an error creating backup instance")
		return nil, err
	}

	return backup, nil
}

func executeBackup(target *model.Target, backupName string, table string) error {

	// try every host until it works or now hosts left
	hosts := strings.Split(target.Ip, ",")
	var err error

	cmd := fmt.Sprintf("use %s; COPY %s TO '%s'", target.Database, table, backupName)

	log.Info().Msgf("cmd: %s, hosts count: %d", cmd, len(hosts))

	for _, host := range hosts {
		_, err := exec.Command(
			"cqlsh",
			"-u",
			target.Username,
			host,
			"-p",
			target.Password,
			"-e",
			cmd,
		).CombinedOutput()

		if err != nil {
			log.Error().Err(err).Msgf("an error occured while executing backup, tried host %s", host)
		}

		if err == nil {
			return nil
		}
	}

	return err
}

func listCassandraTables(target *model.Target) ([]string, error) {
	if session.Closed() {
		return nil, fmt.Errorf("session is closed")
	}

	metadata, err := session.KeyspaceMetadata(target.Database)
	if err != nil {
		log.Error().Err(err).Msgf("error getting keyspace metadata, for keyspace %s", target.Database)
	}

	tables := []string{}
	for t := range metadata.Tables {
		tables = append(tables, t)
	}

	return tables, nil
}
