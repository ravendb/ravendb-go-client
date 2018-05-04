from pyravendb.store import document_store
from pyravendb.raven_operations.server_operations import GetDatabaseNamesOperation

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
    database_names = store.maintenance.server.send(GetDatabaseNamesOperation(0, 3))
    print(database_names)

def main():
    testGetDatabaseNamesOp()

if __name__ == "__main__":
    main()

