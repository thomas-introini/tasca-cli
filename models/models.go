package models

var NoUser = PocketUser{}

type PocketUser struct {
	AccessToken    string
	Username       string
	SavesUpdatedOn int32
}

type PocketSave struct {
	Id              string
	SaveTitle       string
	Url             string
	SaveDescription string
	UpdatedOn       int32
}

func (i PocketSave) Title() string       { return i.SaveTitle }
func (i PocketSave) Description() string { return i.SaveDescription }
func (i PocketSave) FilterValue() string { return i.SaveTitle }
