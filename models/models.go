package models

var NoUser = PocketUser{}

type PocketUser struct {
	AccessToken    string
	Username       string
	SavesUpdatedOn int32
}

const (
	StatusOK       = 0
	StatusArchived = 1
	StatusDeleted  = 2
)

type PocketSave struct {
	Id              string
	SaveTitle       string
	Url             string
	SaveDescription string
	TimeToRead      uint16
	Favorite        bool
	Status          uint8
	Tags            string
	AddedOn         uint32
	UpdatedOn       uint32
}

type ByAddedOnDesc []PocketSave
func (s ByAddedOnDesc) Len() int           { return len(s) }
func (s ByAddedOnDesc) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s ByAddedOnDesc) Less(i, j int) bool { return s[j].AddedOn < s[i].AddedOn }


func (i PocketSave) Title() (title string) {
	if i.SaveTitle == "" {
		title = i.Url
	} else {
		title = i.SaveTitle
	}
	return
}

func (i PocketSave) Description() string {
	if i.SaveDescription == "" {
		return i.Url
	} else {
		return i.SaveDescription
	}
}
func (i PocketSave) FilterValue() string { return i.SaveTitle }
