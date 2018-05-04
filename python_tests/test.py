from pyravendb.store import document_store
from pyravendb.raven_operations.server_operations import GetDatabaseNamesOperation
from pyravendb.raven_operations.maintenance_operations import GetStatisticsOperation

def testLoad():
    store =  document_store.DocumentStore(urls=["http://localhost:9999"], database="PyRavenDB2")
    store.initialize()

    with store.open_session() as session:
        foo = session.load("foos/1")
        print(foo)

    database_names = store.maintenance.server.send(GetDatabaseNamesOperation(0, 3))
    print(database_names)

def testGetDatabaseNamesOp():
    store =  document_store.DocumentStore(urls=["http://localhost:9999"], database="")
    store.initialize()
    res = store.maintenance.server.send(GetDatabaseNamesOperation(0, 3))
    print(res)

def testGetStatisticsOp():
    store =  document_store.DocumentStore(urls=["http://localhost:9999"], database="PyRavenDB")
    store.initialize()
    res = store.maintenance.send(GetStatisticsOperation())
    print(res)

def testGetStatisticsInvalidDb():
    store =  document_store.DocumentStore(urls=["http://localhost:9999"], database="not-exists")
    store.initialize()
    res = store.maintenance.send(GetStatisticsOperation())
    print(res)

def main():
    # testGetDatabaseNamesOp()
    #testGetStatisticsOp()
    testGetStatisticsInvalidDb()

if __name__ == "__main__":
    main()

