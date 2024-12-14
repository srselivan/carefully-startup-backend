package models

type AdditionalInfoType int8

const (
	AdditionalInfoTypeCompanyInfo AdditionalInfoType = iota + 1
	AdditionalInfoTypeAnalytics
)

type AdditionalInfo struct {
	ID          int64
	Name        string
	Description string
	Type        AdditionalInfoType
	Cost        int64
	CompanyID   *int64
	Round       int
}
