from pyravendb.store import document_store
from pyravendb.raven_operations.server_operations import GetDatabaseNamesOperation, CreateDatabaseOperation, DeleteDatabaseOperation
from pyravendb.raven_operations.maintenance_operations import GetStatisticsOperation
from pyravendb.commands.raven_commands import GetTopologyCommand, PutDocumentCommand
import uuid
from builtins import ValueError

# def testLoad():
#     store =  document_store.DocumentStore(urls=["http://localhost:9999"], database="PyRavenDB2")
#     store.initialize()

#     with store.open_session() as session:
#         foo = session.load("foos/1")
#         print(foo)

#     database_names = store.maintenance.server.send(GetDatabaseNamesOperation(0, 3))
#     print(database_names)

testDbName = None

verboseLog = True

# test cration of a database. A pre-requesite for some other tests
def testCreateDatabaseOp():
    global testDbName
    dbName = "tst_" + uuid.uuid4().hex
    store =  document_store.DocumentStore(urls=["http://localhost:9999"], database="")
    store.initialize()
    op = CreateDatabaseOperation(database_name=dbName)
    res = store.maintenance.server.send(op)
    if verboseLog:
        print(res)
    testDbName = dbName
    print("testCreateDatabaseOp ok")

def testGetDatabaseNamesOp():
    store =  document_store.DocumentStore(urls=["http://localhost:9999"], database="")
    store.initialize()
    op = GetDatabaseNamesOperation(0, 32)
    names = store.maintenance.server.send(op)
    if verboseLog:
        print(names)
    if testDbName not in names:
        raise ValueError("{0} not found in {1}".format(testDbName, names))
    print("testGetDatabaseNamesOp ok")

def testGetStatisticsOp():
    store =  document_store.DocumentStore(urls=["http://localhost:9999"], database=testDbName)
    store.initialize()
    op = GetStatisticsOperation()
    res = store.maintenance.send(op)
    print(res)

def testGetStatisticsBadDb():
    store =  document_store.DocumentStore(urls=["http://localhost:9999"], database="not-exists")
    store.initialize()
    op = GetStatisticsOperation()
    res = store.maintenance.send(op)
    print(res)

def testGetTopology():
    store =  document_store.DocumentStore(urls=["http://localhost:9999"], database=testDbName)
    store.initialize()
    with store.open_session() as session:
        op = GetTopologyCommand()
        res = session.requests_executor.execute(op)
        if verboseLog:
            print(res)
        print("testGetTopology ok")

def testGetTopologyBadDb():
    store =  document_store.DocumentStore(urls=["http://localhost:9999"], database="invalid-db")
    store.initialize()
    with store.open_session() as session:
        op = GetTopologyCommand()
        res = session.requests_executor.execute(op)
        if verboseLog:
            print(res)
        print("testGetTopologyBadDb ok")

def testCreateAndDeleteDatabaseOp():
    dbName = "tst_" + uuid.uuid4().hex
    print("name: " + dbName)
    store =  document_store.DocumentStore(urls=["http://localhost:9999"], database="")
    store.initialize()
    op = CreateDatabaseOperation(database_name=dbName)
    res = store.maintenance.server.send(op)
    print(res)
    op = DeleteDatabaseOperation(database_name=dbName, hard_delete=False)
    res = store.maintenance.server.send(op)
    print(res)

def testDeleteDatabaseOp():
    store =  document_store.DocumentStore(urls=["http://localhost:9999"], database="")
    store.initialize()
    op = DeleteDatabaseOperation(database_name=testDbName, hard_delete=True)
    res = store.maintenance.server.send(op)
    if verboseLog:
        print(res)
    print("testDeleteDatabaseOp ok")

# delete all databases named "tst_" + uuid
def deleteTestDatabases():
    store =  document_store.DocumentStore(urls=["http://localhost:9999"], database="")
    store.initialize()
    op = GetDatabaseNamesOperation(0, 45)
    names = store.maintenance.server.send(op)
    print("Database: {0}".format(names))
    for dbName in names:
        if not dbName.startswith("tst_"):
            continue
        print("Deleting database: " + dbName)
        op = DeleteDatabaseOperation(database_name=dbName, hard_delete=True)
        res = store.maintenance.server.send(op)
        print(res)


def testPutGetDelete():
    # create randomly named database
    dbName = "tst_" + uuid.uuid4().hex
    print("name: " + dbName)
    store =  document_store.DocumentStore(urls=["http://localhost:9999"], database="")
    store.initialize()
    op = CreateDatabaseOperation(database_name=dbName)
    res = store.maintenance.server.send(op)
    print(res)

    cmd = PutDocumentCommand("testing/" + str(i),
                                        {"Name": "test" + str(i), "DocNumber": i,
                                        "@metadata": {"@collection": "Testings"}})



def main():
    deleteTestDatabases()

    testCreateDatabaseOp()
    testGetDatabaseNamesOp()
    testGetTopology()
    #testGetTopologyBadDb()

    #testGetStatisticsOp()
    #testGetStatisticsBadDb()
    #testCreateAndDeleteDatabaseOp()

    #testPutGetDelete()

    testDeleteDatabaseOp()

if __name__ == "__main__":
    main()

