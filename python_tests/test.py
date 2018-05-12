from pyravendb.store import document_store
from pyravendb.raven_operations.server_operations import GetDatabaseNamesOperation, CreateDatabaseOperation, DeleteDatabaseOperation
from pyravendb.raven_operations.maintenance_operations import GetStatisticsOperation
from pyravendb.commands.raven_commands import GetTopologyCommand, PutDocumentCommand, GetDocumentCommand, DeleteDocumentCommand
from pyravendb.hilo.hilo_generator import HiLoKeyGenerator
import uuid
from builtins import ValueError

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
    if verboseLog: print(res)
    testDbName = dbName
    print("testCreateDatabaseOp ok")

def testGetDatabaseNamesOp():
    store =  document_store.DocumentStore(urls=["http://localhost:9999"], database="")
    store.initialize()
    op = GetDatabaseNamesOperation(0, 32)
    names = store.maintenance.server.send(op)
    if verboseLog: print(names)
    if testDbName not in names:
        raise ValueError("{0} not found in {1}".format(testDbName, names))
    print("testGetDatabaseNamesOp ok")

def testGetStatisticsOp():
    store =  document_store.DocumentStore(urls=["http://localhost:9999"], database=testDbName)
    store.initialize()
    op = GetStatisticsOperation()
    res = store.maintenance.send(op)
    if verboseLog: print(res)
    print("testGetStatisticsOp ok")

def testGetStatisticsBadDb():
    store =  document_store.DocumentStore(urls=["http://localhost:9999"], database="not-exists")
    store.initialize()
    op = GetStatisticsOperation()
    failed = False
    try:
        res = store.maintenance.send(op)
        if verboseLog: print(res)
    except Exception as e:
        failed = True
    assert failed, "GetTopologyCommand() was supposed to throw an exception"
    print("testGetStatisticsBadDb ok")

def testGetTopology():
    store =  document_store.DocumentStore(urls=["http://localhost:9999"], database=testDbName)
    store.initialize()
    with store.open_session() as session:
        op = GetTopologyCommand()
        res = session.requests_executor.execute(op)
        if verboseLog: print(res)
        print("testGetTopology ok")

def testGetTopologyBadDb():
    store =  document_store.DocumentStore(urls=["http://localhost:9999"], database="invalid-db")
    store.initialize()
    with store.open_session() as session:
        op = GetTopologyCommand()
        failed = False
        try:
            res = session.requests_executor.execute(op)
            if verboseLog: print(res)
        except Exception as e:
            failed = True
        assert failed, "GetTopologyCommand() was supposed to throw an exception"
    print("testGetTopologyBadDb ok")

# def testCreateAndDeleteDatabaseOp():
#     dbName = "tst_" + uuid.uuid4().hex
#     print("name: " + dbName)
#     store =  document_store.DocumentStore(urls=["http://localhost:9999"], database="")
#     store.initialize()
#     op = CreateDatabaseOperation(database_name=dbName)
#     res = store.maintenance.server.send(op)
#     print(res)
#     op = DeleteDatabaseOperation(database_name=dbName, hard_delete=False)
#     res = store.maintenance.server.send(op)
#     print(res)

def testDeleteDatabaseOp():
    store =  document_store.DocumentStore(urls=["http://localhost:9999"], database="")
    store.initialize()
    op = DeleteDatabaseOperation(database_name=testDbName, hard_delete=True)
    res = store.maintenance.server.send(op)
    if verboseLog: print(res)
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

def testPutGetDeleteDocument():
    store =  document_store.DocumentStore(urls=["http://localhost:9999"], database=testDbName)
    store.initialize()
    re = store.get_request_executor()
    doc = {
        "Name": "test1",
        "DocNumber": 1,
        "@metadata": {
            "@collection": "Testings"
        }
    }
    key = "testing/1"
    cmd = PutDocumentCommand(key, doc)
    res = re.execute(cmd)
    if verboseLog: print(res)

    cmd = GetDocumentCommand(key)
    res = re.execute(cmd)
    if verboseLog: print(res)

    # test get of non-existent document
    cmd = GetDocumentCommand("testing/1234")
    res = re.execute(cmd)
    assert res == None, "unexpected res != None"

    cmd = DeleteDocumentCommand(key)
    res = re.execute(cmd)
    if verboseLog: print(res)

    # test delete of non-existent document
    # it succeeds even if document doesn't exist
    cmd = DeleteDocumentCommand("testing/1234")
    res = re.execute(cmd)
    if verboseLog: print(res)
    print("testPutGetDelete ok")

def testHiLoKeyGenerator():
    store =  document_store.DocumentStore(urls=["http://localhost:9999"], database=testDbName)
    store.initialize()
    tag = "my_tag"
    generator = HiLoKeyGenerator(tag, store, testDbName)
    res = generator.generate_document_key()
    if verboseLog: print(res)
    res = generator.return_unused_range()
    if verboseLog: print(res)
    print("testHiLoKeyGenerator ok")

class Foo(object):
   def __init__(self, name, key = None):
        self.name = name
        self.key = key

class FooBar(object):
    def __init__(self, name, foo):
        self.name = name
        self.foo = foo

def testStoreLoad():
    store =  document_store.DocumentStore(urls=["http://localhost:9999"], database=testDbName)
    store.initialize()
    with store.open_session() as session:
        foo = Foo("PyRavenDB")
        session.store(foo)
        session.save_changes()


all_tests = False
def main():
    deleteTestDatabases()
    testCreateDatabaseOp()

    if all_tests:
        testGetDatabaseNamesOp()
        testGetTopology()
        testGetTopologyBadDb()

        testGetStatisticsOp()
        testGetStatisticsBadDb()

        testPutGetDeleteDocument()

    testHiLoKeyGenerator()

    #testDeleteDatabaseOp()

if __name__ == "__main__":
    main()

