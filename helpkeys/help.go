package helpkeys

import "github.com/charmbracelet/bubbles/key"

type ItemdetailsKeys struct {
	Quit       key.Binding
	Open       key.Binding
	GetContent key.Binding
	Archive    key.Binding
	Unarchive  key.Binding
	Delete     key.Binding
	EditTags   key.Binding
}

func (m ItemdetailsKeys) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{m.Quit},
		{m.Open},
		{m.Archive},
		{m.Delete},
		{m.EditTags},
	}
}

func (m ItemdetailsKeys) ShortHelp() []key.Binding {
	return []key.Binding{
		m.Quit,
		m.Open,
		m.Delete,
		m.GetContent,
		m.EditTags,
	}
}
