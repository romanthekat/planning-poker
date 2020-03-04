package mock

//
//import "github.com/EvilKhaosKat/FaceRecognitionBackend/pkg/models"
//
////Person used in mock model
//var Person = &models.Person{
//	FirstName: "First",
//	LastName:  "Name",
//	Email:     "email@email.com",
//	ID:        "1",
//	Encodings: []string{"1 2 3"},
//}
//
//type PersonModel struct {
//}
//
//func (*PersonModel) Update(id, firstName, lastName, email string, encodings []string) (string, error) {
//	switch id {
//	case "1":
//		Person.ID = id
//		Person.FirstName = firstName
//		Person.LastName = lastName
//		Person.Email = email
//		Person.Encodings = encodings
//
//		return id, nil
//	default:
//		return "", models.ErrDbProblem
//	}
//}
//
//func (*PersonModel) Get(id string) (*models.Person, error) {
//	switch id {
//	case "1":
//		return Person, nil
//	default:
//		return nil, models.ErrNoRecord
//	}
//}
//
//func (*PersonModel) Remove(id string) (int64, error) {
//	return 0, nil
//}
//
//func (*PersonModel) GetAll() ([]*models.Person, error) {
//	return []*models.Person{Person}, nil
//}
