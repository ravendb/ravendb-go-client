package tests

import (
	"github.com/ravendb/ravendb-go-client"
	"time"
)

// Note: Java's ReplicationTestBase is folded into RavenTestDriver

func (d *RavenTestDriver) modifyReplicationDestination(replicationNode *ravendb.ReplicationNode) {
	// empty by design
}

func (d *RavenTestDriver) setupReplication(fromStore *ravendb.DocumentStore, destinations ...*ravendb.DocumentStore) error {
	for _, store := range destinations {
		databaseWatcher := ravendb.NewExternalReplication(store.GetDatabase(), "ConnectionString-"+store.GetIdentifier())
		d.modifyReplicationDestination(&databaseWatcher.ReplicationNode)

		if err := d.addWatcherToReplicationTopology(fromStore, databaseWatcher); err != nil {
			return err
		}
	}
	return nil
}

func (d *RavenTestDriver) addWatcherToReplicationTopology(store *ravendb.DocumentStore, watcher *ravendb.ExternalReplication) error {
	connectionString := ravendb.NewRavenConnectionString()
	connectionString.Name = watcher.ConnectionStringName
	connectionString.Database = watcher.Database
	connectionString.TopologyDiscoveryUrls = store.GetUrls()

	err := store.Maintenance().Send(ravendb.NewPutConnectionStringOperation(connectionString))
	if err != nil {
		return err
	}
	op := ravendb.NewUpdateExternalReplicationOperation(watcher)
	return store.Maintenance().Send(op)
}

func (d *RavenTestDriver) waitForDocumentToReplicate(store *ravendb.DocumentStore, result interface{}, id string, timeout time.Duration) error {
	sw := time.Now()

	for {
		time.Sleep(time.Millisecond * 500)
		dur := time.Since(sw)
		if dur > timeout {
			return nil
		}
		{
			session, err := store.OpenSession("")
			if err != nil {
				return err
			}
			err = session.Load(result, id)
			session.Close()
			if err != nil {
				return err
			}
		}
	}
}
