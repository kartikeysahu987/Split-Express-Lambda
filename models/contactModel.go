package models

type Contact struct {
	Name      *string `json:"name"`
	ContactNo *string `json:"contactno"`
}
type ContactInfo struct {
	Name      			*string 					`json:"name"`
	ContactNo 			*string 					`json:"contactno"`
	Uid       			*string 					`json:"uid"`
	UserName  			*string 					`json:"username"`
}
type GetContact struct {
	Contacts			[]Contact 				`json:"contacts"`
}

type PostContact struct {
	ContactsInfo 		[]ContactInfo 			`json:"contactsinfo"`
}
