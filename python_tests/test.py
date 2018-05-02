from pyravendb.store import document_store

store =  document_store.DocumentStore(urls=["http://localhost:9999"], database="PyRavenDB")
store.initialize()
with store.open_session() as session:
    foo = session.load("foos/1")
    print(foo)

