package models

type Company struct {
	ID       int64
	Name     string
	Archived *bool
}

func (c *Company) IsArchived() bool {
	if c.Archived == nil {
		return false
	}
	return *c.Archived
}

type CompanyWithShares struct {
	Company
	Shares map[int]int64
}
