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
	AddedOn         int32
	UpdatedOn       int32
}

type ByUpdatedOnDesc []PocketSave

func (s ByUpdatedOnDesc) Len() int           { return len(s) }
func (s ByUpdatedOnDesc) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s ByUpdatedOnDesc) Less(i, j int) bool { return s[j].UpdatedOn < s[i].UpdatedOn }

func (i PocketSave) Title() string {
	return i.SaveTitle
}

func (i PocketSave) Description() string {
	if i.SaveDescription == "" {
		return i.Url
	} else {
		return i.SaveDescription
	}
}
func (i PocketSave) FilterValue() string { return i.SaveTitle }
