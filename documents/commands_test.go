package documents

import (
	"testing"
	testingUtils "../testing"
)

func TestCRUD(t *testing.T){

	store, _ := NewDocumentStore("test")
	session, _ := store.OpenSession()
	session.Store(testingUtils.User{Name: "user1"}, 1, "user/1")
	user2 := testingUtils.User{Name: "user2", Age: 1}
	session.Store(user2,  2, "user/2")
	user3 := testingUtils.User{Name: "user3", Age: 1}
	session.Store(user3, 3, "user/3")
	session.Store(testingUtils.User{Name: "user4"}, 4, "user/4")

	session.Delete(user2)
	user3.Age = 3
	session.SaveChanges()

	tempUser, ok := session.Load("user/2")
	if ok || tempUser != nil{
		t.Fail()
	}
	tempUser, ok = session.Load("user/3")
	if !ok || tempUser.Age != 3{
		t.Fail()
	}
	user1 := session.Load("users/1")
	user4 := session.Load("users/4")

	session.Delete(user4)
	user1.Age = 10
	session.SaveChanges()

	tempUser, ok = session.Load("user/4")
	if ok || tempUser != nil{
		t.Fail()
	}
	tempUser, ok = session.Load("user/1")
	if !ok || tempUser.Age != 10{
		t.Fail()
	}
}