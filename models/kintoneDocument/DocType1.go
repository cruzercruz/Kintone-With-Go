package DocType1

type DataObj struct {
	Records []Record `json:"records"`
}

type Record struct {
	Approved_datetime FieldType `json:"Approved_datetime"`
	DEPT              FieldType `json:"DEPT"`
	Record_number     FieldType `json:"Record_number"`
	Approved_by       FieldType `json:"Approved_by"`
	IsSync            FieldType `json:"IsSync"`
}

type FieldType struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}
